package handler

import (
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ReportHandler struct {
	jadwalRepo    repository.JadwalRepository
	kehadiranRepo repository.KehadiranRepository
	asnRepo       repository.ASNRepository
}

func NewReportHandler(jadwalRepo repository.JadwalRepository, kehadiranRepo repository.KehadiranRepository, asnRepo repository.ASNRepository) *ReportHandler {
	return &ReportHandler{
		jadwalRepo:    jadwalRepo,
		kehadiranRepo: kehadiranRepo,
		asnRepo:       asnRepo,
	}
}

// GetMonthlyRecap menyediakan data untuk PDF Laporan Kehadiran Pegawai Bulanan
func (h *ReportHandler) GetMonthlyRecap(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64))
	bulan := c.Query("bulan")
	tahun := c.Query("tahun")

	// Pad bulan to 2 digits if needed
	if len(bulan) == 1 {
		bulan = "0" + bulan
	}

	if bulan == "" || tahun == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Bulan dan Tahun wajib diisi"})
	}

	// 1. Ambil Semua Pegawai Organisasi
	// Asumsi ada method update repository untuk GetAllByOrgID
	asns, err := h.asnRepo.GetAllByOrganisasiID(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data pegawai"})
	}

	// 2. Ambil Jadwal & Kehadiran Bulan Ini
	jadwals, _ := h.jadwalRepo.GetByMonth(bulan, tahun, orgID)
	kehadirans, _ := h.kehadiranRepo.GetByMonthAndOrg(bulan, tahun, orgID)

	// Map untuk akses cepat
	// Map[ASNID][Tanggal] = Jadwal
	jadwalMap := make(map[uint]map[string]model.Jadwal)
	for _, j := range jadwals {
		if _, ok := jadwalMap[j.ASNID]; !ok {
			jadwalMap[j.ASNID] = make(map[string]model.Jadwal)
		}
		jadwalMap[j.ASNID][j.Tanggal] = j
	}

	// Map[JadwalID] = Kehadiran
	attendanceMap := make(map[uint]model.Kehadiran)
	for _, k := range kehadirans {
		attendanceMap[k.JadwalID] = k
	}

	// 3. Bangun Struktur Data Laporan
	var reportData []fiber.Map
	daysInMonth := getDaysInMonth(bulan, tahun)

	for _, asn := range asns {
		row := fiber.Map{
			"nip":  asn.NIP,
			"nama": asn.Nama,
		}

		// Counters
		tl, cp, tk, cuti, izin := 0, 0, 0, 0, 0
		t1, t2, t3, t4 := 0, 0, 0, 0
		totalJadwal := 0

		// Generate Daily Codes (01 - 31)
		dailyCodes := make(map[string]string)

		for d := 1; d <= daysInMonth; d++ {
			dateDate := time.Date(parseYear(tahun), time.Month(parseMonth(bulan)), d, 0, 0, 0, 0, time.Local)
			dateStr := dateDate.Format("2006-01-02")
			dayKey := dateDate.Format("02")

			code := "" // Default kosong (Jadwal tidak aktif/Libur/Tidak ada jadwal)

			// Cek Jadwal
			var jadwal model.Jadwal
			hasJadwal := false

			if userJadwal, ok := jadwalMap[asn.ID]; ok {
				if j, exists := userJadwal[dateStr]; exists {
					jadwal = j
					hasJadwal = true
				}
			}

			if hasJadwal && jadwal.IsActive {
				totalJadwal++
				// Cek Kehadiran
				if k, attended := attendanceMap[jadwal.ID]; attended {

					// Cek Validitas Lokasi (User Request: Invalid & No Permit = TK)
					isLokasiValid := k.StatusLokasiMasuk == "VALID"
					hasIzinLokasi := k.PerizinanLokasiID != nil
					isCutiOrIzin := k.StatusMasuk == "CUTI" || k.StatusMasuk == "IZIN"

					if !isLokasiValid && !hasIzinLokasi && !isCutiOrIzin {
						// Lokasi Invalid & Tidak Ada Izin & Bukan Cuti/Izin -> Hitung TK
						code = "-"
						tk++
					} else {
						// Punya data absen & Lokasi Valid/Ada Izin
						if k.StatusMasuk == "CUTI" {
							code = "C"
							cuti++
						} else if k.StatusMasuk == "IZIN" {
							code = "I"
							izin++
						} else if k.StatusMasuk == "HADIR" || k.StatusMasuk == "TERLAMBAT" || k.StatusPulang == "PULANG_CEPAT" {
							code = "H" // Tetap H di tabel

							// Hitung TL / CP hanya jika TIDAK ADA IZIN STATUS (PerizinanKehadiranID == nil)
							if k.PerizinanKehadiranID == nil {
								if k.StatusMasuk == "TERLAMBAT" {
									tl++
									// Hitung Range Keterlambatan
									minutesLate := calculateMinutesLate(jadwal.Shift.JamMasuk, k.JamMasukReal)
									if minutesLate <= 30 {
										t1++
									} else if minutesLate <= 60 {
										t2++
									} else if minutesLate <= 90 {
										t3++
									} else {
										t4++
									}
								}
								if k.StatusPulang == "PULANG_CEPAT" {
									cp++
								}
							}
						}
					}
				} else {
					// Tidak ada absen tapi jadwal aktif -> TK (Tanpa Keterangan)
					// Hanya jika tanggal sudah lewat
					if dateStr < time.Now().Format("2006-01-02") {
						code = "-"
						tk++
					} else {
						// Future or Today (not passed yet) -> Empty White Cell
						code = " "
					}
				}
			}

			dailyCodes[dayKey] = code
			// dailyCodes[dayKey + "_dark"] = isDark // Bisa dikirim ke FE untuk styling css
		}

		row["daily"] = dailyCodes
		row["stats"] = fiber.Map{
			"tl": tl, "cp": cp, "tk": tk, "c": cuti, "i": izin,
			"t1": t1, "t2": t2, "t3": t3, "t4": t4,
			"total_kehadiran": totalJadwal - tk - cuti - izin,
		}

		reportData = append(reportData, row)
	}

	return c.JSON(fiber.Map{
		"organisasi":  "Dinas Komunikasi dan Informatika", // Hardcode dulu atau ambil dari relasi org
		"bulan_tahun": convertMonthToIndonesian(bulan) + " " + tahun,
		"data":        reportData,
		"days_count":  daysInMonth,
	})
}

