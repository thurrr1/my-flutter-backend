package model

import "gorm.io/gorm"

type Device struct {
	gorm.Model
	UUID           string `json:"uuid" gorm:"unique;not null"` // Token unik generate-an
	Brand          string `json:"brand"`          // Contoh: Samsung, Xiaomi
	Series         string `json:"series"`         // Contoh: Galaxy S21, Redmi Note 10
	FirebaseToken  string `json:"firebase_token"` // Untuk Push Notification
	AdsID          string `json:"ads_id"`         // Untuk tracking iklan
	UserID         uint   `json:"user_id"`         // Relasi ke User
}