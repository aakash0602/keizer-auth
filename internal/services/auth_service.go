package services

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"keizer-auth/internal/models"
	"keizer-auth/internal/repositories"
	"keizer-auth/internal/utils"
	"keizer-auth/internal/validators"
)

type AuthService struct {
	userRepo  *repositories.UserRepository
	redisRepo *repositories.RedisRepository
}

func NewAuthService(userRepo *repositories.UserRepository, redisRepo *repositories.RedisRepository) *AuthService {
	return &AuthService{userRepo: userRepo, redisRepo: redisRepo}
}

func (as *AuthService) RegisterUser(userRegister *validators.SignUpUser) error {
	passwordHash, err := utils.HashPassword(userRegister.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	err = as.redisRepo.Set("registration-verification-otp-"+userRegister.Email, otp, time.Minute)
	if err != nil {
		return fmt.Errorf("failed to save otp in redis: %w", err)
	}

	// TODO: email should be sent using async func
	if err = SendOTPEmail(userRegister.Email, otp); err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	if err = as.userRepo.CreateUser(&models.User{
		Email:        userRegister.Email,
		FirstName:    userRegister.FirstName,
		LastName:     userRegister.LastName,
		PasswordHash: passwordHash,
	}); err != nil {
		return err
	}

	return nil
}

func (as *AuthService) VerifyPassword(email string, password string) (bool, *models.User, error) {
	user, err := as.userRepo.GetUserByStruct(&models.User{Email: email})
	if err != nil {
		return false, nil, err
	}
	if user == nil {
		return false, nil, err
	}

	isValid, err := utils.VerifyPassword(password, user.PasswordHash)
	if err != nil {
		return false, nil, err
	}
	if !isValid {
		return false, nil, nil
	}

	return true, user, nil
}

func (as *AuthService) VerifyOTP(verifyOtpBody *validators.VerifyOTP) (bool, error) {
	val, err := as.redisRepo.Get("registration-verification-otp-" + verifyOtpBody.Email)
	if err != nil {
		if err == redis.Nil {
			return false, fmt.Errorf("otp expired")
		}
		return false, fmt.Errorf("failed to get otp from redis %w", err)
	}

	if val != verifyOtpBody.Otp {
		return false, nil
	}
	return true, nil
}

func (as *AuthService) SetIsVerified(id string) error {
	return as.userRepo.UpdateUser(id, &models.User{IsVerified: true})
}
