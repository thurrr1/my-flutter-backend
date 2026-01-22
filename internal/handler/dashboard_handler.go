package handler

import (
	"my-flutter-backend/internal/repository"
	"time"

	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct {
	repo repository.DashboardRepository
}

func NewDashboardHandler(repo repository.DashboardRepository) *DashboardHandler {
	return &DashboardHandler{repo: repo}
}

func (h *DashboardHandler) GetStats(c *fiber.Ctx) error {
	orgID := uint(c.Locals("organisasi_id").(float64))

	now := time.Now()
	date := now.Format("2006-01-02")
	month := int(now.Month())
	year := now.Year()

	stats, err := h.repo.GetDashboardStats(orgID, date, month, year)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data dashboard"})
	}

	return c.JSON(fiber.Map{
		"message": "Berhasil mengambil statistik",
		"data":    stats,
	})
}
