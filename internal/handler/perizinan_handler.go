package handler

import (
	"fmt"
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type PerizinanHandler struct {
	repo          repository.PerizinanRepository
	kehadiranRepo repository.KehadiranRepository
	asnRepo       repository.ASNRepository
	jadwalRepo    repository.JadwalRepository
}

func NewPerizinanHandler(repo repository.PerizinanRepository, kRepo repository.KehadiranRepository, asnRepo repository.ASNRepository, jadwalRepo repository.JadwalRepository) *PerizinanHandler {
	return &PerizinanHandler{repo: repo, kehadiranRepo: kRepo, asnRepo: asnRepo, jadwalRepo: jadwalRepo}
}

type PengajuanIzinRequest struct {
	Tipe           string `json:"tipe"`
	JenisIzin      string `json:"jenis_izin"`
	TanggalMulai   string `json:"tanggal_mulai"`
	TanggalSelesai string `json:"tanggal_selesai"`
	Keterangan     string `json:"keterangan"`
}

func (h *PerizinanHandler) AjukanIzin(c *fiber.Ctx) error {
	asnID := uint(c.Locals("user_id").(float64))

	var req PengajuanIzinRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	// Validasi Format Tanggal (YYYY-MM-DD)
	if _, err := time.Parse("2006-01-02", req.TanggalMulai); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format tanggal mulai salah (Gunakan YYYY-MM-DD)"})
	}
	if _, err := time.Parse("2006-01-02", req.TanggalSelesai); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format tanggal selesai salah (Gunakan YYYY-MM-DD)"})
	}

	// Handle File Upload (Bukti Izin)
	file, errFile := c.FormFile("file_bukti")
	pathFile := ""
	if errFile == nil {
		// Buat folder jika belum ada
		uploadDir := "./uploads/perizinan"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.MkdirAll(uploadDir, 0755)
		}

		// Simpan file: uploads/perizinan/asnID_timestamp_namafile
		filename := fmt.Sprintf("%d_%d_%s", asnID, time.Now().Unix(), filepath.Base(file.Filename))
		pathFile = fmt.Sprintf("uploads/perizinan/%s", filename)

		c.SaveFile(file, pathFile)
	}

	// Ambil data ASN untuk mendapatkan NIP Atasan
	asn, err := h.asnRepo.FindByID(asnID)
	nipAtasan := ""
	if err == nil && asn.Atasan != nil {
		nipAtasan = asn.Atasan.NIP
	}

	izin := model.PerizinanCuti{
		ASNID:          asnID,
		NIPAtasan:      nipAtasan,
		Tipe:           req.Tipe,
		Jenis:          req.JenisIzin,
		TanggalMulai:   req.TanggalMulai,
		TanggalSelesai: req.TanggalSelesai,
		Alasan:         req.Keterangan,
		Status:         "MENUNGGU",
		PathFile:       pathFile,
	}

	if err := h.repo.CreateCuti(&izin); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengajukan izin"})
	}

	return c.JSON(fiber.Map{
		"message": "Pengajuan izin berhasil dikirim",
		"data":    izin,
	})
}

func (h *PerizinanHandler) GetRiwayatIzin(c *fiber.Ctx) error {
	asnID := uint(c.Locals("user_id").(float64))

	list, err := h.repo.GetByASNID(asnID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data"})
	}

	return c.JSON(fiber.Map{
		"message": "Berhasil mengambil riwayat izin",
		"data":    list,
	})
}

func (h *PerizinanHandler) GetPengajuanBawahan(c *fiber.Ctx) error {
	// ID user yang login (sebagai Atasan)
	atasanID := uint(c.Locals("user_id").(float64))

	list, err := h.repo.GetByAtasanID(atasanID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data pengajuan bawahan"})
	}

	return c.JSON(fiber.Map{
		"message": "Berhasil mengambil daftar pengajuan bawahan",
		"data":    list,
	})
}

type ApprovalRequest struct {
	PerizinanID uint   `json:"perizinan_id"`
	Status      string `json:"status"` // "DISETUJUI" atau "DITOLAK"
}

