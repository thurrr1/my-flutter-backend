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
	app.Post("/api/web-login", hdl.WebLogin) // Endpoint Login Khusus Web
	app.Post("/api/refresh-token", hdl.RefreshToken)

	// Forgot Password Routes (Mobile Flow)
	app.Post("/api/forgot-password/request", hdl.RequestOTP)       // 1. Cek NIP & Email -> Kirim OTP
	app.Post("/api/forgot-password/verify", hdl.VerifyOTP)         // 2. Cek OTP -> Lanjut ke halaman password baru
	app.Post("/api/forgot-password/reset", hdl.ResetPasswordFinal) // 3. Submit Password Baru

	// Profile Routes (Protected)
	api := app.Group("/api/asn", middleware.Auth)
	api.Get("/profile", hdl.GetProfile)
	api.Put("/profile", hdl.UpdateProfile)
	api.Put("/password", hdl.ChangePassword)
	api.Get("/atasan-list", hdl.GetListAtasan)      // Get List Kandidat Atasan
	api.Post("/atasan", hdl.UpdateAtasan)           // Update Atasan Saya
	api.Get("/bawahan", hdl.GetSubordinates)        // Get List Bawahan (Untuk Atasan)
	api.Post("/upload-foto", hdl.UploadFotoProfile) // Upload Foto Profile

	// Admin Routes (Kelola Pegawai)
	// admin := app.Group("/api/admin/asn", middleware.Auth, middleware.Role("Admin", "Super Admin"))
	// Ganti jadi Permission Based:
	admin := app.Group("/api/admin/asn", middleware.Auth, middleware.Permission("edit_jadwal"))
	admin.Get("/", hdl.GetAll)
	admin.Get("/:id", hdl.GetASNDetail) // Route baru untuk detail
	admin.Post("/", hdl.CreateASN)
	admin.Post("/import", hdl.ImportASN) // Route Import Excel
	admin.Put("/:id", hdl.UpdateASN)
	admin.Patch("/:id/status", hdl.ToggleActiveASN)         // Toggle Status Aktif/Nonaktif
	admin.Put("/:id/reset-password", hdl.ResetUserPassword) // Route reset password (Lupa Password)
	admin.Delete("/:id", hdl.DeleteASN)
	admin.Delete("/:id/device", hdl.ResetDevice)

	// Public Routes (Image Serving)
	app.Get("/api/public/asn/:id/foto", hdl.GetFotoProfile)
}
