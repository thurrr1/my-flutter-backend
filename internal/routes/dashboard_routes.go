package routes

import (
	"my-flutter-backend/internal/handler"
	"my-flutter-backend/internal/middleware"
	"my-flutter-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupDashboardRoutes(app *fiber.App, db *gorm.DB) {
	asnRepo := repository.NewASNRepository(db)
	kehadiranRepo := repository.NewKehadiranRepository(db)
	hdl := handler.NewDashboardHandler(asnRepo, kehadiranRepo)

	api := app.Group("/api/admin/dashboard", middleware.Auth, middleware.Role("Admin"))
	api.Get("/", hdl.GetStats)
}
