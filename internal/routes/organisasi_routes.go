package routes

import (
	"my-flutter-backend/internal/handler"
	"my-flutter-backend/internal/middleware"
	"my-flutter-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func SetupOrganisasiRoutes(app *fiber.App, db *gorm.DB) {
	repo := repository.NewOrganisasiRepository(db)
	hdl := handler.NewOrganisasiHandler(repo)

	api := app.Group("/api/admin/organisasi", middleware.Auth, middleware.Role("Admin"))

	api.Get("/", hdl.GetInfo)
	api.Put("/", hdl.UpdateInfo)
	api.Post("/lokasi", hdl.AddLokasi)          // Tambah Lokasi Baru
	api.Put("/lokasi/:id", hdl.UpdateLokasi)    // Update Lokasi
	api.Delete("/lokasi/:id", hdl.DeleteLokasi) // Hapus Lokasi
}
