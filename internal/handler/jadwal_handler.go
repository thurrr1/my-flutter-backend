package handler

import (
	"fmt"
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type JadwalHandler struct {
	repo          repository.JadwalRepository
	hariLiburRepo repository.HariLiburRepository // Tambah ini
	kehadiranRepo repository.KehadiranRepository
	shiftRepo     repository.ShiftRepository
	asnRepo       repository.ASNRepository
}

func NewJadwalHandler(repo repository.JadwalRepository, hlRepo repository.HariLiburRepository, kRepo repository.KehadiranRepository, sRepo repository.ShiftRepository, aRepo repository.ASNRepository) *JadwalHandler {
	return &JadwalHandler{repo: repo, hariLiburRepo: hlRepo, kehadiranRepo: kRepo, shiftRepo: sRepo, asnRepo: aRepo}
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
		ASNID:    req.ASNID,
		ShiftID:  req.ShiftID,
		Tanggal:  req.Tanggal,
		IsActive: true,
	}

	if err := h.repo.Upsert(&jadwal); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan jadwal"})
	}

	return c.JSON(fiber.Map{
		"message": "Jadwal berhasil dibuat",
		"data":    jadwal,
	})
}

type GenerateJadwalRequest struct {
	ASNIDs         []uint `json:"asn_ids"` // UBAH: Array of ID (Checkbox)
	ShiftID        uint   `json:"shift_id"`
	TanggalMulai   string `json:"tanggal_mulai"`
	TanggalSelesai string `json:"tanggal_selesai"`
	Days           []int  `json:"days"` // 0=Minggu, 1=Senin, ..., 6=Sabtu
}

func (h *JadwalHandler) GenerateJadwalBulanan(c *fiber.Ctx) error {
	var req GenerateJadwalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	var listJadwal []model.Jadwal

	startDate, err := time.Parse("2006-01-02", req.TanggalMulai)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format tanggal mulai salah"})
	}
	endDate, err := time.Parse("2006-01-02", req.TanggalSelesai)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format tanggal selesai salah"})
	}

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
				ASNID:    asnID,
				ShiftID:  req.ShiftID,
				Tanggal:  d.Format("2006-01-02"),
				IsActive: true,
			}
			listJadwal = append(listJadwal, jadwal)
		}
	}

	if len(listJadwal) > 0 {
		countSuccess := 0
		for _, j := range listJadwal {
			if err := h.repo.Upsert(&j); err == nil {
				countSuccess++
			}
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
			ASNID:    asnID,
			ShiftID:  req.ShiftID,
			Tanggal:  req.Tanggal,
			IsActive: true,
		}
		listJadwal = append(listJadwal, jadwal)
	}

	if len(listJadwal) > 0 {
		for _, j := range listJadwal {
			h.repo.Upsert(&j)
		}
	}
	return c.JSON(fiber.Map{"message": "Berhasil membuat jadwal harian"})
}

type ImportJadwalItem struct {
	NIP       string `json:"nip"`
	Tanggal   string `json:"tanggal"`    // YYYY-MM-DD
	JamMasuk  string `json:"jam_masuk"`  // HH:mm
	JamPulang string `json:"jam_pulang"` // HH:mm
	IsActive  bool   `json:"is_active"`
}

