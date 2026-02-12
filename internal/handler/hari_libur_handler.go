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
	orgID := uint(c.Locals("organisasi_id").(float64))
	data, err := h.repo.GetAll(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data"})
	}
	return c.JSON(fiber.Map{"data": data})
}

func (h *HariLiburHandler) Create(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64))
	var req model.HariLibur
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	req.OrganisasiID = orgID // Set Org ID

	if err := h.repo.Create(&req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan data"})
	}

	return c.JSON(fiber.Map{"message": "Hari libur berhasil ditambahkan", "data": req})
}

func (h *HariLiburHandler) Update(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var req model.HariLibur
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	libur, err := h.repo.GetByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Data tidak ditemukan"})
	}

	libur.Tanggal = req.Tanggal
	libur.Keterangan = req.Keterangan

	if err := h.repo.Update(libur); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update data"})
	}
	return c.JSON(fiber.Map{"message": "Data berhasil diupdate", "data": libur})
}

func (h *HariLiburHandler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.repo.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghapus data"})
	}
	return c.JSON(fiber.Map{"message": "Data berhasil dihapus"})
}
