package routes

import (
	"my-flutter-backend/internal/handler"
	"my-flutter-backend/internal/middleware"
	"my-flutter-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupHariLiburRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewHariLiburRepository(db)
	hdl := handler.NewHariLiburHandler(repo)

	api := app.Group("/api/admin/hari-libur", middleware.Auth, middleware.Permission("edit_jadwal"))

	api.Get("/", hdl.GetAll)
	api.Post("/", hdl.Create)
	api.Put("/:id", hdl.Update)
	api.Delete("/:id", hdl.Delete)
}
