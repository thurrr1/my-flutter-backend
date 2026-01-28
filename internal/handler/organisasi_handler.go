package handler

import (
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type OrganisasiHandler struct {
	repo repository.OrganisasiRepository
}

func NewOrganisasiHandler(repo repository.OrganisasiRepository) *OrganisasiHandler {
	return &OrganisasiHandler{repo: repo}
}

func (h *OrganisasiHandler) GetInfo(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64))
	org, err := h.repo.GetByID(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data organisasi"})
	}
	return c.JSON(fiber.Map{"data": org})
}

type UpdateOrganisasiRequest struct {
	NamaOrganisasi string `json:"nama_organisasi"`
	EmailAdmin     string `json:"email_admin"`
}

func (h *OrganisasiHandler) UpdateOrganisasi(c *fiber.Ctx) error {
	// Ambil ID Organisasi dari token Admin yang login
	orgID := uint(c.Locals("organisasi_id").(float64))

	var req UpdateOrganisasiRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	org, err := h.repo.GetByID(orgID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Organisasi tidak ditemukan"})
	}

	if req.NamaOrganisasi != "" {
		org.NamaOrganisasi = req.NamaOrganisasi
	}
	org.EmailAdmin = req.EmailAdmin // Email boleh kosong/diupdate

	if err := h.repo.Update(org); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update organisasi"})
	}

	return c.JSON(fiber.Map{"message": "Informasi organisasi berhasil diperbarui", "data": org})
}

type UpdateLokasiRequest struct {
	NamaLokasi  string  `json:"nama_lokasi"`
	Alamat      string  `json:"alamat"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	RadiusMeter float64 `json:"radius_meter"`
}

func (h *OrganisasiHandler) UpdateLokasi(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var req UpdateLokasiRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	lokasi, err := h.repo.GetLokasiByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Lokasi tidak ditemukan"})
	}

	lokasi.NamaLokasi = req.NamaLokasi
	lokasi.Alamat = req.Alamat
	lokasi.Latitude = req.Latitude
	lokasi.Longitude = req.Longitude
	lokasi.RadiusMeter = req.RadiusMeter

	if err := h.repo.UpdateLokasi(lokasi); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update lokasi"})
	}

	return c.JSON(fiber.Map{"message": "Lokasi berhasil diupdate", "data": lokasi})
}

func (h *OrganisasiHandler) AddLokasi(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64))
	var req UpdateLokasiRequest // Gunakan struct yang sama
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	lokasi := model.Lokasi{
		OrganisasiID: orgID,
		NamaLokasi:   req.NamaLokasi,
		Alamat:       req.Alamat,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		RadiusMeter:  req.RadiusMeter,
	}

	if err := h.repo.CreateLokasi(&lokasi); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menambah lokasi"})
	}

	return c.JSON(fiber.Map{"message": "Lokasi berhasil ditambahkan", "data": lokasi})
}

func (h *OrganisasiHandler) DeleteLokasi(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	if err := h.repo.DeleteLokasi(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghapus lokasi"})
	}

	return c.JSON(fiber.Map{"message": "Lokasi berhasil dihapus"})
}
