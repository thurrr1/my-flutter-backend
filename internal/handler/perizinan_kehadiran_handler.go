package handler

import (
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"

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

type AjukanKoreksiRequest struct {
	TanggalKehadiran string `json:"tanggal_kehadiran"`
	TipeKoreksi      string `json:"tipe_koreksi"` // TELAT, PULANG_CEPAT, LUAR_RADIUS
	Alasan           string `json:"alasan"`
}

func (h *PerizinanKehadiranHandler) AjukanKoreksi(c *fiber.Ctx) error {
	asnID := uint(c.Locals("user_id").(float64))

	var req AjukanKoreksiRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
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
		TanggalKehadiran: req.TanggalKehadiran,
		TipeKoreksi:      req.TipeKoreksi,
		Alasan:           req.Alasan,
		Status:           "MENUNGGU",
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
			kehadiran.PerizinanKehadiranID = &koreksi.ID
			h.kehadiranRepo.Update(kehadiran)
		}
	}

	return c.JSON(fiber.Map{"message": "Status koreksi berhasil diperbarui"})
}
