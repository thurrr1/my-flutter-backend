package model

import "gorm.io/gorm"

type Organisasi struct {
	gorm.Model
	NamaOrganisasi string   `json:"nama_organisasi" gorm:"not null"`
	ASN            []ASN    `json:"asn"`
	Lokasi         []Lokasi `json:"lokasi"`
}

type Lokasi struct {
	gorm.Model
	OrganisasiID uint    `json:"organisasi_id"`
	NamaLokasi   string  `json:"nama_lokasi"`
	Alamat       string  `json:"alamat"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	RadiusMeter  float64 `json:"radius_meter"`
}