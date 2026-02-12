package handler

import (
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type ShiftHandler struct {
	repo       repository.ShiftRepository
	jadwalRepo repository.JadwalRepository
}

func NewShiftHandler(repo repository.ShiftRepository, jadwalRepo repository.JadwalRepository) *ShiftHandler {
	return &ShiftHandler{repo: repo, jadwalRepo: jadwalRepo}
}

func (h *ShiftHandler) GetAll(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64))
	shifts, err := h.repo.GetAll(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data shift"})
	}
	return c.JSON(fiber.Map{"data": shifts})
}

func (h *ShiftHandler) Create(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64))
	var shift model.Shift
	if err := c.BodyParser(&shift); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	shift.OrganisasiID = orgID // Set Org ID

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

	// Validasi: Cek apakah shift sedang digunakan di jadwal
	count, _ := h.jadwalRepo.CountByShiftID(uint(id))
	if count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Shift tidak bisa dihapus karena sedang digunakan dalam jadwal"})
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghapus shift"})
	}
	return c.JSON(fiber.Map{"message": "Shift berhasil dihapus"})
}