// GetMonthlyRecapByAtasan menyediakan data rekap bulanan khusus untuk bawahan dari atasan yang login
func (h *ReportHandler) GetMonthlyRecapByAtasan(c *fiber.Ctx) error {
	// Ambil ID Atasan dari Token
	userID := uint(c.Locals("user_id").(float64))
	orgID := uint(c.Locals("organisasi_id").(float64)) // Tetap butuh orgID untuk filter jadwal
	bulan := c.Query("bulan")
	tahun := c.Query("tahun")

	// Pad bulan to 2 digits if needed
	if len(bulan) == 1 {
		bulan = "0" + bulan
	}

	if bulan == "" || tahun == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Bulan dan Tahun wajib diisi"})
	}

	// 1. Ambil List Bawahan (ASN yang atasan_id nya == userID)
	allAsns, err := h.asnRepo.GetByAtasanID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data bawahan"})
	}

	// Filter hanya yang aktif
	var asns []model.ASN
	for _, a := range allAsns {
		if a.IsActive {
			asns = append(asns, a)
		}
	}

	// 2. Ambil Jadwal & Kehadiran Bulan Ini (Scope Organisasi)
	// Kita ambil 1 org dulu, nanti difilter by map
	jadwals, _ := h.jadwalRepo.GetByMonth(bulan, tahun, orgID)
	kehadirans, _ := h.kehadiranRepo.GetByMonthAndOrg(bulan, tahun, orgID)

	// Map untuk akses cepat
	// Map[ASNID][Tanggal] = Jadwal
	jadwalMap := make(map[uint]map[string]model.Jadwal)
	for _, j := range jadwals {
		if _, ok := jadwalMap[j.ASNID]; !ok {
			jadwalMap[j.ASNID] = make(map[string]model.Jadwal)
		}
		jadwalMap[j.ASNID][j.Tanggal] = j
	}

	// Map[JadwalID] = Kehadiran
	attendanceMap := make(map[uint]model.Kehadiran)
	for _, k := range kehadirans {
		attendanceMap[k.JadwalID] = k
	}

	// 3. Bangun Agregat Statistik (Accumulator for ALL subordinates)
	var (
		totalHadirTepatWaktu int
		totalTlCp            int // Terlambat / Pulang Cepat (Tanpa Izin Status)
		totalTlCpDiizinkan   int // Terlambat / Pulang Cepat (Dengan Izin Status)
		totalIzin            int
		totalCuti            int
		totalAlfa            int
		totalBelumAbsen      int // Placeholder / Future
		totalJadwalCount     int
	)

	daysInMonth := getDaysInMonth(bulan, tahun)
	todayStr := time.Now().Format("2006-01-02")

	for _, asn := range asns {
		// Iterate through all days in the month for this ASN
		for d := 1; d <= daysInMonth; d++ {
			dateDate := time.Date(parseYear(tahun), time.Month(parseMonth(bulan)), d, 0, 0, 0, 0, time.Local)
			dateStr := dateDate.Format("2006-01-02")

			// Cek Jadwal
			var jadwal model.Jadwal
			hasJadwal := false

			if userJadwal, ok := jadwalMap[asn.ID]; ok {
				if j, exists := userJadwal[dateStr]; exists {
					jadwal = j
					hasJadwal = true
				}
			}

			if hasJadwal && jadwal.IsActive {
				totalJadwalCount++

				// Cek Kehadiran
				if k, attended := attendanceMap[jadwal.ID]; attended {
					// Cek Validitas Lokasi / Izin Lokasi
					// Cek Validitas Lokasi / Izin Lokasi
					isLokasiValid := k.StatusLokasiMasuk == "VALID"
					hasIzinLokasi := k.PerizinanLokasiID != nil
					isCutiOrIzin := k.StatusMasuk == "CUTI" || k.StatusMasuk == "IZIN"

					if !isLokasiValid && !hasIzinLokasi && !isCutiOrIzin {
						// Invalid Lokasi & No Permit & Not Cuti/Izin -> Treat as TK (Alfa)
						totalAlfa++
					} else {
						// Valid Attendance Logic
						switch k.StatusMasuk {
						case "CUTI":
							totalCuti++
						case "IZIN":
							totalIzin++
						case "HADIR", "TERLAMBAT":
							// Handle Hadir logic
							// Cek Status Izin (PerizinanKehadiranID)

							// Cek apakah ada masalah kehadiran (Telat atau Pulang Cepat)
							hasIssue := k.StatusMasuk == "TERLAMBAT" || k.StatusPulang == "PULANG_CEPAT"

							if k.PerizinanKehadiranID != nil {
								// Ada Izin Status
								if hasIssue {
									totalTlCpDiizinkan++
								} else {
									// Punya izin tapi tidak telat/cp -> Hadir Tepat Waktu
									totalHadirTepatWaktu++
								}
							} else {
								// Tidak Ada Izin Status
								if hasIssue {
									totalTlCp++
								} else {
									// Tidak telat & tidak cp -> Hadir Tepat Waktu
									totalHadirTepatWaktu++
								}
							}
						}
					}
				} else {
					// Tidak ada absen tapi jadwal aktif
					if dateStr < todayStr {
						// Date passed -> TK (Alfa)
						totalAlfa++
					} else {
						// Future or Today -> Belum Absen
						totalBelumAbsen++
					}
				}
			}
		}
	}

	return c.JSON(fiber.Map{
		"message": "Berhasil mengambil statistik",
		"data": fiber.Map{
			"bulan_ini": fiber.Map{
				"hadir_tepat_waktu": totalHadirTepatWaktu,
				"tl_cp":             totalTlCp,
				"tl_cp_diizinkan":   totalTlCpDiizinkan,
				"izin":              totalIzin,
				"cuti":              totalCuti,
				"alfa":              totalAlfa,
				"belum_absen":       totalBelumAbsen,
				"total_jadwal":      totalJadwalCount,
			},
			"total_pegawai": len(asns),
		},
	})
}

