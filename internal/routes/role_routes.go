package routes

import (
	"my-flutter-backend/internal/handler"
	"my-flutter-backend/internal/middleware"
	"my-flutter-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupRoleRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewRoleRepository(db)
	hdl := handler.NewRoleHandler(repo)

	app.Get("/api/admin/roles", middleware.Auth, middleware.Role("Admin"), hdl.GetAll)
}