func (h *JadwalHandler) ImportJadwal(c *fiber.Ctx) error {
	var reqs []ImportJadwalItem
	if err := c.BodyParser(&reqs); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format data tidak valid"})
	}

	if len(reqs) == 0 {
		return c.JSON(fiber.Map{"message": "Tidak ada data untuk diimport"})
	}

	// ---------------------------------------------------------
	// OPTIMISASI CACHING REFERENCE DATA (ASN & SHIFT)
	// ---------------------------------------------------------

	// 1. Cache Semua Shift (Key: "JamMasuk-JamPulang", Value: ID)
	shiftCache := make(map[string]uint)
	shifts, _ := h.shiftRepo.GetAll()
	for _, s := range shifts {
		key := fmt.Sprintf("%s-%s", s.JamMasuk, s.JamPulang)
		shiftCache[key] = s.ID
	}

	// 2. Cache Semua ASN (Key: NIP, Value: ID)
	// NOTE: Jika ASN sangat banyak (>10k), sebaiknya query `WhereIn("nip", requestedNIPs)`
	// Tapi untuk skala < 5k pegawai, load all masih oke dan cepat.
	asnCache := make(map[string]uint)
	asns, _ := h.asnRepo.GetAll("") // Empty search fetches all
	for _, a := range asns {
		asnCache[a.NIP] = a.ID
	}

	// ---------------------------------------------------------
	// PREPARE DATA FOR BATCH INSERT
	// ---------------------------------------------------------

	var jadwalsToUpsert []model.Jadwal
	newShiftsMap := make(map[string]*model.Shift) // Temp storage untuk shift baru yang belum di DB

	for _, item := range reqs {
		// A. Lookup ASN ID
		asnID, exists := asnCache[item.NIP]
		if !exists {
			continue // Skip jika NIP tidak valid/tidak ditemukan
		}

		// B. Lookup / Prepare Shift ID
		shiftKey := fmt.Sprintf("%s-%s", item.JamMasuk, item.JamPulang)
		shiftID, exists := shiftCache[shiftKey]

		if !exists {
			// Cek apakah sudah kita queue untuk dibuat di iterasi sebelumnya?
			if s, inQueue := newShiftsMap[shiftKey]; inQueue {
				// Sudah ada di map sementara, (ID masih 0, tapi ini complex untuk batch insert relasi)
				// KEEP SIMPLE: Jika shift baru, insert direct one-by-one.
				// Jarang terjadi mass creation shift lewat import jadwal biasanya.
				err := h.shiftRepo.Create(s)
				if err == nil {
					shiftID = s.ID
					shiftCache[shiftKey] = s.ID    // Update cache
					delete(newShiftsMap, shiftKey) // Hapus dari queue
				}
			} else {
				// Belum ada sama sekali, buat object baru
				newShift := model.Shift{
					NamaShift: shiftKey,
					JamMasuk:  item.JamMasuk,
					JamPulang: item.JamPulang,
				}
				// Insert langsung agar dapat ID
				if err := h.shiftRepo.Create(&newShift); err == nil {
					shiftID = newShift.ID
					shiftCache[shiftKey] = newShift.ID
				}
			}
		}

		// Jika Shiftgagal dibuat/ditemukan, skip
		if shiftID == 0 {
			continue
		}

		// C. Append to Batch List
		jadwalsToUpsert = append(jadwalsToUpsert, model.Jadwal{
			ASNID:    asnID,
			ShiftID:  shiftID,
			Tanggal:  item.Tanggal,
			IsActive: item.IsActive,
		})
	}

	// ---------------------------------------------------------
	// EXECUTE BATCH UPSERT
	// ---------------------------------------------------------
	if len(jadwalsToUpsert) > 0 {
		// Bagi menjadi chunk jika sangat besar (misal max 1000 per insert)
		// GORM biasanya handle ini, tapi bisa kita bantu manual kalau mau.
		// Untuk sekarang langsung upsert semua saja.
		if err := h.repo.UpsertBatch(jadwalsToUpsert); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan data jadwal: " + err.Error()})
		}
	}

	return c.JSON(fiber.Map{
		"message":         fmt.Sprintf("Berhasil memproses %d baris, %d jadwal tersimpan/diupdate", len(reqs), len(jadwalsToUpsert)),
		"total_processed": len(jadwalsToUpsert),
	})
}

type JadwalWithStatus struct {
	model.Jadwal
	StatusKehadiran string `json:"status_kehadiran"`
	JamMasukReal    string `json:"jam_masuk_real"`
	JamPulangReal   string `json:"jam_pulang_real"`
}

// GET /api/admin/jadwal?tanggal=2024-10-25
func (h *JadwalHandler) GetJadwalHarian(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64))
	tanggal := c.Query("tanggal")
	search := c.Query("search")

	if tanggal == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Parameter tanggal wajib diisi"})
	}

	jadwals, err := h.repo.GetByDate(tanggal, orgID, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil jadwal"})
	}

	// Ambil Data Kehadiran pada tanggal tersebut untuk Organisasi ini
	kehadirans, err := h.kehadiranRepo.GetByDateAndOrg(tanggal, orgID)
	kehadiranMap := make(map[uint]model.Kehadiran)
	if err == nil {
		for _, k := range kehadirans {
			kehadiranMap[k.ASNID] = k
		}
	}

	// Gabungkan Jadwal dengan Status Kehadiran
	response := make([]JadwalWithStatus, 0)
	today, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	targetDate, _ := time.Parse("2006-01-02", tanggal)

	for _, j := range jadwals {
		status := "BELUM ABSEN"
		if !j.IsActive {
			status = "NONAKTIF"
		}

		jamMasuk := ""
		jamPulang := ""

		if k, exists := kehadiranMap[j.ASNID]; exists {
			jamMasuk = k.JamMasukReal
			jamPulang = k.JamPulangReal

			if k.StatusMasuk == "IZIN" {
				status = "IZIN"
			} else if k.StatusMasuk == "CUTI" {
				status = "CUTI"
			} else if k.StatusMasuk == "TERLAMBAT" || k.StatusPulang == "PULANG_CEPAT" {
				status = "TERLAMBAT"
				if k.PerizinanKehadiranID != nil {
					status += " (Diizinkan)"
				}
			} else {
				status = "HADIR"
			}
		} else {
			// Jika tidak ada data kehadiran dan tanggal sudah lewat -> ALPHA
			// HANYA JIKA JADWAL AKTIF
			if j.IsActive && targetDate.Before(today) {
				status = "ALPHA"
			}
		}

		response = append(response, JadwalWithStatus{
			Jadwal:          j,
			StatusKehadiran: status,
			JamMasukReal:    jamMasuk,
			JamPulangReal:   jamPulang,
		})
	}

	return c.JSON(fiber.Map{"data": response})
}