func (h *PerizinanHandler) ProcessApproval(c *fiber.Ctx) error {
	var req ApprovalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	// Ambil data izin
	izin, err := h.repo.GetByID(req.PerizinanID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Data perizinan tidak ditemukan"})
	}

	// Validasi: Pastikan yang approve adalah Atasan yang sesuai
	nipUser := c.Locals("nip").(string)
	roleUser := c.Locals("role").(string)

	// Kita izinkan Admin untuk override (jaga-jaga), tapi utamanya harus Atasan yang bersangkutan
	if izin.NIPAtasan != nipUser && roleUser != "Admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Anda bukan atasan dari pegawai ini"})
	}

	// Update Status
	izin.Status = req.Status
	if err := h.repo.Update(izin); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update status"})
	}

	// Jika DISETUJUI, generate data kehadiran otomatis
	if req.Status == "DISETUJUI" {
		go h.generateKehadiran(izin) // Jalankan di background (goroutine) agar respon cepat
	}

	return c.JSON(fiber.Map{"message": "Status perizinan berhasil diperbarui"})
}

// normalizeDateString menangani variasi format seperti "YYYY-M-D" dan mengubahnya menjadi "YYYY-MM-DD".
func normalizeDateString(dateStr string) string {
	parts := strings.Split(dateStr, "-")
	if len(parts) != 3 {
		return dateStr // Kembalikan apa adanya jika format tidak bisa di-parse
	}
	if len(parts[1]) == 1 {
		parts[1] = "0" + parts[1] // Tambahkan 0 di depan bulan
	}
	if len(parts[2]) == 1 {
		parts[2] = "0" + parts[2] // Tambahkan 0 di depan hari
	}
	return strings.Join(parts, "-")
}

// Helper untuk generate data kehadiran berhari-hari
func (h *PerizinanHandler) generateKehadiran(izin *model.PerizinanCuti) {
	// Normalisasi string tanggal sebelum di-parse untuk menangani format seperti "2026-01-9"
	startDate, err := time.Parse("2006-01-02", normalizeDateString(izin.TanggalMulai))
	if err != nil {
		fmt.Printf("Error parsing start date for Izin ID %d: %v\n", izin.ID, err)
		return
	}
	endDate, err := time.Parse("2006-01-02", normalizeDateString(izin.TanggalSelesai))
	if err != nil {
		fmt.Printf("Error parsing end date for Izin ID %d: %v\n", izin.ID, err)
		return
	}

	// Loop dari tanggal mulai sampai selesai
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")

		// Cari Jadwal pada tanggal tersebut untuk mengisi JadwalID
		jadwal, err := h.jadwalRepo.GetByASNAndDate(izin.ASNID, dateStr)

		// [FIX] HANYA buat record Kehadiran (IZIN/CUTI) jika ada jadwal kerja pada hari tersebut.
		if err != nil {
			continue
		}

		status := "IZIN"
		if izin.Tipe == "CUTI" {
			status = "CUTI"
		}

		// Cek apakah sudah ada data kehadiran (misal: ALFA atau sudah absen)
		existing, err := h.kehadiranRepo.GetByDate(izin.ASNID, dateStr)
		if err == nil && existing.ID != 0 {
			// UPDATE: Jika sudah ada, update statusnya menjadi IZIN/CUTI
			existing.StatusMasuk = status
			existing.StatusPulang = status
			existing.PerizinanCutiID = &izin.ID
			h.kehadiranRepo.Update(existing)
		} else {
			// CREATE: Jika belum ada, buat baru
			k := model.Kehadiran{
				ASNID:           izin.ASNID,
				PerizinanCutiID: &izin.ID,
				JadwalID:        jadwal.ID, // JadwalID dijamin valid karena jadwal != nil
				Tanggal:         dateStr,
				Tahun:           d.Format("2006"),
				Bulan:           d.Format("01"),
				StatusMasuk:     status,
				StatusPulang:    status,
			}
			h.kehadiranRepo.Create(k)
		}
	}
}
