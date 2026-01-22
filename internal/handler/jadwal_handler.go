package handler

import (
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type JadwalHandler struct {
	repo          repository.JadwalRepository
	hariLiburRepo repository.HariLiburRepository // Tambah ini
}

func NewJadwalHandler(repo repository.JadwalRepository, hlRepo repository.HariLiburRepository) *JadwalHandler {
	return &JadwalHandler{repo: repo, hariLiburRepo: hlRepo}
}

type CreateJadwalRequest struct {
	ASNID   uint   `json:"asn_id"`
	ShiftID uint   `json:"shift_id"`
	Tanggal string `json:"tanggal"` // Format: YYYY-MM-DD
}

func (h *JadwalHandler) CreateJadwal(c *fiber.Ctx) error {
	var req CreateJadwalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	jadwal := model.Jadwal{
		ASNID:   req.ASNID,
		ShiftID: req.ShiftID,
		Tanggal: req.Tanggal,
	}

	if err := h.repo.Create(&jadwal); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat jadwal"})
	}

	return c.JSON(fiber.Map{
		"message": "Jadwal berhasil dibuat",
		"data":    jadwal,
	})
}

type GenerateJadwalRequest struct {
	ASNIDs  []uint `json:"asn_ids"` // UBAH: Array of ID (Checkbox)
	ShiftID uint   `json:"shift_id"`
	Bulan   int    `json:"bulan"`
	Tahun   int    `json:"tahun"`
	Days    []int  `json:"days"` // 0=Minggu, 1=Senin, ..., 6=Sabtu
}

func (h *JadwalHandler) GenerateJadwalBulanan(c *fiber.Ctx) error {
	var req GenerateJadwalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	var listJadwal []model.Jadwal

	// Tentukan tanggal awal bulan
	startDate := time.Date(req.Tahun, time.Month(req.Bulan), 1, 0, 0, 0, 0, time.Local)
	// Tentukan tanggal akhir bulan (tanggal 0 bulan berikutnya = tanggal terakhir bulan ini)
	endDate := startDate.AddDate(0, 1, -1)

	// Loop untuk setiap Pegawai yang dipilih
	for _, asnID := range req.ASNIDs {
		// Loop dari tanggal 1 sampai akhir bulan
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			// 1. Cek apakah hari ini (d.Weekday) ada di daftar hari yang dipilih user
			isSelectedDay := false
			for _, day := range req.Days {
				if int(d.Weekday()) == day {
					isSelectedDay = true
					break
				}
			}
			if !isSelectedDay {
				continue // Skip jika hari tidak dicentang
			}

			// Cek apakah tanggal ini adalah Hari Libur Nasional
			isHoliday, _ := h.hariLiburRepo.IsHoliday(d.Format("2006-01-02"))
			if isHoliday {
				continue // Skip jika tanggal merah
			}

			jadwal := model.Jadwal{
				ASNID:   asnID,
				ShiftID: req.ShiftID,
				Tanggal: d.Format("2006-01-02"),
			}
			listJadwal = append(listJadwal, jadwal)
		}
	}

	if len(listJadwal) > 0 {
		if err := h.repo.CreateMany(listJadwal); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal generate jadwal"})
		}
	}

	return c.JSON(fiber.Map{
		"message":    "Berhasil generate jadwal bulanan",
		"total_hari": len(listJadwal),
	})
}

type GenerateJadwalHarianRequest struct {
	ASNIDs  []uint `json:"asn_ids"`
	ShiftID uint   `json:"shift_id"`
	Tanggal string `json:"tanggal"` // YYYY-MM-DD
}

func (h *JadwalHandler) GenerateJadwalHarian(c *fiber.Ctx) error {
	var req GenerateJadwalHarianRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	var listJadwal []model.Jadwal
	for _, asnID := range req.ASNIDs {
		jadwal := model.Jadwal{
			ASNID:   asnID,
			ShiftID: req.ShiftID,
			Tanggal: req.Tanggal,
		}
		listJadwal = append(listJadwal, jadwal)
	}

	if len(listJadwal) > 0 {
		if err := h.repo.CreateMany(listJadwal); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat jadwal harian"})
		}
	}
	return c.JSON(fiber.Map{"message": "Berhasil membuat jadwal harian"})
}

// GET /api/admin/jadwal?tanggal=2024-10-25
func (h *JadwalHandler) GetJadwalHarian(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64))
	tanggal := c.Query("tanggal")

	if tanggal == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Parameter tanggal wajib diisi"})
	}

	jadwals, err := h.repo.GetByDate(tanggal, orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil jadwal"})
	}

	return c.JSON(fiber.Map{"data": jadwals})
}

func (h *JadwalHandler) GetJadwalDetail(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	jadwal, err := h.repo.GetByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Jadwal tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"data": jadwal})
}

type UpdateJadwalRequest struct {
	ShiftID uint `json:"shift_id"`
}

func (h *JadwalHandler) UpdateJadwal(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	var req UpdateJadwalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	jadwal, err := h.repo.GetByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Jadwal tidak ditemukan"})
	}

	jadwal.ShiftID = req.ShiftID
	h.repo.Update(jadwal)
	return c.JSON(fiber.Map{"message": "Jadwal berhasil diupdate"})
}

func (h *JadwalHandler) DeleteJadwal(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))
	if err := h.repo.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghapus jadwal"})
	}
	return c.JSON(fiber.Map{"message": "Jadwal berhasil dihapus"})
}

func (h *JadwalHandler) DeleteJadwalByDate(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64))
	tanggal := c.Query("tanggal")
	if tanggal == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Tanggal wajib diisi"})
	}

	h.repo.DeleteByDate(tanggal, orgID)
	return c.JSON(fiber.Map{"message": "Semua jadwal pada tanggal tersebut berhasil dihapus"})
}
