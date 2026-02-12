package handler

import (
	"fmt"
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type BannerHandler struct {
	repo repository.BannerRepository
}

func NewBannerHandler(repo repository.BannerRepository) *BannerHandler {
	return &BannerHandler{repo: repo}
}

func (h *BannerHandler) GetAll(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64)) // Ambil OrgID dari Token
	banners, err := h.repo.GetAllActive(orgID)         // Mobile hanya lihat yang aktif & sesuai org
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil banner"})
	}

	// Modifikasi Foto agar menjadi Full URL
	baseURL := c.BaseURL() // Otomatis mendeteksi http://localhost:3000 atau IP
	for i := range banners {
		banners[i].Foto = baseURL + "/" + banners[i].Foto
	}

	return c.JSON(fiber.Map{"data": banners})
}

func (h *BannerHandler) GetAllAdmin(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64))
	banners, err := h.repo.GetAll(orgID) // Admin lihat semua di organisasinya
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil banner"})
	}

	baseURL := c.BaseURL()
	for i := range banners {
		banners[i].Foto = baseURL + "/" + banners[i].Foto
	}

	return c.JSON(fiber.Map{"data": banners})
}

func (h *BannerHandler) Create(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64))
	// Ambil Title dari Form
	title := c.FormValue("title")
	if title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Judul banner wajib diisi"})
	}

	// Handle File Upload
	file, err := c.FormFile("foto")
	pathFile := ""
	if err == nil {
		uploadDir := "./uploads/banners"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.MkdirAll(uploadDir, 0755)
		}
		filename := fmt.Sprintf("banner_%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
		pathFile = fmt.Sprintf("uploads/banners/%s", filename)
		if err := c.SaveFile(file, pathFile); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan file gambar"})
		}
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Foto banner wajib diupload"})
	}

	banner := model.Banner{
		OrganisasiID: orgID,
		Title:        title,
		Foto:         pathFile,
		IsActive:     true,
	}
	if err := h.repo.Create(&banner); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat banner"})
	}
	return c.JSON(fiber.Map{"message": "Banner berhasil dibuat", "data": banner})
}

func (h *BannerHandler) Delete(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.repo.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menonaktifkan banner"})
	}
	return c.JSON(fiber.Map{"message": "Banner berhasil dinonaktifkan"})
}

func (h *BannerHandler) ToggleStatus(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.repo.ToggleStatus(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengubah status banner"})
	}
	return c.JSON(fiber.Map{"message": "Status banner berhasil diubah"})
}
