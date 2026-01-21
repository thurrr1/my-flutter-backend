package handler

import (
	"fmt"
	"math"
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"time"

	"github.com/gofiber/fiber/v2"
)

type KehadiranHandler struct {
	repo       repository.KehadiranRepository
	asnRepo    repository.ASNRepository
	jadwalRepo repository.JadwalRepository // Tambah ini
}

func NewKehadiranHandler(repo repository.KehadiranRepository, asnRepo repository.ASNRepository, jadwalRepo repository.JadwalRepository) *KehadiranHandler {
	return &KehadiranHandler{repo: repo, asnRepo: asnRepo, jadwalRepo: jadwalRepo}
}

type CheckInRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (h *KehadiranHandler) CheckIn(c *fiber.Ctx) error {
	// 1. Ambil Data User dari Middleware
	asnID := uint(c.Locals("user_id").(float64))
	orgID := uint(c.Locals("organisasi_id").(float64))

	var req CheckInRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	// 2. Cek Double Check-in
	existing, _ := h.repo.GetTodayAttendance(asnID)
	if existing != nil {
		if existing.StatusMasuk == "IZIN" || existing.StatusMasuk == "CUTI" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Anda sedang dalam status Izin/Cuti hari ini"})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Anda sudah melakukan Check-in hari ini"})
	}

	// 3. Ambil Jadwal Hari Ini (Untuk Cek Shift)
	now := time.Now()
	jadwal, err := h.jadwalRepo.GetByASNAndDate(asnID, now.Format("2006-01-02"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Jadwal kerja hari ini belum ditentukan. Hubungi Admin."})
	}

	// 4. Ambil Lokasi Kantor & Validasi Radius
	lokasiKantor, err := h.asnRepo.GetLokasiByOrganisasiID(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Lokasi kantor tidak ditemukan"})
	}

	jarak := calculateDistance(req.Latitude, req.Longitude, lokasiKantor.Latitude, lokasiKantor.Longitude)

	// Debugging: Tampilkan jarak di console server
	// fmt.Printf("Jarak user ke kantor: %.2f meter (Radius: %d)\n", jarak, lokasiKantor.RadiusMeter)

	statusLokasiMasuk := "VALID"
	if jarak > float64(lokasiKantor.RadiusMeter) {
		statusLokasiMasuk = "INVALID"
	}

	// 5. Tentukan Status (HADIR / TERLAMBAT)
	statusMasuk := "HADIR"

	// Parse Jam Masuk dari Shift (Format "07:30")
	jamMasukShift, _ := time.Parse("15:04", jadwal.Shift.JamMasuk)
	// Gabungkan dengan tanggal hari ini agar bisa dibandingin
	waktuMasukShift := time.Date(now.Year(), now.Month(), now.Day(), jamMasukShift.Hour(), jamMasukShift.Minute(), 0, 0, now.Location())

	// Jika waktu sekarang > waktu shift, maka TERLAMBAT
	if now.After(waktuMasukShift) {
		statusMasuk = "TERLAMBAT"
	}

	kehadiran := model.Kehadiran{
		ASNID:             asnID,
		JadwalID:          jadwal.ID, // Simpan ID Jadwal
		Tanggal:           now.Format("2006-01-02"),
		JamMasukReal:      now.Format("15:04:05"),
		KoordinatMasuk:    fmt.Sprintf("%f,%f", req.Latitude, req.Longitude),
		StatusMasuk:       statusMasuk,
		StatusLokasiMasuk: statusLokasiMasuk,
		Tahun:             now.Format("2006"),
		Bulan:             now.Format("01"),
	}

	if err := h.repo.Create(kehadiran); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan absensi"})
	}

	return c.JSON(fiber.Map{
		"message": "Check-in berhasil",
		"status":  statusMasuk,
		"waktu":   kehadiran.JamMasukReal,
		"jarak":   jarak,
	})
}

