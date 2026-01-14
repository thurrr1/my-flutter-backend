package http

import (
	"my-flutter-backend/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	usecase *usecase.UserUsecase
}

func NewUserHandler(u *usecase.UserUsecase) *UserHandler {
	return &UserHandler{usecase: u}
}

func (h *UserHandler) Register(c *fiber.Ctx) error {
	var input struct {
		Name     string `json:"name"`
		NIP      string `json:"nip"` // Ganti jadi NIP
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Input salah"})
	}

	// Panggil usecase dengan NIP
	err := h.usecase.Register(input.Name, input.NIP, input.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal registrasi: " + err.Error()})
	}

	return c.JSON(fiber.Map{"message": "User dengan NIP berhasil terdaftar!"})
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	var input struct {
		NIP           string `json:"nip"`
		Password      string `json:"password"`
		UUID          string `json:"uuid"`           // Dari Flutter
		Brand         string `json:"brand"`          // Dari Flutter
		Series        string `json:"series"`         // Dari Flutter
		FirebaseToken string `json:"firebase_token"` // Dari Flutter
		AdsID         string `json:"ads_id"`         // Dari Flutter
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Input tidak valid"})
	}

	// Masukkan semua parameter ke usecase
	token, name, err := h.usecase.Login(
		input.NIP,
		input.Password,
		input.UUID,
		input.Brand,
		input.Series,
		input.FirebaseToken,
		input.AdsID,
	)

	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Login Berhasil!",
		"token":   token,
		"name":    name,
	})
}
