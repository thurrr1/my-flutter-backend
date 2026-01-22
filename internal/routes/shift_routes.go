package routes

import (
	"my-flutter-backend/internal/handler"
	"my-flutter-backend/internal/middleware"
	"my-flutter-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupShiftRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewShiftRepository(db)
	jadwalRepo := repository.NewJadwalRepository(db) // Tambah ini
	hdl := handler.NewShiftHandler(repo, jadwalRepo)

	api := app.Group("/api/admin/shift", middleware.Auth, middleware.Role("Admin"))
	api.Get("/", hdl.GetAll)
	api.Post("/", hdl.Create)
	api.Put("/:id", hdl.Update)
	api.Delete("/:id", hdl.Delete)
}
