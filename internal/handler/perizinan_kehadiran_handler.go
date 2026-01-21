package handler

import (
	"fmt"
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

type PerizinanKehadiranHandler struct {
	repo          repository.PerizinanKehadiranRepository
	kehadiranRepo repository.KehadiranRepository
	asnRepo       repository.ASNRepository
}

func NewPerizinanKehadiranHandler(repo repository.PerizinanKehadiranRepository, kRepo repository.KehadiranRepository, asnRepo repository.ASNRepository) *PerizinanKehadiranHandler {
	return &PerizinanKehadiranHandler{repo: repo, kehadiranRepo: kRepo, asnRepo: asnRepo}
}

func (h *PerizinanKehadiranHandler) AjukanKoreksi(c *fiber.Ctx) error {
	asnID := uint(c.Locals("user_id").(float64))

	tanggal := c.FormValue("tanggal_kehadiran")
	tipe := c.FormValue("tipe_koreksi")
	alasan := c.FormValue("alasan")

	if tanggal == "" || tipe == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Tanggal dan Tipe Koreksi wajib diisi"})
	}

	// Handle File Upload (Bukti Koreksi)
	file, errFile := c.FormFile("file_bukti")
	pathFile := ""
	if errFile == nil {
		uploadDir := "./uploads/perizinan_kehadiran"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.MkdirAll(uploadDir, 0755)
		}

		filename := fmt.Sprintf("%d_%d_%s", asnID, time.Now().Unix(), filepath.Base(file.Filename))
		pathFile = fmt.Sprintf("uploads/perizinan_kehadiran/%s", filename)

		c.SaveFile(file, pathFile)
	}

	// Ambil NIP Atasan
	asn, err := h.asnRepo.FindByID(asnID)
	nipAtasan := ""
	if err == nil && asn.Atasan != nil {
		nipAtasan = asn.Atasan.NIP
	}

	koreksi := model.PerizinanKehadiran{
		ASNID:            asnID,
		NIPAtasan:        nipAtasan,
		TanggalKehadiran: tanggal,
		TipeKoreksi:      tipe,
		Alasan:           alasan,
		Status:           "MENUNGGU",
		PathFile:         pathFile,
	}

	if err := h.repo.Create(&koreksi); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengajukan koreksi"})
	}

	return c.JSON(fiber.Map{
		"message": "Pengajuan koreksi berhasil dikirim",
		"data":    koreksi,
	})
}

func (h *PerizinanKehadiranHandler) GetRiwayat(c *fiber.Ctx) error {
	asnID := uint(c.Locals("user_id").(float64))
	list, err := h.repo.GetByASNID(asnID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data"})
	}
	return c.JSON(fiber.Map{"data": list})
}

func (h *PerizinanKehadiranHandler) GetBawahan(c *fiber.Ctx) error {
	atasanID := uint(c.Locals("user_id").(float64))
	list, err := h.repo.GetByAtasanID(atasanID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data"})
	}
	return c.JSON(fiber.Map{"data": list})
}

type ApprovalKoreksiRequest struct {
	KoreksiID uint   `json:"koreksi_id"`
	Status    string `json:"status"` // DISETUJUI / DITOLAK
}

func (h *PerizinanKehadiranHandler) ProcessApproval(c *fiber.Ctx) error {
	var req ApprovalKoreksiRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	koreksi, err := h.repo.GetByID(req.KoreksiID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Data tidak ditemukan"})
	}

	koreksi.Status = req.Status
	if err := h.repo.Update(koreksi); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update status"})
	}

	// Jika DISETUJUI, update data kehadiran (masukkan ID Koreksi)
	// Jika DISETUJUI, update data kehadiran
	if req.Status == "DISETUJUI" {
		kehadiran, err := h.kehadiranRepo.GetByDate(koreksi.ASNID, koreksi.TanggalKehadiran)
		if err == nil {
			// KASUS 1: Data Absen Ada (Misal status TERLAMBAT ingin dikoreksi jadi HADIR)
			kehadiran.PerizinanKehadiranID = &koreksi.ID
			kehadiran.StatusMasuk = "HADIR" // Koreksi dianggap valid, ubah jadi HADIR
			h.kehadiranRepo.Update(kehadiran)
		} else {
			// KASUS 2: Data Absen Tidak Ada (Misal LUPA ABSEN / ALPHA)
			// Kita buatkan data kehadiran baru
			tgl, _ := time.Parse("2006-01-02", koreksi.TanggalKehadiran)
			newKehadiran := model.Kehadiran{
				ASNID:                koreksi.ASNID,
				Tanggal:              koreksi.TanggalKehadiran,
				Tahun:                tgl.Format("2006"),
				Bulan:                tgl.Format("01"),
				StatusMasuk:          "HADIR", // Dianggap Hadir karena koreksi disetujui
				StatusPulang:         "PULANG",
				PerizinanKehadiranID: &koreksi.ID,
			}
			h.kehadiranRepo.Create(newKehadiran)
		}
	}

	return c.JSON(fiber.Map{"message": "Status koreksi berhasil diperbarui"})
}
