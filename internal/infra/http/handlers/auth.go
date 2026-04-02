package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"mikhailjbs/user-auth-service/internal/domain/auth"
	"mikhailjbs/user-auth-service/internal/domain/session"
	"mikhailjbs/user-auth-service/internal/domain/user"
	"mikhailjbs/user-auth-service/internal/infra/security"
	authusecase "mikhailjbs/user-auth-service/internal/usecase/auth"
)

const (
	accessTokenCookieName  = "access_token"
	refreshTokenCookieName = "refresh_token"
)

// AuthHandler exposes HTTP endpoints for authentication workflows.
type AuthHandler interface {
	Register(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	Refresh(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
	Me(c *fiber.Ctx) error
}

type authHandler struct {
	registerUC     authusecase.RegisterUseCase
	loginUC        authusecase.LoginUseCase
	meUC           authusecase.GetMeUseCase
	sessionService session.Service
	tokenManager   *security.TokenManager
	cookieDomain   string
}

func NewAuthHandler(
	registerUC authusecase.RegisterUseCase,
	loginUC authusecase.LoginUseCase,
	meUC authusecase.GetMeUseCase,
	sessionService session.Service,
	tokenManager *security.TokenManager,
	cookieDomain string,
) AuthHandler {
	return &authHandler{
		registerUC:     registerUC,
		loginUC:        loginUC,
		meUC:           meUC,
		sessionService: sessionService,
		tokenManager:   tokenManager,
		cookieDomain:   cookieDomain,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user to the system.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.RegisterRequest true "Register User Request"
// @Success 201 {object} handlers.SuccessResponse{data=user.User}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 409 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *authHandler) Register(c *fiber.Ctx) error {
	var req auth.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return SendError(c, fiber.StatusBadRequest, "invalid request body")
	}

	createdUser, err := h.registerUC.Execute(c.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrEmailTaken):
			return SendError(c, fiber.StatusConflict, err.Error())
		default:
			return SendError(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return SendSuccess(c, fiber.StatusCreated, "user registered successfully", sanitizeUser(createdUser))
}

// Login godoc
// @Summary Login a user
// @Description Authenticates a user and issues JWT tokens in cookies.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.LoginRequest true "Login Request"
// @Success 200 {object} handlers.SuccessResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *authHandler) Login(c *fiber.Ctx) error {
	var req auth.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return SendError(c, fiber.StatusBadRequest, "invalid request body")
	}

	if req.IPAddress == "" {
		req.IPAddress = c.IP()
	}
	if req.UserAgent == "" {
		req.UserAgent = c.Get("User-Agent")
	}

	authenticatedUser, err := h.loginUC.Execute(c.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidCredentials):
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		case errors.Is(err, auth.ErrUserNotFound):
			return SendError(c, fiber.StatusNotFound, err.Error())
		default:
			return SendError(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	pair, err := h.tokenManager.GenerateTokenPair(
		authenticatedUser.ID,
		authenticatedUser.Email,
		authenticatedUser.Username,
		[]string{string(authenticatedUser.Role)},
	)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "failed to generate tokens")
	}

	if err := h.persistSession(pair.SID, authenticatedUser.ID, req.IPAddress, req.UserAgent, pair.RefreshToken, pair.RefreshExp); err != nil {
		return SendError(c, fiber.StatusInternalServerError, "failed to persist session")
	}

	h.setAuthCookies(c, pair)
	data := map[string]interface{}{
		"user":                     sanitizeUser(authenticatedUser),
		"session_id":               pair.SID,
		"access_token_expires_at":  pair.AccessExp,
		"refresh_token_expires_at": pair.RefreshExp,
	}

	return SendSuccess(c, fiber.StatusOK, "user logged in successfully", data)
}

