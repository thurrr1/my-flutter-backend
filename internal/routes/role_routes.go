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

	api := app.Group("/api/admin/roles", middleware.Auth, middleware.Role("Admin"))
	api.Get("/", hdl.GetAll)
	api.Get("/permissions", hdl.GetAllPermissions) // List semua permission yang tersedia
	api.Get("/:id", hdl.GetDetail)
	api.Post("/", hdl.Create)
	api.Put("/:id", hdl.Update)
	api.Delete("/:id", hdl.Delete)
}
