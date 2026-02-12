package routes

import (
	"my-flutter-backend/internal/handler"
	"my-flutter-backend/internal/middleware"
	"my-flutter-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupDashboardRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewDashboardRepository(db)
	hdl := handler.NewDashboardHandler(repo)

	api := app.Group("/api/admin/dashboard", middleware.Auth, middleware.Permission("edit_jadwal"))
	api.Get("/", hdl.GetStats)
}
