package routes

import (
	"my-flutter-backend/internal/handler"
	"my-flutter-backend/internal/middleware"
	"my-flutter-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupJadwalRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewJadwalRepository(db)
	hlRepo := repository.NewHariLiburRepository(db) // Tambah ini
	hdl := handler.NewJadwalHandler(repo, hlRepo)

	api := app.Group("/api/admin", middleware.Auth, middleware.Role("Admin"))

	api.Post("/jadwal", hdl.CreateJadwal)
	api.Post("/jadwal/generate", hdl.GenerateJadwalBulanan)
}
