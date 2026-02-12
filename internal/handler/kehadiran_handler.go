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
	jadwalRepo repository.JadwalRepository     // Tambah ini
	orgRepo    repository.OrganisasiRepository // Tambah ini (Organisasi)
}

func NewKehadiranHandler(repo repository.KehadiranRepository, asnRepo repository.ASNRepository, jadwalRepo repository.JadwalRepository, orgRepo repository.OrganisasiRepository) *KehadiranHandler {
	return &KehadiranHandler{repo: repo, asnRepo: asnRepo, jadwalRepo: jadwalRepo, orgRepo: orgRepo}
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

	// 4. Ambil Semua Lokasi Kantor & Validasi Radius
	org, err := h.orgRepo.GetByID(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Data organisasi tidak ditemukan"})
	}

	statusLokasiMasuk := "INVALID"
	minJarak := math.MaxFloat64
	var validLokasiID *uint

	if len(org.Lokasis) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Belum ada lokasi kantor yang disetting"})
	}

	for i := range org.Lokasis {
		loc := &org.Lokasis[i]
		jarak := calculateDistance(req.Latitude, req.Longitude, loc.Latitude, loc.Longitude)

		if jarak <= float64(loc.RadiusMeter) {
			statusLokasiMasuk = "VALID"
			validLokasiID = &loc.ID
			minJarak = jarak
			break // Found valid location, stop searching
		}

		// Keep track of closest location even if invalid
		if jarak < minJarak {
			minJarak = jarak
		}
	}

	// Use closest distance for reporting
	jarak := minJarak

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
		JadwalID:          jadwal.ID,     // Simpan ID Jadwal
		LokasiID:          validLokasiID, // Simpan Lokasi Valid (jika ada)
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

	var req CheckInRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	// 2. Cek Apakah Sudah Check-in di HARI INI
	var attendance *model.Kehadiran
	attendance, err := h.repo.GetTodayAttendance(asnID)

	// Jika tidak ada check-in hari ini, CEK KEMARIN (Logic Lintas Hari)
	if err != nil || attendance == nil {
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		prevAttendance, errPrev := h.repo.GetByDate(asnID, yesterday)

		// Jika kemarin ada check-in DAN belum check-out -> Kita anggap ini checkout untuk shift kemarin
		if errPrev == nil && prevAttendance != nil && prevAttendance.JamPulangReal == "" {
			attendance = prevAttendance
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Anda belum melakukan Check-in (Hari ini maupun Shift kemarin)"})
		}
	}

	if attendance.StatusMasuk == "IZIN" || attendance.StatusMasuk == "CUTI" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Anda sedang Cuti/Izin, tidak perlu Check-out"})
	}

	if attendance.JamPulangReal != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Anda sudah melakukan Check-out"})
	}

	// 3. Ambil Jadwal Sesuai Tanggal Absensi (Penting untuk Shift Lintas Hari)
	// Kita gunakan tanggal dari record attendance, BUKAN time.Now()
	jadwal, err := h.jadwalRepo.GetByASNAndDate(asnID, attendance.Tanggal)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Jadwal kerja tidak ditemukan."})
	}

	// 4. Validasi Lokasi (Multi-Location)
	org, err := h.orgRepo.GetByID(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Data organisasi tidak ditemukan"})
	}

	if len(org.Lokasis) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Belum ada lokasi kantor yang disetting"})
	}

	statusLokasiPulang := "INVALID"
	minJarak := math.MaxFloat64

	for i := range org.Lokasis {
		loc := &org.Lokasis[i]
		jarak := calculateDistance(req.Latitude, req.Longitude, loc.Latitude, loc.Longitude)

		if jarak <= float64(loc.RadiusMeter) {
			statusLokasiPulang = "VALID"
			minJarak = jarak
			break
		}

		if jarak < minJarak {
			minJarak = jarak
		}
	}

	jarak := minJarak

	// 5. Update Data Pulang
	now := time.Now()
	attendance.JamPulangReal = now.Format("15:04:05")
	attendance.KoordinatPulang = fmt.Sprintf("%f,%f", req.Latitude, req.Longitude)

	// Tentukan Status Pulang
	// Perbandingan waktu harus MEMPERHATIKAN TANGGAL
	// Parse Jam Pulang Shift
	jamPulangShift, _ := time.Parse("15:04", jadwal.Shift.JamPulang)

	// Konstruksi Waktu Pulang Seharusnya
	// Default: Tanggal sama dengan Tanggal Jadwal
	tglJadwal, _ := time.Parse("2006-01-02", attendance.Tanggal)
	waktuPulangShift := time.Date(tglJadwal.Year(), tglJadwal.Month(), tglJadwal.Day(), jamPulangShift.Hour(), jamPulangShift.Minute(), 0, 0, time.Local)

	// Logic Cross-Day: Jika Jam Pulang <= Jam Masuk, berarti shift berakhir di hari berikutnya (H+1) (Support shift 24 jam)
	jamMasukShift, _ := time.Parse("15:04", jadwal.Shift.JamMasuk)
	if jamPulangShift.Before(jamMasukShift) || jamPulangShift.Equal(jamMasukShift) {
		waktuPulangShift = waktuPulangShift.AddDate(0, 0, 1) // Tambah 1 hari
	}

	// Bandingkan Now dengan Waktu Pulang Seharusnya
	statusPulang := "PULANG"
	if now.Before(waktuPulangShift) {
		statusPulang = "PULANG_CEPAT"
	}
	attendance.StatusPulang = statusPulang

	// Validasi Radius Pulang
	attendance.StatusLokasiPulang = statusLokasiPulang

	if err := h.repo.Update(attendance); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan data pulang"})
	}

	return c.JSON(fiber.Map{
		"message":       "Check-out berhasil",
		"status":        statusPulang,
		"waktu":         attendance.JamPulangReal,
		"jarak":         jarak,
		"tanggal_absen": attendance.Tanggal,
	})
}

