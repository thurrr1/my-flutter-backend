package main

import (
	"my-flutter-backend/config"
	"my-flutter-backend/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	config.ConnectDB()

	app := fiber.New()

	// Middleware Global
	app.Use(cors.New())   // Agar API bisa diakses dari domain/port lain
	app.Use(logger.New()) // Agar log request muncul di terminal (Debugging)

	// Serve Static Files (Agar gambar bisa dibuka via http://localhost:3000/uploads/...)
	app.Static("/uploads", "./uploads")

	routes.SetupASNRoutes(app, config.DB)
	routes.SetupKehadiranRoutes(app, config.DB)
	routes.SetupPerizinanRoutes(app, config.DB)
	routes.SetupJadwalRoutes(app, config.DB)
	routes.SetupPerizinanKehadiranRoutes(app, config.DB)
	routes.SetupBannerRoutes(app, config.DB)
	routes.SetupDashboardRoutes(app, config.DB)
	routes.SetupShiftRoutes(app, config.DB)
	routes.SetupOrganisasiRoutes(app, config.DB)
	routes.SetupHariLiburRoutes(app, config.DB)

	app.Listen(":3000")
}