// GetDailyRecap menyediakan data untuk PDF Laporan Harian
func (h *ReportHandler) GetDailyRecap(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64))
	tanggal := c.Query("tanggal")

	if tanggal == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Tanggal wajib diisi"})
	}

	// 1. Ambil Jadwal Hari Ini
	jadwals, _ := h.jadwalRepo.GetByDate(tanggal, orgID, "")

	// 2. Ambil Kehadiran Hari Ini
	kehadirans, _ := h.kehadiranRepo.GetByDateAndOrg(tanggal, orgID)
	attendanceMap := make(map[uint]model.Kehadiran)
	for _, k := range kehadirans {
		attendanceMap[k.JadwalID] = k
	}

	var reportData []fiber.Map

	for _, j := range jadwals {
		// Skip jika tidak aktif (Opsional, tergantung request user apakah mau nampilkan yang libur)
		// User request: "bila jadwalnya tidak ada, atau dihapus, atau nonaktif, maka dikosongkan" -> berarti skip di list atau tampil kosong
		// Di contoh PDF Harian, semua pegawai muncul.

		row := fiber.Map{
			"nip":        j.ASN.NIP,
			"nama":       j.ASN.Nama,
			"masuk":      "-",
			"pulang":     "-",
			"keterangan": "",
		}

		if !j.IsActive {
			// Jadwal Libur
			row["keterangan"] = "Libur"
		} else {
			if k, exists := attendanceMap[j.ID]; exists {
				// Cek Validitas Lokasi / Izin Lokasi
				showMasuk := false
				showPulang := false

				// Masuk: Valid Lokasi ATAU Punya Izin Lokasi
				if k.StatusLokasiMasuk == "VALID" || k.PerizinanLokasiID != nil {
					showMasuk = true
				}
				// Pulang: Valid Lokasi ATAU Punya Izin Lokasi
				if k.StatusLokasiPulang == "VALID" || k.PerizinanLokasiID != nil {
					showPulang = true
				}

				if showMasuk {
					row["masuk"] = formatTime(k.JamMasukReal)
				}
				if showPulang {
					row["pulang"] = formatTime(k.JamPulangReal)
				}

				// Keterangan Logic
				if k.StatusMasuk == "IZIN" {
					row["keterangan"] = "Izin"
				} else if k.StatusMasuk == "CUTI" {
					row["keterangan"] = "Cuti"
				} else if k.StatusMasuk == "TERLAMBAT" {
					row["keterangan"] = "TL"
				} else if k.StatusPulang == "PULANG_CEPAT" {
					row["keterangan"] = "CP"
				}

				// Izin Status override keterangan? Atau append?
				// "TK" logic?
			} else {
				// Tidak ada absen -> TK (jika sudah lewat jamnya / tanggalnya)
				// Asumsi report di generate sore/besoknya
				row["keterangan"] = "TK"
			}
		}

		reportData = append(reportData, row)
	}

	return c.JSON(fiber.Map{
		"organisasi":   "Dinas Komunikasi dan Informatika",
		"tanggal_full": convertDateToIndonesian(tanggal),
		"data":         reportData,
	})
}