func (h *KehadiranHandler) GetHistory(c *fiber.Ctx) error {
	asnID := uint(c.Locals("user_id").(float64))
	bulan := c.Query("bulan")
	tahun := c.Query("tahun")

	var history []model.Kehadiran
	var err error

	if bulan != "" && tahun != "" {
		history, err = h.repo.GetByMonth(asnID, bulan, tahun)
	} else {
		history, err = h.repo.GetHistory(asnID)
	}

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

func (h *KehadiranHandler) GetTodayStatus(c *fiber.Ctx) error {
	asnID := uint(c.Locals("user_id").(float64))
	today := time.Now().Format("2006-01-02")

	// 1. Ambil Jadwal Hari Ini (PENTING: Agar aplikasi tahu shift terbaru secara realtime)
	jadwal, errJadwal := h.jadwalRepo.GetByASNAndDate(asnID, today)
	var jadwalInfo interface{} = nil

	if errJadwal == nil && jadwal != nil {
		jadwalInfo = fiber.Map{
			"id":         jadwal.ID,
			"shift_id":   jadwal.ShiftID,
			"nama_shift": jadwal.Shift.NamaShift,
			"jam_masuk":  jadwal.Shift.JamMasuk,
			"jam_pulang": jadwal.Shift.JamPulang,
		}
	}

	// 2. Cek Status Kehadiran
	kehadiran, err := h.repo.GetTodayAttendance(asnID)

	// Jika belum absen (record tidak ditemukan), return status khusus tapi bukan error 500
	if err != nil {
		return c.JSON(fiber.Map{
			"message": "Belum ada data kehadiran hari ini",
			"status":  "BELUM_ABSEN",
			"data":    nil,
			"jadwal":  jadwalInfo,
		})
	}

	return c.JSON(fiber.Map{
		"message": "Data kehadiran ditemukan",
		"status":  kehadiran.StatusMasuk, // HADIR, TERLAMBAT, IZIN, CUTI
		"data":    kehadiran,
		"jadwal":  jadwalInfo,
	})
}

func (h *KehadiranHandler) CheckLocationValidity(c *fiber.Ctx) error {
	// 1. Ambil Data User
	orgID := uint(c.Locals("organisasi_id").(float64))

	var req CheckInRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	// 2. Ambil Organisasi & Lokasi
	org, err := h.orgRepo.GetByID(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Data organisasi tidak ditemukan"})
	}

	if len(org.Lokasis) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Belum ada lokasi kantor yang disetting"})
	}

	// 3. Cek Lokasi (Multi-Location Logic)
	statusLokasi := "INVALID"
	minJarak := math.MaxFloat64
	var lokasiTerdekat interface{}

	for i := range org.Lokasis {
		loc := &org.Lokasis[i]
		jarak := calculateDistance(req.Latitude, req.Longitude, loc.Latitude, loc.Longitude)

		if jarak <= float64(loc.RadiusMeter) {
			statusLokasi = "VALID"
			minJarak = jarak
			lokasiTerdekat = fiber.Map{
				"id":           loc.ID,
				"nama_lokasi":  loc.NamaLokasi,
				"alamat":       loc.Alamat,
				"latitude":     loc.Latitude,
				"longitude":    loc.Longitude,
				"radius_meter": loc.RadiusMeter,
			}
			break // Found valid location
		}

		if jarak < minJarak {
			minJarak = jarak
			lokasiTerdekat = fiber.Map{
				"id":           loc.ID,
				"nama_lokasi":  loc.NamaLokasi,
				"alamat":       loc.Alamat,
				"latitude":     loc.Latitude,
				"longitude":    loc.Longitude,
				"radius_meter": loc.RadiusMeter,
			}
		}
	}

	return c.JSON(fiber.Map{
		"message":         "Pengecekan lokasi berhasil",
		"status_lokasi":   statusLokasi,
		"jarak_terdekat":  minJarak,
		"lokasi_terdekat": lokasiTerdekat,
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
