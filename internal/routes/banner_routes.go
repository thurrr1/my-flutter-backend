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

	// Public / User Route
	app.Get("/api/banner", middleware.Auth, hdl.GetAll)

	// Admin Route
	admin := app.Group("/api/admin/banner", middleware.Auth, middleware.Role("Admin"))
	admin.Post("/", hdl.Create)
	admin.Delete("/:id", hdl.Delete)
}