// Refresh godoc
// @Summary Refresh access token
// @Description Issues a new access and refresh token pair using a valid refresh token cookie.
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} handlers.SuccessResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *authHandler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies(refreshTokenCookieName)
	if refreshToken == "" {
		return SendError(c, fiber.StatusUnauthorized, "missing refresh token")
	}

	payload, err := h.tokenManager.ParseRefreshToken(refreshToken)
	if err != nil {
		return SendError(c, fiber.StatusUnauthorized, "invalid refresh token")
	}

	sess, err := h.sessionService.GetSessionByID(payload.SID)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "failed to lookup session")
	}
	if sess == nil || !sess.Valid || time.Now().After(sess.ExpiresAt) {
		return SendError(c, fiber.StatusUnauthorized, "session expired or revoked")
	}

	if !security.CompareTokenHash(sess.RefreshTokenHash, refreshToken) {
		_ = h.sessionService.DeleteSession(payload.SID)
		h.clearAuthCookies(c)
		return SendError(c, fiber.StatusUnauthorized, "refresh token reuse detected")
	}

	userRecord, err := h.meUC.Execute(c.Context(), sess.ID)
	if err != nil {
		return SendError(c, fiber.StatusUnauthorized, "linked user not found")
	}

	pair, err := h.tokenManager.GenerateTokenPair(
		userRecord.ID,
		userRecord.Email,
		userRecord.Username,
		[]string{string(userRecord.Role)},
	)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "failed to rotate tokens")
	}

	if err := h.rotateSession(sess.ID, pair.RefreshToken, pair.RefreshExp, c.IP(), c.Get("User-Agent")); err != nil {
		return SendError(c, fiber.StatusInternalServerError, "failed to rotate session")
	}

	h.setAuthCookies(c, pair)
	data := map[string]interface{}{
		"user":                     sanitizeUser(userRecord),
		"session_id":               pair.SID,
		"access_token_expires_at":  pair.AccessExp,
		"refresh_token_expires_at": pair.RefreshExp,
	}

	return SendSuccess(c, fiber.StatusOK, "tokens refreshed", data)
}

// Logout godoc
// @Summary Logout a user
// @Description Revokes the user session and clears authentication cookies.
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handlers.SuccessResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/v1/auth/logout [post]
func (h *authHandler) Logout(c *fiber.Ctx) error {
	sessionID := h.sessionIDFromRequest(c)
	if sessionID == "" {
		return SendError(c, fiber.StatusUnauthorized, "unable to determine session")
	}

	if err := h.sessionService.DeleteSession(sessionID); err != nil {
		return SendError(c, fiber.StatusInternalServerError, "failed to revoke session")
	}

	h.clearAuthCookies(c)
	return SendSuccess(c, fiber.StatusOK, "logged out successfully", nil)
}

// Me godoc
// @Summary Get current user profile
// @Description Fetches the profile of the currently authenticated user.
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} handlers.SuccessResponse{data=user.User}
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/v1/auth/me [get]
func (h *authHandler) Me(c *fiber.Ctx) error {
	token := extractBearerToken(c.Get("Authorization"))
	if token == "" {
		token = c.Cookies(accessTokenCookieName)
	}
	if token == "" {
		token = c.Query("token")
	}
	if token == "" {
		return SendError(c, fiber.StatusUnauthorized, "missing authorization token")
	}

	u, err := h.meUC.Execute(c.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrSessionNotFound):
			return SendError(c, fiber.StatusUnauthorized, err.Error())
		case errors.Is(err, auth.ErrUserNotFound), errors.Is(err, user.ErrNotFound):
			return SendError(c, fiber.StatusNotFound, err.Error())
		default:
			return SendError(c, fiber.StatusInternalServerError, err.Error())
		}
	}

	return SendSuccess(c, fiber.StatusOK, "authenticated user retrieved", sanitizeUser(u))
}

