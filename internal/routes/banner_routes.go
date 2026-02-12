package routes

import (
	"my-flutter-backend/internal/handler"
	"my-flutter-backend/internal/middleware"
	"my-flutter-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupBannerRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewBannerRepository(db)
	hdl := handler.NewBannerHandler(repo)

	app.Get("/api/banner", middleware.Auth, hdl.GetAll) // Mobile (Active Only)

	admin := app.Group("/api/admin/banner", middleware.Auth, middleware.Permission("edit_jadwal"))
	admin.Get("/", hdl.GetAllAdmin) // Admin (All)
	admin.Post("/", hdl.Create)
	admin.Put("/:id/toggle", hdl.ToggleStatus) // Toggle Active/Inactive
}