// Helper Functions

func getDaysInMonth(monthStr, yearStr string) int {
	year := parseYear(yearStr)
	month := time.Month(parseMonth(monthStr))
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func parseYear(y string) int {
	t, _ := time.Parse("2006", y)
	return t.Year()
}

func parseMonth(m string) int {
	t, _ := time.Parse("01", m)
	return int(t.Month())
}

func calculateMinutesLate(scheduleTime, actualTime string) int {
	sched, _ := time.Parse("15:04", scheduleTime)
	act, _ := time.Parse("15:04:05", actualTime) // format jam masuk real biasanya ada detiknya

	// Normalize date
	schedBase := time.Date(2000, 1, 1, sched.Hour(), sched.Minute(), 0, 0, time.UTC)
	actBase := time.Date(2000, 1, 1, act.Hour(), act.Minute(), act.Second(), 0, time.UTC)

	if actBase.After(schedBase) {
		diff := actBase.Sub(schedBase)
		return int(diff.Minutes())
	}
	return 0
}

func formatTime(t string) string {
	parsed, err := time.Parse("15:04:05", t)
	if err != nil {
		return t
	}
	return parsed.Format("15:04")
}

func convertMonthToIndonesian(m string) string {
	months := map[string]string{
		"01": "JANUARI", "02": "FEBRUARI", "03": "MARET", "04": "APRIL",
		"05": "MEI", "06": "JUNI", "07": "JULI", "08": "AGUSTUS",
		"09": "SEPTEMBER", "10": "OKTOBER", "11": "NOVEMBER", "12": "DESEMBER",
	}
	return months[m]
}

func convertDateToIndonesian(dateStr string) string {
	// 2026-01-28 -> RABU, 28 JANUARI 2026
	t, _ := time.Parse("2006-01-02", dateStr)
	days := map[string]string{
		"Sunday": "MINGGU", "Monday": "SENIN", "Tuesday": "SELASA", "Wednesday": "RABU",
		"Thursday": "KAMIS", "Friday": "JUMAT", "Saturday": "SABTU",
	}
	months := map[string]string{
		"January": "JANUARI", "February": "FEBRUARI", "March": "MARET",
		"April": "APRIL", "May": "MEI", "June": "JUNI",
		"July": "JULI", "August": "AGUSTUS", "September": "SEPTEMBER",
		"October": "OKTOBER", "November": "NOVEMBER", "December": "DESEMBER",
	}

	dayName := days[t.Format("Monday")]
	monthName := months[t.Format("January")]
	return dayName + ", " + t.Format("02") + " " + monthName + " " + t.Format("2006")
}
