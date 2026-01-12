package usecase

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"time"
)

var jwtSecret = []byte("rahasia-negara-sangat-kuat")

type UserUsecase struct {
	repo *repository.UserRepository
}

func NewUserUsecase(repo *repository.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (u *UserUsecase) Register(name, nip, password string) error {
	// 1. Hashing Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 2. Simpan ke Database
	user := model.User{
		Name:     name,
		NIP:      nip,
		Password: string(hashedPassword),
	}
	return u.repo.Create(user)
}

func (u *UserUsecase) Login(nip, password string) (string, error) {
	// 1. Cari user berdasarkan NIP
	user, err := u.repo.GetByNIP(nip)
	if err != nil {
		fmt.Println("Bcrypt Error:", err)
		return "", err // User tidak ditemukan
	}

	// 2. Bandingkan Password (Input vs Hash di DB)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		fmt.Println("Bcrypt Error:", err)
		return "", err // Password salah
	}

	// 3. Jika benar, buat Token JWT
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"nip":     user.NIP,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token expired dalam 24 jam
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return t, nil
}
