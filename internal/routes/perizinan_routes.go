package routes

import (
	"my-flutter-backend/internal/handler"
	"my-flutter-backend/internal/middleware"
	"my-flutter-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupPerizinanRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewPerizinanRepository(db)
	kehadiranRepo := repository.NewKehadiranRepository(db) // Dibutuhkan untuk Approval
	asnRepo := repository.NewASNRepository(db)             // Dibutuhkan untuk cari NIP Atasan
	jadwalRepo := repository.NewJadwalRepository(db)

	hdl := handler.NewPerizinanHandler(repo, kehadiranRepo, asnRepo, jadwalRepo)

	api := app.Group("/api/perizinan", middleware.Auth)

	// Endpoint untuk Pegawai
	api.Post("/cuti", hdl.AjukanIzin)
	api.Get("/riwayat", hdl.GetRiwayatIzin)
	api.Put("/cuti/:id", hdl.EditIzin)
	api.Delete("/cuti/:id", hdl.DeleteIzin)

	// Endpoint untuk Atasan (Approval)
	api.Get("/bawahan", hdl.GetPengajuanBawahan)
	api.Post("/approval", hdl.ProcessApproval)

	// Endpoint untuk Pembatalan
	api.Post("/cancel/:id", hdl.CancelIzin)
	api.Post("/approve-cancel", hdl.ApproveCancel)
}