func (h *KehadiranHandler) CheckOut(c *fiber.Ctx) error {
	// 1. Ambil Data User
	asnID := uint(c.Locals("user_id").(float64))
	orgID := uint(c.Locals("organisasi_id").(float64))

	var req CheckInRequest // Kita pakai struct request yang sama (butuh koordinat)
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	// 2. Cek Apakah Sudah Check-in Hari Ini
	kehadiran, err := h.repo.GetTodayAttendance(asnID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Anda belum melakukan Check-in hari ini"})
	}

	if kehadiran.StatusMasuk == "IZIN" || kehadiran.StatusMasuk == "CUTI" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Anda sedang Cuti/Izin, tidak perlu Check-out"})
	}

	if kehadiran.JamPulangReal != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Anda sudah melakukan Check-out hari ini"})
	}

	// 3. Ambil Jadwal Hari Ini (Untuk Cek Shift Pulang)
	now := time.Now()
	jadwal, err := h.jadwalRepo.GetByASNAndDate(asnID, now.Format("2006-01-02"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Jadwal kerja hari ini tidak ditemukan."})
	}

	// 3. Validasi Lokasi (Sama seperti Check-in)
	lokasiKantor, err := h.asnRepo.GetLokasiByOrganisasiID(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Lokasi kantor tidak ditemukan"})
	}

	jarak := calculateDistance(req.Latitude, req.Longitude, lokasiKantor.Latitude, lokasiKantor.Longitude)

	// 4. Update Data Pulang
	kehadiran.JamPulangReal = now.Format("15:04:05")
	kehadiran.KoordinatPulang = fmt.Sprintf("%f,%f", req.Latitude, req.Longitude)

	// Tentukan Status Pulang (PULANG / PULANG_CEPAT)
	statusPulang := "PULANG"
	jamPulangShift, _ := time.Parse("15:04", jadwal.Shift.JamPulang)
	// Gabungkan dengan tanggal hari ini
	waktuPulangShift := time.Date(now.Year(), now.Month(), now.Day(), jamPulangShift.Hour(), jamPulangShift.Minute(), 0, 0, now.Location())

	if now.Before(waktuPulangShift) {
		statusPulang = "PULANG_CEPAT"
	}
	kehadiran.StatusPulang = statusPulang

	// Validasi Radius Pulang
	kehadiran.StatusLokasiPulang = "VALID"
	if jarak > float64(lokasiKantor.RadiusMeter) {
		// Kebijakan: Apakah boleh checkout di luar radius?
		// Biasanya boleh tapi statusnya INVALID atau butuh keterangan.
		kehadiran.StatusLokasiPulang = "INVALID"
	}

	if err := h.repo.Update(kehadiran); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan data pulang"})
	}

	return c.JSON(fiber.Map{
		"message": "Check-out berhasil",
		"status":  statusPulang,
		"waktu":   kehadiran.JamPulangReal,
		"jarak":   jarak,
	})
}

func (h *KehadiranHandler) GetHistory(c *fiber.Ctx) error {
	asnID := uint(c.Locals("user_id").(float64))

	history, err := h.repo.GetHistory(asnID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data riwayat"})
	}

	return c.JSON(fiber.Map{
		"message": "Berhasil mengambil riwayat",
		"data":    history,
	})
}

func (h *KehadiranHandler) GetRekap(c *fiber.Ctx) error {
	asnID := uint(c.Locals("user_id").(float64))
	bulan := c.Query("bulan") // Format: "01", "02", ...
	tahun := c.Query("tahun") // Format: "2026"

	if bulan == "" || tahun == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Parameter bulan dan tahun wajib diisi"})
	}

	data, err := h.repo.GetByMonth(asnID, bulan, tahun)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data rekap"})
	}

	// Hitung Statistik
	hadir := 0
	terlambat := 0
	izin := 0
	cuti := 0

	for _, k := range data {
		if k.StatusMasuk == "HADIR" {
			hadir++
		}
		if k.StatusMasuk == "TERLAMBAT" {
			terlambat++
		}
		if k.StatusMasuk == "IZIN" {
			izin++
		}
		if k.StatusMasuk == "CUTI" {
			cuti++
		}
	}

	return c.JSON(fiber.Map{
		"message": "Rekap berhasil",
		"data": fiber.Map{
			"hadir":     hadir,
			"terlambat": terlambat,
			"izin":      izin,
			"cuti":      cuti,
			"detail":    data,
		},
	})
}

// Rumus Haversine untuk menghitung jarak dua titik koordinat (dalam meter)
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Radius bumi dalam meter
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	lat1Rad := lat1 * (math.Pi / 180.0)
	lat2Rad := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
