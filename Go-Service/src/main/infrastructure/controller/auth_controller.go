package controller

import (
	"Go-Service/src/main/domain/interface/logger"
	"Go-Service/src/main/infrastructure/config"
	"Go-Service/src/main/infrastructure/middleware"
	"Go-Service/src/main/infrastructure/repository"
	"net/http"
	"time"

	"errors"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	Log            logger.Logger
	UserRepository *repository.UserRepository
}

// NewAuthController creates a new AuthController instance
func NewAuthController(log logger.Logger, userRepo *repository.UserRepository) *AuthController {
	return &AuthController{
		Log:            log,
		UserRepository: userRepo,
	}
}

// Login handles the authentication and generation of JWT tokens
func (ac *AuthController) Login(c *gin.Context) {
	var loginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		ac.Log.Error(c, "Invalid login request: "+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid login request"})
		return
	}

	// Find user by username
	user, err := ac.UserRepository.FindByUsername(c, loginRequest.Username)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
		} else {
			ac.Log.Error(c, "Error finding user: "+err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		}
		return
	}

	// Verify password (assuming the password is stored hashed)
	if !verifyPassword(user.Password, loginRequest.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid credentials"})
		return
	}

	// Create a new token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, middleware.Claims{
		UserID: user.ID,
		Role:   user.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(), // Token expires in 1 hour
		},
	})

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWT.SecretKey))
	if err != nil {
		ac.Log.Error(c, "Failed to sign token: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// Helper function to verify password (implement this based on your hashing method)
func verifyPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}