package handlers

import (
	"mikhailjbs/user-auth-service/internal/domain/user"
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

func (h *userHandler) UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")
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
