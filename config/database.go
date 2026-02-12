package config

import (
	"fmt"
	"my-flutter-backend/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// Format: user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	// Jika pakai XAMPP default, user adalah "root" dan password kosong ""
	dsn := "root:@tcp(127.0.0.1:3306)/my_flutter_db?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Gagal koneksi ke database!")
	}

	fmt.Println("Koneksi Database Berhasil!")

	// Auto Migration: Membuat tabel otomatis berdasarkan struct di folder model
	db.AutoMigrate(
		&model.Organisasi{}, &model.Lokasi{}, &model.Role{}, &model.Permission{},
		&model.ASN{}, &model.Kehadiran{}, &model.PerizinanCuti{},
		&model.PerizinanKehadiran{}, &model.Jadwal{}, &model.Shift{}, &model.HariLibur{},
		&model.Device{}, &model.Banner{},
	)
	// database.SeedAll(db) // Dipindahkan ke cmd/seeder/main.go

	DB = db
}