// GET /api/jadwal/saya?bulan=02&tahun=2026
func (h *JadwalHandler) GetJadwalSaya(c *fiber.Ctx) error {
	asnID := uint(c.Locals("user_id").(float64))
	now := time.Now()
	bulan := c.Query("bulan", now.Format("01"))
	tahun := c.Query("tahun", now.Format("2006"))

	// 1. Ambil Jadwal Saya Bulan Ini
	jadwals, err := h.repo.GetByASNAndMonth(asnID, bulan, tahun)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil jadwal"})
	}

	// 2. Ambil Riwayat Kehadiran Bulan Ini
	kehadirans, err := h.kehadiranRepo.GetByMonth(asnID, bulan, tahun)
	kehadiranMap := make(map[string]model.Kehadiran) // Key by Tanggal YYYY-MM-DD
	if err == nil {
		for _, k := range kehadirans {
			kehadiranMap[k.Tanggal] = k
		}
	}

	// 3. Gabungkan
	var response []fiber.Map
	today := now.Format("2006-01-02")

	for _, j := range jadwals {
		status := "BELUM ABSEN"
		jamMasuk := ""
		jamPulang := ""

		if !j.IsActive {
			status = "LIBUR"
		} else {
			if k, exists := kehadiranMap[j.Tanggal]; exists {
				jamMasuk = k.JamMasukReal
				jamPulang = k.JamPulangReal

				if k.StatusMasuk == "IZIN" {
					status = "IZIN"
				} else if k.StatusMasuk == "CUTI" {
					status = "CUTI"
				} else if k.StatusMasuk == "TERLAMBAT" || k.StatusPulang == "PULANG_CEPAT" {
					status = "TERLAMBAT"
					if k.PerizinanKehadiranID != nil {
						status += " (Diizinkan)"
					}
				} else {
					status = "HADIR"
				}
			} else {
				if j.Tanggal < today {
					status = "ALFA"
				}
			}
		}

		response = append(response, fiber.Map{
			"id":               j.ID,
			"tanggal":          j.Tanggal,
			"nama_shift":       j.Shift.NamaShift,
			"jam_masuk_shift":  j.Shift.JamMasuk,
			"jam_pulang_shift": j.Shift.JamPulang,
			"status":           status,
			"jam_masuk_real":   jamMasuk,
			"jam_pulang_real":  jamPulang,
		})
	}

	return c.JSON(fiber.Map{
		"message": "Berhasil mengambil jadwal saya",
		"data":    response,
		"meta": fiber.Map{
			"bulan": bulan,
			"tahun": tahun,
		},
	})
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
	ShiftID  uint  `json:"shift_id"`
	IsActive *bool `json:"is_active"` // Pointer agar bisa deteksi true/false
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

	if req.ShiftID != 0 {
		jadwal.ShiftID = req.ShiftID
	}
	if req.IsActive != nil {
		jadwal.IsActive = *req.IsActive
	}

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

