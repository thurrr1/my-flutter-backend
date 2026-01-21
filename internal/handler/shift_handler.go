package handler

import (
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ShiftHandler struct {
	repo repository.ShiftRepository
}

func NewShiftHandler(repo repository.ShiftRepository) *ShiftHandler {
	return &ShiftHandler{repo: repo}
}

func (h *ShiftHandler) GetAll(c *fiber.Ctx) error {
	shifts, err := h.repo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data shift"})
	}
	return c.JSON(fiber.Map{"data": shifts})
}

func (h *ShiftHandler) Create(c *fiber.Ctx) error {
	var shift model.Shift
	if err := c.BodyParser(&shift); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	if err := h.repo.Create(&shift); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat shift"})
	}
	return c.JSON(fiber.Map{"message": "Shift berhasil dibuat", "data": shift})
}

func (h *ShiftHandler) Update(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var req model.Shift
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	shift, err := h.repo.GetByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Shift tidak ditemukan"})
	}

	shift.NamaShift = req.NamaShift
	shift.JamMasuk = req.JamMasuk
	shift.JamPulang = req.JamPulang

	if err := h.repo.Update(shift); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update shift"})
	}
	return c.JSON(fiber.Map{"message": "Shift berhasil diupdate", "data": shift})
}

func (h *ShiftHandler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.repo.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghapus shift"})
	}
	return c.JSON(fiber.Map{"message": "Shift berhasil dihapus"})
}
