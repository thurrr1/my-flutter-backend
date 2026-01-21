package model

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	NamaRole    string       `json:"nama_role" gorm:"unique;not null"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
	ASN         []ASN        `json:"asn"`
}

type Permission struct {
	gorm.Model
	NamaPermission string `json:"nama_permission" gorm:"unique;not null"`
}