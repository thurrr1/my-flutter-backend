package routes

import (
	"my-flutter-backend/internal/handler"
	"my-flutter-backend/internal/middleware"
	"my-flutter-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupReportRoutes(app *fiber.App, db *gorm.DB) {
	jadwalRepo := repository.NewJadwalRepository(db)
	kehadiranRepo := repository.NewKehadiranRepository(db)
	asnRepo := repository.NewASNRepository(db)

	hdl := handler.NewReportHandler(jadwalRepo, kehadiranRepo, asnRepo)

	api := app.Group("/api/admin/reports", middleware.Auth, middleware.Role("Admin"))
	api.Get("/monthly", hdl.GetMonthlyRecap)
	api.Get("/daily", hdl.GetDailyRecap)
}
