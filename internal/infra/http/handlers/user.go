package handlers

import (
	"mikhailjbs/user-auth-service/internal/domain/user"
	"mikhailjbs/user-auth-service/internal/infra/middleware"
	usecase "mikhailjbs/user-auth-service/internal/usecase/user"

	"github.com/gofiber/fiber/v2"
)

type UserHandler interface {
	CreateUser(c *fiber.Ctx) error
	GetUsers(c *fiber.Ctx) error
	GetUser(c *fiber.Ctx) error
	UpdateUser(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
}

type userHandler struct {
	createUserUC usecase.CreateUserUseCase
	getUsersUC   usecase.GetUsersUseCase
	getUserUC    usecase.GetUserUseCase
	updateUserUC usecase.UpdateUserUseCase
	deleteUserUC usecase.DeleteUserUseCase
}

func NewUserHandler(
	createUserUC usecase.CreateUserUseCase,
	getUsersUC usecase.GetUsersUseCase,
	getUserUC usecase.GetUserUseCase,
	updateUserUC usecase.UpdateUserUseCase,
	deleteUserUC usecase.DeleteUserUseCase,
) UserHandler {
	return &userHandler{
		createUserUC: createUserUC,
		getUsersUC:   getUsersUC,
		getUserUC:    getUserUC,
		updateUserUC: updateUserUC,
		deleteUserUC: deleteUserUC,
	}
}

// CreateUser godoc
// @Summary Create a new user
// @Description Creates a new user in the system. Required Admin role.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body user.CreateUserRequest true "Create User Request"
// @Success 201 {object} handlers.SuccessResponse{data=user.User}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 409 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/v1/users [post]
func (h *userHandler) CreateUser(c *fiber.Ctx) error {
	var req user.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	createdUser, err := h.createUserUC.Execute(c.Context(), &req)
	if err != nil {
		if err == user.ErrEmailTaken {
			return SendError(c, fiber.StatusConflict, err.Error())
		}
		return SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return SendSuccess(c, fiber.StatusCreated, "User created successfully", createdUser)
}

// GetUsers godoc
// @Summary List users
// @Description Fetch all users with optional filters. Required Admin role.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param email query string false "Filter by exact email"
// @Param role query string false "Filter by exact role (user, admin)"
// @Param search query string false "Search by username or fullname"
// @Success 200 {object} handlers.SuccessResponse{data=[]user.User}
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/v1/users [get]
func (h *userHandler) GetUsers(c *fiber.Ctx) error {
	params := &user.UserQueryParams{}

	if email := c.Query("email"); email != "" {
		params.Email = &email
	}
	if role := c.Query("role"); role != "" {
		params.Role = &role
	}
	if search := c.Query("search"); search != "" {
		params.Search = &search
	}

	users, err := h.getUsersUC.Execute(c.Context(), params)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Users retrieved successfully", users)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Fetches a single user by their unique identifier. Required Admin role.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} handlers.SuccessResponse{data=user.User}
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/v1/users/{id} [get]
func (h *userHandler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	u, err := h.getUserUC.Execute(c.Context(), id)
	if err != nil {
		if err == user.ErrNotFound {
			return SendError(c, fiber.StatusNotFound, err.Error())
		}
		return SendError(c, fiber.StatusInternalServerError, err.Error())
	}
	return SendSuccess(c, fiber.StatusOK, "User retrieved successfully", u)
}

// UpdateUser godoc
// @Summary Update user
// @Description Updates user profile information. Required Admin role.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body user.UpdateUserRequest true "Update User Request"
// @Success 200 {object} handlers.SuccessResponse{data=user.User}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/v1/users/{id} [put]
func (h *userHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")

	// Get claims from context
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, "Missing auth claims")
	}

	// Check if user is admin OR updating their own profile
	isAdmin := false
	for _, role := range claims.Roles {
		if role == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin && claims.UserID != id {
		return SendError(c, fiber.StatusForbidden, "You can only update your own profile")
	}

	var req user.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return SendError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	updatedUser, err := h.updateUserUC.Execute(c.Context(), id, &req)
	if err != nil {
		if err == user.ErrNotFound {
			return SendError(c, fiber.StatusNotFound, err.Error())
		}
		return SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "User updated successfully", updatedUser)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Deletes a user from the system. Required Admin role.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} handlers.SuccessResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/v1/users/{id} [delete]
func (h *userHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.deleteUserUC.Execute(c.Context(), id); err != nil {
		if err == user.ErrNotFound {
			return SendError(c, fiber.StatusNotFound, err.Error())
		}
		return SendError(c, fiber.StatusInternalServerError, err.Error())
	}
	return SendSuccess(c, fiber.StatusOK, "User deleted successfully", nil)
}
