package routes

import (
	"my-flutter-backend/internal/handler"
	"my-flutter-backend/internal/middleware"
	"my-flutter-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupKehadiranRoutes(app *fiber.App, db *gorm.DB) {
	asnRepo := repository.NewASNRepository(db)
	kehadiranRepo := repository.NewKehadiranRepository(db)
	jadwalRepo := repository.NewJadwalRepository(db) // Tambah ini
	hdl := handler.NewKehadiranHandler(kehadiranRepo, asnRepo, jadwalRepo)

	// Grouping route khusus kehadiran
	api := app.Group("/api/kehadiran", middleware.Auth)

	api.Post("/checkin", hdl.CheckIn)
	api.Post("/checkout", hdl.CheckOut)
	api.Get("/riwayat", hdl.GetHistory)
	api.Get("/rekap", hdl.GetRekap)
}
