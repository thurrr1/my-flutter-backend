package handler

import (
	"my-flutter-backend/internal/repository"
	"time"

	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct {
	asnRepo       repository.ASNRepository
	kehadiranRepo repository.KehadiranRepository
}

func NewDashboardHandler(asnRepo repository.ASNRepository, kRepo repository.KehadiranRepository) *DashboardHandler {
	return &DashboardHandler{asnRepo: asnRepo, kehadiranRepo: kRepo}
}

func (h *DashboardHandler) GetStats(c *fiber.Ctx) error {
	today := time.Now().Format("2006-01-02")

	// 1. Total Pegawai
	totalASN, _ := h.asnRepo.Count()

	// 2. Statistik Kehadiran Hari Ini
	hadir, _ := h.kehadiranRepo.CountByStatus(today, "HADIR")
	terlambat, _ := h.kehadiranRepo.CountByStatus(today, "TERLAMBAT")
	izin, _ := h.kehadiranRepo.CountByStatus(today, "IZIN")
	cuti, _ := h.kehadiranRepo.CountByStatus(today, "CUTI")

	// Hitung Alpha (Belum Absen)
	// Alpha = Total Pegawai - (Hadir + Terlambat + Izin + Cuti)
	sudahAbsen := hadir + terlambat + izin + cuti
	alpha := totalASN - sudahAbsen
	if alpha < 0 {
		alpha = 0
	}

	return c.JSON(fiber.Map{
		"total_pegawai": totalASN,
		"hadir":         hadir + terlambat, // Terlambat tetap dihitung hadir secara fisik
		"terlambat":     terlambat,
		"izin_cuti":     izin + cuti,
		"alpha":         alpha,
		"tanggal":       today,
	})
}