func (h *JadwalHandler) GetDashboardStats(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64))

	// Filter Bulan & Tahun
	now := time.Now()
	bulan := c.Query("bulan")
	if bulan == "" {
		bulan = now.Format("01")
	} else if len(bulan) == 1 {
		bulan = "0" + bulan
	}

	tahun := c.Query("tahun")
	if tahun == "" {
		tahun = now.Format("2006")
	}

	today := now.Format("2006-01-02")

	// 1. Ambil Semua Jadwal Bulan Ini
	// Pastikan repository memiliki method GetByMonth(bulan, tahun, orgID)
	jadwals, err := h.repo.GetByMonth(bulan, tahun, orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data jadwal"})
	}

	// 2. Ambil Semua Kehadiran Bulan Ini
	// Pastikan repository memiliki method GetByMonthAndOrg(bulan, tahun, orgID)
	kehadirans, err := h.kehadiranRepo.GetByMonthAndOrg(bulan, tahun, orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data kehadiran"})
	}

	// Map Kehadiran by JadwalID untuk akses cepat
	attendanceMap := make(map[uint]model.Kehadiran)
	for _, k := range kehadirans {
		if k.JadwalID != 0 {
			attendanceMap[k.JadwalID] = k
		}
	}

	// Inisialisasi Counters
	// Gunakan variabel terpisah agar mudah dihitung
	totalJadwal := len(jadwals)
	hadir := 0
	tlCp := 0
	tlCpIzin := 0
	izin := 0
	cuti := 0
	alfa := 0
	belumAbsen := 0

	statsHari := map[string]int{"hadir_tepat_waktu": 0, "tl_cp": 0, "tl_cp_diizinkan": 0, "izin": 0, "cuti": 0, "alfa": 0, "belum_absen": 0}

	var details []fiber.Map

	for _, j := range jadwals {
		k, exists := attendanceMap[j.ID]

		// Tentukan status untuk detail list
		statusMasuk := "BELUM_ABSEN"
		statusPulang := ""

		if exists {
			statusMasuk = k.StatusMasuk
			statusPulang = k.StatusPulang

			if k.StatusMasuk == "IZIN" {
				izin++
			} else if k.StatusMasuk == "CUTI" {
				cuti++
			} else if k.StatusMasuk == "TERLAMBAT" || k.StatusPulang == "PULANG_CEPAT" {
				if k.PerizinanKehadiranID != nil {
					tlCpIzin++
				} else {
					tlCp++
				}
				statusMasuk = "TERLAMBAT" // Tetap TERLAMBAT agar terdeteksi frontend
			} else {
				hadir++
			}
		} else {
			if j.Tanggal < today {
				statusMasuk = "ALFA"
				alfa++
			} else {
				belumAbsen++
			}
		}

		// Tambahkan ke detail list untuk grafik harian
		details = append(details, fiber.Map{
			"tanggal":                j.Tanggal,
			"status_masuk":           statusMasuk,
			"status_pulang":          statusPulang,
			"perizinan_kehadiran_id": k.PerizinanKehadiranID,
		})

		// Hitung Statistik Harian (Hanya jika tanggal jadwal == hari ini)
		if j.Tanggal == today {
			if exists {
				if k.StatusMasuk == "IZIN" {
					statsHari["izin"]++
				} else if k.StatusMasuk == "CUTI" {
					statsHari["cuti"]++
				} else if k.StatusMasuk == "TERLAMBAT" || k.StatusPulang == "PULANG_CEPAT" {
					if k.PerizinanKehadiranID != nil {
						statsHari["tl_cp_diizinkan"]++
					} else {
						statsHari["tl_cp"]++
					}
				} else {
					statsHari["hadir_tepat_waktu"]++
				}
			} else {
				statsHari["belum_absen"]++
			}
		}
	}

	// Hitung Persentase Kehadiran Bulanan (Hadir + TL/CP) / (Total Jadwal - Belum Absen)
	totalSudahLewat := totalJadwal - belumAbsen
	persentaseHadir := 0.0
	if totalSudahLewat > 0 {
		hadirCount := hadir + tlCp + tlCpIzin
		persentaseHadir = (float64(hadirCount) / float64(totalSudahLewat)) * 100
	}

	// Nama Bulan untuk Meta
	mInt, _ := strconv.Atoi(bulan)
	monthName := time.Month(mInt).String()

	return c.JSON(fiber.Map{
		"bulan_ini": fiber.Map{
			"total_jadwal":      totalJadwal,
			"hadir_tepat_waktu": hadir,
			"tl_cp":             tlCp,
			"tl_cp_diizinkan":   tlCpIzin,
			"izin":              izin,
			"cuti":              cuti,
			"alfa":              alfa,
			"belum_absen":       belumAbsen,
			"detail":            details,
		},
		"hari_ini":         statsHari,
		"persentase_hadir": fmt.Sprintf("%.1f", persentaseHadir),
		"meta": fiber.Map{
			"bulan": monthName,
			"tahun": tahun,
		},
	})
}
