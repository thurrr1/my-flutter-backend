package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func Auth(c *fiber.Ctx) error {
	// 1. Ambil token dari Header Authorization
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token tidak ditemukan"})
	}

	// Format header biasanya: "Bearer <token>"
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

	// 2. Parse dan Validasi Token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		// Pastikan key ini SAMA PERSIS dengan yang ada di asn_handler.go
		return []byte("rahasia_negara"), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token tidak valid atau kadaluwarsa"})
	}

	// 3. Simpan data user (Claims) ke Context agar bisa dipakai di Handler
	claims := token.Claims.(jwt.MapClaims)
	c.Locals("user_id", claims["user_id"])
	c.Locals("nip", claims["nip"])
	c.Locals("role", claims["role"]) // Simpan Role ke Context
	c.Locals("organisasi_id", claims["organisasi_id"])

	return c.Next()
}
