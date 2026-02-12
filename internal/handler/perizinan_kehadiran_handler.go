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
	isLokasi := c.FormValue("is_lokasi") == "true"

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

		IsLokasi: isLokasi,
		Alasan:   alasan,
		Status:   "MENUNGGU",
		PathFile: pathFile,
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
	if req.Status == "DISETUJUI" {
		kehadiran, err := h.kehadiranRepo.GetByDate(koreksi.ASNID, koreksi.TanggalKehadiran)
		if err == nil {
			// Hanya update data kehadiran yang sudah ada (Inject ID Izin)
			if koreksi.IsLokasi {
				kehadiran.PerizinanLokasiID = &koreksi.ID
			} else {
				kehadiran.PerizinanKehadiranID = &koreksi.ID
			}
			h.kehadiranRepo.Update(kehadiran)
		}
	}

	return c.JSON(fiber.Map{"message": "Status koreksi berhasil diperbarui"})
}

func (h *PerizinanKehadiranHandler) EditKoreksi(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	asnID := uint(c.Locals("user_id").(float64))

	// Cek Data Lama
	koreksi, err := h.repo.GetByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Data koreksi tidak ditemukan"})
	}

	// Validasi Pemilik
	if koreksi.ASNID != asnID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Anda tidak berhak mengubah data ini"})
	}

	// Validasi Status
	if koreksi.Status != "MENUNGGU" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak dapat diubah karena status sudah " + koreksi.Status})
	}

	// Parse Form Data
	tanggal := c.FormValue("tanggal_kehadiran")
	tipe := c.FormValue("tipe_koreksi")
	alasan := c.FormValue("alasan")
	isLokasi := c.FormValue("is_lokasi") == "true"

	if tanggal == "" || tipe == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Tanggal dan Tipe Koreksi wajib diisi"})
	}

	// Handle File Upload Baru (Jika Ada)
	file, errFile := c.FormFile("file_bukti")
	if errFile == nil {
		// Hapus file lama jika ada
		if koreksi.PathFile != "" {
			os.Remove(koreksi.PathFile)
		}

		// Upload file baru
		uploadDir := "./uploads/perizinan_kehadiran"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.MkdirAll(uploadDir, 0755)
		}
		filename := fmt.Sprintf("%d_%d_%s", asnID, time.Now().Unix(), filepath.Base(file.Filename))
		pathFile := fmt.Sprintf("uploads/perizinan_kehadiran/%s", filename)
		c.SaveFile(file, pathFile)

		koreksi.PathFile = pathFile
	}

	// Update Fields
	koreksi.TanggalKehadiran = tanggal
	koreksi.TipeKoreksi = tipe
	koreksi.Alasan = alasan
	koreksi.IsLokasi = isLokasi

	if err := h.repo.Update(koreksi); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal update data koreksi"})
	}

	return c.JSON(fiber.Map{
		"message": "Data koreksi berhasil diperbarui",
		"data":    koreksi,
	})
}

func (h *PerizinanKehadiranHandler) DeleteKoreksi(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	asnID := uint(c.Locals("user_id").(float64))

	// Cek Data
	koreksi, err := h.repo.GetByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Data koreksi tidak ditemukan"})
	}

	// Validasi Pemilik
	if koreksi.ASNID != asnID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Anda tidak berhak menghapus data ini"})
	}

	// Validasi Status
	if koreksi.Status != "MENUNGGU" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak dapat dihapus karena status sudah " + koreksi.Status})
	}

	// Hapus File Bukti
	if koreksi.PathFile != "" {
		os.Remove(koreksi.PathFile)
	}

	// Hapus dari DB
	if err := h.repo.Delete(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menghapus data koreksi"})
	}

	return c.JSON(fiber.Map{"message": "Data koreksi berhasil dihapus"})
}
