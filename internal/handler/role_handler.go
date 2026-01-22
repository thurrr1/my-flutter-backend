package handler

import (
	"my-flutter-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
)

type RoleHandler struct {
	repo repository.RoleRepository
}

func NewRoleHandler(repo repository.RoleRepository) *RoleHandler {
	return &RoleHandler{repo: repo}
}

func (h *RoleHandler) GetAll(c *fiber.Ctx) error {
	roles, err := h.repo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data role"})
	}
	return c.JSON(fiber.Map{"data": roles})
}
