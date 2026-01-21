package handler

import (
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type BannerHandler struct {
	repo repository.BannerRepository
}

func NewBannerHandler(repo repository.BannerRepository) *BannerHandler {
	return &BannerHandler{repo: repo}
}

func (h *BannerHandler) GetAll(c *fiber.Ctx) error {
	banners, err := h.repo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil banner"})
	}
	return c.JSON(fiber.Map{"data": banners})
}

func (h *BannerHandler) Create(c *fiber.Ctx) error {
	var banner model.Banner
	if err := c.BodyParser(&banner); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	banner.IsActive = true
	if err := h.repo.Create(&banner); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat banner"})
	}
	return c.JSON(fiber.Map{"message": "Banner berhasil dibuat", "data": banner})
}

func (h *BannerHandler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.repo.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghapus banner"})
	}
	return c.JSON(fiber.Map{"message": "Banner berhasil dihapus"})
}
