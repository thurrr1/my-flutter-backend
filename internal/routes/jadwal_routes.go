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
	hlRepo := repository.NewHariLiburRepository(db)
	kehadiranRepo := repository.NewKehadiranRepository(db)
	shiftRepo := repository.NewShiftRepository(db)
	asnRepo := repository.NewASNRepository(db) // Tambah ini
	hdl := handler.NewJadwalHandler(repo, hlRepo, kehadiranRepo, shiftRepo, asnRepo)

	// Mobile Routes
	app.Get("/api/jadwal/saya", middleware.Auth, hdl.GetJadwalSaya)

	api := app.Group("/api/admin", middleware.Auth, middleware.Role("Admin", "Super Admin"))
	api.Get("/jadwal", hdl.GetJadwalHarian)                   // Lihat per tanggal
	api.Get("/jadwal/dashboard-stats", hdl.GetDashboardStats) // PENTING: Taruh ini SEBELUM :id
	api.Get("/jadwal/:id", hdl.GetJadwalDetail)               // Detail untuk Edit
	api.Post("/jadwal", hdl.CreateJadwal)                     // Buat manual satu
	api.Post("/jadwal/import", hdl.ImportJadwal)              // Import Excel
	api.Post("/jadwal/generate", hdl.GenerateJadwalBulanan)
	api.Post("/jadwal/generate-daily", hdl.GenerateJadwalHarian) // Bulk Harian
	api.Put("/jadwal/:id", hdl.UpdateJadwal)                     // Edit Shift
	api.Delete("/jadwal/:id", hdl.DeleteJadwal)                  // Hapus
	api.Delete("/jadwal/date/bulk", hdl.DeleteJadwalByDate)      // Hapus Massal per Tanggal
}
