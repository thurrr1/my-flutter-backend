package routes

import (
	"my-flutter-backend/internal/handler"
	"my-flutter-backend/internal/middleware"
	"my-flutter-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupPerizinanKehadiranRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewPerizinanKehadiranRepository(db)
	kehadiranRepo := repository.NewKehadiranRepository(db)
	asnRepo := repository.NewASNRepository(db)
	hdl := handler.NewPerizinanKehadiranHandler(repo, kehadiranRepo, asnRepo)

	api := app.Group("/api/koreksi", middleware.Auth)

	api.Post("/ajukan", hdl.AjukanKoreksi)
	api.Get("/riwayat", hdl.GetRiwayat)
	api.Put("/ajukan/:id", hdl.EditKoreksi)
	api.Delete("/ajukan/:id", hdl.DeleteKoreksi)

	// Approval Routes
	approval := api.Group("/", middleware.Permission("approve_cuti"))
	approval.Get("/bawahan", hdl.GetBawahan)
	approval.Post("/approval", hdl.ProcessApproval)
}
