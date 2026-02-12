package main

import (
	"fmt"
	"log"
	"my-flutter-backend/config"
	"my-flutter-backend/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🌱 Memulai Database Seeding...")

	// Load .env manual karena ini script terpisah
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: File .env tidak ditemukan, menggunakan environment variables sistem.")
	}

	// Koneksi DB (Tanpa Auto Migrate & Seed Otomatis, kita panggil manual)
	// Kita perlu modif config.ConnectDB sedikit agar fleksibel atau kita copy logic koneksinya disini
	// Tapi lebih baik kita panggil config.ConnectDB() lalu matikan auto-seed di sana
	config.ConnectDB()

	// Jalankan Seeder
	fmt.Println("🚀 Menjalankan SeedAll...")
	database.SeedAll(config.DB)

	fmt.Println("✅ Seeding Selesai!")
}
