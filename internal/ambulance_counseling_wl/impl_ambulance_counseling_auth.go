package ambulance_counseling_wl

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/AKoricansky/wac-be-xkoricansky/internal/db_service"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type implAmbulanceCounselingAuthAPI struct {
	userDbService db_service.DbService[User]
}

func NewAmbulanceCounselingAuthApi(userDbService db_service.DbService[User]) AmbulanceCounselingAuthAPI {
	return &implAmbulanceCounselingAuthAPI{
		userDbService: userDbService,
	}
}

// for user creation
func generateRandomID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// password hashing
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// hash check
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (o *implAmbulanceCounselingAuthAPI) UserLogin(c *gin.Context) {
	var loginForm LoginForm
	if err := c.ShouldBindJSON(&loginForm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login data"})
		return
	}

	email := strings.ToLower(loginForm.Email)

	ctx := context.Background()
	users, err := o.userDbService.FindDocumentsByField(ctx, "email", email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if len(users) == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	user := users[0]

	if !checkPasswordHash(loginForm.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	tokenString, err := GenerateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := UserLogin200Response{
		Token: tokenString,
	}

	c.JSON(http.StatusOK, gin.H{
		"token": response.Token,
		"user":  user,
	})
}

func (o *implAmbulanceCounselingAuthAPI) UserRegister(c *gin.Context) {
	var registrationForm RegistrationForm
	if err := c.ShouldBindJSON(&registrationForm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid registration data"})
		return
	}

	email := strings.ToLower(registrationForm.Email)

	ctx := context.Background()

	existingUsers, err := o.userDbService.FindDocumentsByField(ctx, "email", email)
	if err != nil && err != db_service.ErrNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if len(existingUsers) == 0 {
		existingUsers, err = o.userDbService.FindDocumentsByField(ctx, "email", email)
		if err != nil && err != db_service.ErrNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
	}

	if len(existingUsers) > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	id, err := generateRandomID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate user ID"})
		return
	}

	hashedPassword, err := hashPassword(registrationForm.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	user := User{
		Id:           id,
		Name:         registrationForm.Name,
		Email:        email,
		Type:         "patient",
		PasswordHash: hashedPassword,
	}

	err = o.userDbService.CreateDocument(ctx, user.Id, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusCreated, user)
}
