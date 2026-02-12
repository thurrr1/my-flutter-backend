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
	asnRepo := repository.NewASNRepository(db)
	hdl := handler.NewOrganisasiHandler(repo, asnRepo)

	// Allow Admin and Super Admin
	api := app.Group("/api/admin/organisasi", middleware.Auth, middleware.Permission("edit_jadwal"))

	api.Get("/", hdl.GetInfo)
	api.Put("/", hdl.UpdateOrganisasi)          // Update Info Organisasi (Nama & Email)
	api.Post("/lokasi", hdl.AddLokasi)          // Tambah Lokasi Baru
	api.Put("/lokasi/:id", hdl.UpdateLokasi)    // Update Lokasi
	api.Delete("/lokasi/:id", hdl.DeleteLokasi) // Hapus Lokasi

	// Super Admin Only Routes
	// Super Admin Only Routes
	api.Post("/", middleware.Role("Super Admin"), hdl.CreateOrganisasi)       // Buat Organisasi Baru
	api.Get("/all", middleware.Role("Super Admin"), hdl.GetAllOrganisasi)     // List Semua Organisasi
	api.Put("/:id", middleware.Role("Super Admin"), hdl.UpdateOrganisasiByID) // Update Org Lain
	api.Get("/:id/admins", middleware.Role("Super Admin"), hdl.GetAdmins)     // Get List Admin per Org
}
