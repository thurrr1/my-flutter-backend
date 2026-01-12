package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `json:"name"`
	NIP      string `json:"nip" gorm:"column:nip;unique;not null"`
	Password string `json:"password"`
}
