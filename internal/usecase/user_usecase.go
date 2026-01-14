package usecase

import (
	"errors"
	"fmt"
	"my-flutter-backend/internal/model"
	"my-flutter-backend/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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

func (u *UserUsecase) Login(nip, password, uuid, brand, series, fcmToken, adsID string) (string, string, error) {
	user, err := u.repo.GetByNIP(nip)
	if err != nil {
		fmt.Println("Bcrypt Error:", err)
		return "", "", err
	}

	// 2. Bandingkan Password (Input vs Hash di DB)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		fmt.Println("Bcrypt Error:", err)
		return "", "", err
	}

	var deviceExists bool
	for _, d := range user.Devices {
		if d.UUID == uuid {
			deviceExists = true
			break
		}
	}

	// 3. Logika Binding
	if !deviceExists {
		// Jika belum ada device terdaftar sama sekali (First Login)
		if len(user.Devices) == 0 {
			newDevice := model.Device{
				UUID:          uuid,
				Brand:         brand,
				Series:        series,
				FirebaseToken: fcmToken,
				AdsID:         adsID,
				UserID:        user.ID,
			}
			// Simpan device baru (nanti buat fungsi di repo)
			u.repo.AddDevice(newDevice)
		} else {
			// Jika sudah ada device lain yang terdaftar
			return "", "", errors.New("Akun terikat di perangkat lain!")
		}
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
		return "", "", err
	}

	return t, user.Name, nil
}
