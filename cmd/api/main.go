package main

import (
	"my-flutter-backend/config"
	"my-flutter-backend/internal/delivery/http"
	"my-flutter-backend/internal/repository"
	"my-flutter-backend/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// 1. Koneksi Database
	config.ConnectDB()

	// 2. Setup Layer Clean Architecture
	// Kita buat dari yang paling dalam (Repository -> Usecase -> Handler)
	userRepo := repository.NewUserRepository(config.DB)
	userUsecase := usecase.NewUserUsecase(userRepo)
	userHandler := http.NewUserHandler(userUsecase)

	// 3. Inisialisasi Fiber
	app := fiber.New()
	app.Use(cors.New())

	// 4. Routing
	// Grouping route agar rapi
	api := app.Group("/api")

	// Route untuk Register
	api.Post("/register", userHandler.Register)
	api.Post("/login", userHandler.Login)

	// Test Route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Server Go Clean Arch Jalan!"})
	})

	app.Listen(":3000")
}
