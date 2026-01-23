package routes

import (
	"my-flutter-backend/internal/handler"
	"my-flutter-backend/internal/middleware"
	"my-flutter-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupASNRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewASNRepository(db)
	hdl := handler.NewASNHandler(repo)

	// Auth Routes
	app.Post("/api/login", hdl.Login)
	app.Post("/api/refresh-token", hdl.RefreshToken)

	// Profile Routes (Protected)
	api := app.Group("/api/asn", middleware.Auth)
	api.Get("/profile", hdl.GetProfile)
	api.Put("/profile", hdl.UpdateProfile)
	api.Put("/password", hdl.ChangePassword)
	api.Get("/atasan-list", hdl.GetListAtasan) // Get List Kandidat Atasan
	api.Post("/atasan", hdl.UpdateAtasan)      // Update Atasan Saya

	// Admin Routes (Kelola Pegawai)
	admin := app.Group("/api/admin/asn", middleware.Auth, middleware.Role("Admin"))
	admin.Get("/", hdl.GetAll)
	admin.Get("/:id", hdl.GetASNDetail) // Route baru untuk detail
	admin.Post("/", hdl.CreateASN)
	admin.Put("/:id", hdl.UpdateASN)
	admin.Delete("/:id", hdl.DeleteASN)
	admin.Delete("/:id/device", hdl.ResetDevice)
}
