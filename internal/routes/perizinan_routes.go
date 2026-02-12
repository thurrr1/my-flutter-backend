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
	// Hanya yang punya permission 'approve_cuti' yang bisa akses
	approval := api.Group("/", middleware.Permission("approve_cuti"))
	approval.Get("/bawahan", hdl.GetPengajuanBawahan)
	approval.Post("/approval", hdl.ProcessApproval)
	approval.Post("/approve-cancel", hdl.ApproveCancel)

	// Endpoint untuk Pembatalan
	api.Post("/cancel/:id", hdl.CancelIzin)

}
