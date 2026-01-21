package handler

import (
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type HariLiburHandler struct {
	repo repository.HariLiburRepository
}

func NewHariLiburHandler(repo repository.HariLiburRepository) *HariLiburHandler {
	return &HariLiburHandler{repo: repo}
}

func (h *HariLiburHandler) GetAll(c *fiber.Ctx) error {
	data, err := h.repo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data"})
	}
	return c.JSON(fiber.Map{"data": data})
}

func (h *HariLiburHandler) Create(c *fiber.Ctx) error {
	var req model.HariLibur
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	if err := h.repo.Create(&req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan data"})
	}

	return c.JSON(fiber.Map{"message": "Hari libur berhasil ditambahkan", "data": req})
}

func (h *HariLiburHandler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.repo.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghapus data"})
	}
	return c.JSON(fiber.Map{"message": "Data berhasil dihapus"})
}