func (h *authHandler) persistSession(sessionID, userID, ip, userAgent, refreshToken string, expiresAt time.Time) error {
	refreshHash := security.HashToken(refreshToken)
	now := time.Now().UTC()
	sess := &session.Session{
		ID:               sessionID,
		UserID:           userID,
		IPAddress:        ip,
		UserAgent:        userAgent,
		Valid:            true,
		ExpiresAt:        expiresAt,
		RefreshTokenHash: refreshHash,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	return h.sessionService.CreateSession(sess)
}

func (h *authHandler) rotateSession(sessionID, refreshToken string, expiresAt time.Time, ip, userAgent string) error {
	refreshHash := security.HashToken(refreshToken)
	updated := &session.Session{
		RefreshTokenHash: refreshHash,
		ExpiresAt:        expiresAt,
		IPAddress:        ip,
		UserAgent:        userAgent,
		UpdatedAt:        time.Now().UTC(),
	}
	_, err := h.sessionService.UpdateSession(sessionID, updated)
	return err
}

func (h *authHandler) setAuthCookies(c *fiber.Ctx, pair *security.TokenPair) {
	accessMaxAge := int(time.Until(pair.AccessExp).Seconds())
	refreshMaxAge := int(time.Until(pair.RefreshExp).Seconds())
	if accessMaxAge <= 0 {
		accessMaxAge = int(h.tokenManager.AccessTTL().Seconds())
	}
	if refreshMaxAge <= 0 {
		refreshMaxAge = int(h.tokenManager.RefreshTTL().Seconds())
	}

	cookieBase := func(name, value string, maxAge int) *fiber.Cookie {
		// Determine if we're in production (using a real domain, not localhost)
		isProduction := h.cookieDomain != "" && h.cookieDomain != "localhost"

		cookie := &fiber.Cookie{
			Name:     name,
			Value:    value,
			Path:     "/",
			HTTPOnly: true,
			Secure:   isProduction, // Secure=true for production (HTTPS required)
			SameSite: fiber.CookieSameSiteLaxMode,
			MaxAge:   maxAge,
		}

		// For cross-subdomain cookies in production
		if isProduction {
			cookie.Domain = h.cookieDomain
			// Use SameSite=None for cross-subdomain, but requires Secure=true
			cookie.SameSite = fiber.CookieSameSiteNoneMode
		}

		return cookie
	}

	c.Cookie(cookieBase(accessTokenCookieName, pair.AccessToken, accessMaxAge))
	c.Cookie(cookieBase(refreshTokenCookieName, pair.RefreshToken, refreshMaxAge))
}

func (h *authHandler) clearAuthCookies(c *fiber.Ctx) {
	clear := func(name string) {
		// Determine if we're in production
		isProduction := h.cookieDomain != "" && h.cookieDomain != "localhost"

		cookie := &fiber.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			HTTPOnly: true,
			Secure:   isProduction,
			SameSite: fiber.CookieSameSiteLaxMode,
			Expires:  time.Unix(0, 0),
		}

		if isProduction {
			cookie.Domain = h.cookieDomain
			cookie.SameSite = fiber.CookieSameSiteNoneMode
		}

		c.Cookie(cookie)
	}
	clear(accessTokenCookieName)
	clear(refreshTokenCookieName)
}

func (h *authHandler) sessionIDFromRequest(c *fiber.Ctx) string {
	if token := extractBearerToken(c.Get("Authorization")); token != "" {
		if payload, err := h.tokenManager.ParseAccessToken(token); err == nil {
			return payload.SID
		}
	}
	if token := c.Cookies(accessTokenCookieName); token != "" {
		if payload, err := h.tokenManager.ParseAccessToken(token); err == nil {
			return payload.SID
		}
	}
	if token := c.Cookies(refreshTokenCookieName); token != "" {
		if payload, err := h.tokenManager.ParseRefreshToken(token); err == nil {
			return payload.SID
		}
	}
	return c.Query("token")
}

func extractBearerToken(header string) string {
	if header == "" {
		return ""
	}
	header = strings.TrimSpace(header)
	const prefix = "Bearer "
	if strings.HasPrefix(strings.ToLower(header), strings.ToLower(prefix)) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return header
}

func sanitizeUser(u *user.User) *user.User {
	if u == nil {
		return nil
	}

	clone := *u
	clone.PasswordHash = ""
	return &clone
}
