package handler

import (
	"my-flutter-backend/internal/model"
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
	orgID := uint(c.Locals("organisasi_id").(float64))
	today := time.Now().Format("2006-01-02")

	// 1. Ambil Semua Pegawai & Filter by Organisasi
	allASNs, err := h.asnRepo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data pegawai"})
	}

	var totalPegawai int64
	var filteredASNs []model.ASN
	for _, a := range allASNs {
		if a.OrganisasiID == orgID {
			filteredASNs = append(filteredASNs, a)
			totalPegawai++
		}
	}

	// 2. Hitung Statistik Kehadiran Manual (Looping per Pegawai)
	var hadir, terlambat, izin, cuti int64

	for _, asn := range filteredASNs {
		k, err := h.kehadiranRepo.GetTodayAttendance(asn.ID)
		if err == nil && k != nil {
			switch k.StatusMasuk {
			case "HADIR":
				hadir++
			case "TERLAMBAT":
				terlambat++
			case "IZIN":
				izin++
			case "CUTI":
				cuti++
			}
		}
	}

	// Hitung Alpha (Belum Absen)
	// Alpha = Total Pegawai - (Hadir + Terlambat + Izin + Cuti)
	sudahAbsen := hadir + terlambat + izin + cuti
	alpha := totalPegawai - sudahAbsen
	if alpha < 0 {
		alpha = 0
	}

	return c.JSON(fiber.Map{
		"total_pegawai": totalPegawai,
		"hadir":         hadir + terlambat, // Terlambat tetap dihitung hadir secara fisik
		"terlambat":     terlambat,
		"izin_cuti":     izin + cuti,
		"alpha":         alpha,
		"tanggal":       today,
	})
}
