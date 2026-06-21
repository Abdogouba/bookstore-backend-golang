package tests

import (
	"bookstore-backend/config"
	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/models"
	"bookstore-backend/internal/repositories"
	"bookstore-backend/internal/seeder"
	"bookstore-backend/internal/utils"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestRegisterSuccess(t *testing.T) {

	cleanDatabase()

	payload := map[string]interface{}{
		"name":         "Ahmed",
		"email":        "ahmed@test.com",
		"password":     "password123",
		"phone_number": "01012345678",
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(
		http.MethodPost,
		"/auth/register",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	// -------------------------
	// 1. STATUS CODE CHECK
	// -------------------------
	assert.Equal(t, http.StatusCreated, res.Code)

	// -------------------------
	// 2. RESPONSE DTO CHECK
	// -------------------------
	var response map[string]interface{}
	_ = json.Unmarshal(res.Body.Bytes(), &response)

	assert.Equal(t, "Ahmed", response["name"])
	assert.Equal(t, "ahmed@test.com", response["email"])
	assert.Equal(t, "01012345678", response["phone_number"])
	assert.Equal(t, "user", response["role"])

	// ID should exist (not zero / nil)
	assert.NotNil(t, response["id"])

	// -------------------------
	// 3. DATABASE CHECK
	// -------------------------
	var user models.User
	err := testDB.
		Where("email = ?", "ahmed@test.com").
		First(&user).
		Error

	assert.Nil(t, err)

	// All model fields
	assert.Equal(t, "Ahmed", user.Name)
	assert.Equal(t, "ahmed@test.com", user.Email)
	assert.Equal(t, "01012345678", user.PhoneNumber)
	assert.Equal(t, "user", user.Role)

	// timestamps (basic sanity checks)
	assert.False(t, user.CreatedAt.IsZero())
	assert.False(t, user.UpdatedAt.IsZero())

	// ID should be generated
	assert.NotZero(t, user.ID)

	// -------------------------
	// 4. SECURITY CHECK
	// -------------------------
	assert.NotEqual(t, "password123", user.PasswordHash)
	assert.NotEmpty(t, user.PasswordHash)

	// -------------------------
	// 5. CONSISTENCY CHECK
	// -------------------------
	assert.Equal(t, response["email"], user.Email)
	assert.Equal(t, response["role"], user.Role)
}

func TestRegisterDuplicateEmail(t *testing.T) {

	cleanDatabase()

	// -------------------------
	// Seed existing user directly in DB
	// -------------------------
	repo := repositories.NewUserRepository(testDB)

	err := repo.Create(&models.User{
		Name:         "Existing User",
		Email:        "duplicate@test.com",
		PasswordHash: "hashedpassword",
		Role:         "user",
	})

	assert.Nil(t, err)

	// -------------------------
	// API request (duplicate email)
	// -------------------------
	payload := map[string]interface{}{
		"name":     "New User",
		"email":    "duplicate@test.com",
		"password": "password123",
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(
		http.MethodPost,
		"/auth/register",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	// -------------------------
	// 1. STATUS CHECK
	// -------------------------
	assert.Equal(t, http.StatusBadRequest, res.Code)

	// -------------------------
	// 2. DB CHECK (still only 1 user)
	// -------------------------
	var count int64
	testDB.Model(&models.User{}).
		Where("email = ?", "duplicate@test.com").
		Count(&count)

	assert.Equal(t, int64(1), count)

	// -------------------------
	// 3. RESPONSE CHECK (optional but useful)
	// -------------------------
	var response map[string]interface{}
	_ = json.Unmarshal(res.Body.Bytes(), &response)

	assert.NotNil(t, response) // ensures body exists
}

func TestRegisterValidationError(t *testing.T) {

	cleanDatabase()

	// -------------------------
	// Missing required fields (email + password)
	// -------------------------
	payload := map[string]interface{}{
		"name": "Ahmed",
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(
		http.MethodPost,
		"/auth/register",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	// -------------------------
	// 1. STATUS CHECK
	// -------------------------
	assert.Equal(t, http.StatusBadRequest, res.Code)

	// -------------------------
	// 2. DB MUST BE EMPTY
	// -------------------------
	var count int64
	testDB.Model(&models.User{}).Count(&count)

	assert.Equal(t, int64(0), count)

	// -------------------------
	// 3. RESPONSE CHECK
	// -------------------------
	var response map[string]interface{}
	_ = json.Unmarshal(res.Body.Bytes(), &response)

	assert.NotNil(t, response)
}

func TestSeedAdminCreatesAdmin(t *testing.T) {

	cleanDatabase()

	err := seeder.SeedAdmin(testDB)
	assert.Nil(t, err)

	var admin models.User

	err = testDB.
		Where("role = ?", "admin").
		First(&admin).Error

	assert.Nil(t, err)

	// -------------------------
	// BASIC FIELD CHECKS
	// -------------------------
	assert.Equal(t, "admin", admin.Role)
	assert.Equal(t, config.AppConfig.AdminEmail, admin.Email)
	assert.NotEmpty(t, admin.PasswordHash)

	// -------------------------
	// PASSWORD HASH VALIDATION (IMPORTANT CHANGE)
	// -------------------------
	err = utils.CheckPassword(
		config.AppConfig.AdminPassword,
		admin.PasswordHash,
	)

	assert.Nil(t, err)
}

func TestSeedAdminIsIdempotent(t *testing.T) {

	cleanDatabase()

	// first run
	err := seeder.SeedAdmin(testDB)
	assert.Nil(t, err)

	// second run
	err = seeder.SeedAdmin(testDB)
	assert.Nil(t, err)

	// count admins
	var count int64
	testDB.Model(&models.User{}).
		Where("role = ?", "admin").
		Count(&count)

	assert.Equal(t, int64(1), count)
}

func TestLoginSuccess(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	reqBody := map[string]string{
		"email":    "ahmed@test.com",
		"password": "password123",
	}

	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	// -------------------------
	// STATUS CODE
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	// -------------------------
	// RESPONSE BODY
	// -------------------------

	var response dto.LoginResponse

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(t, user.ID, response.User.ID)
	assert.Equal(t, user.Name, response.User.Name)
	assert.Equal(t, user.Email, response.User.Email)
	assert.Equal(t, user.PhoneNumber, response.User.PhoneNumber)
	assert.Equal(t, user.Role, response.User.Role)

	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)

	// -------------------------
	// DATABASE CHANGES
	// -------------------------

	var refreshToken models.RefreshToken

	err = testDB.
		Where("user_id = ?", user.ID).
		First(&refreshToken).
		Error

	assert.NoError(t, err)

	assert.Equal(
		t,
		user.ID,
		refreshToken.UserID,
	)

	assert.Equal(
		t,
		response.RefreshToken,
		refreshToken.Token,
	)

	assert.False(
		t,
		refreshToken.ExpiresAt.IsZero(),
	)
}

func TestLoginEmailNotFound(t *testing.T) {

	cleanDatabase()

	reqBody := map[string]string{
		"email":    "missing@test.com",
		"password": "password123",
	}

	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(
		t,
		http.StatusBadRequest,
		recorder.Code,
	)

	var response map[string]string

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		"invalid credentials",
		response["error"],
	)

	var count int64

	testDB.
		Model(&models.RefreshToken{}).
		Count(&count)

	assert.Equal(
		t,
		int64(0),
		count,
	)
}

func TestLoginValidation(t *testing.T) {

	cleanDatabase()

	reqBody := map[string]string{
		"email":    "",
		"password": "",
	}

	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(
		t,
		http.StatusBadRequest,
		recorder.Code,
	)

	var response map[string]string

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Contains(
		t,
		response["error"],
		"Email",
	)

	var count int64

	testDB.
		Model(&models.RefreshToken{}).
		Count(&count)

	assert.Equal(
		t,
		int64(0),
		count,
	)
}

func TestLoginWrongPassword(t *testing.T) {

	cleanDatabase()

	createTestUser(t)

	reqBody := map[string]string{
		"email":    "ahmed@test.com",
		"password": "wrong-password",
	}

	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(
		t,
		http.StatusBadRequest,
		recorder.Code,
	)

	var response map[string]string

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		"invalid credentials",
		response["error"],
	)

	var count int64

	testDB.
		Model(&models.RefreshToken{}).
		Count(&count)

	assert.Equal(
		t,
		int64(0),
		count,
	)
}

func TestLoginReplacesOldRefreshToken(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	oldToken := models.RefreshToken{
		UserID:    user.ID,
		Token:     "old-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err := testDB.Create(&oldToken).Error
	assert.NoError(t, err)

	reqBody := map[string]string{
		"email":    "ahmed@test.com",
		"password": "password123",
	}

	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	var count int64

	testDB.
		Model(&models.RefreshToken{}).
		Where("user_id = ?", user.ID).
		Count(&count)

	// only one token should exist
	assert.Equal(
		t,
		int64(1),
		count,
	)

	var refreshToken models.RefreshToken

	err = testDB.
		Where("user_id = ?", user.ID).
		First(&refreshToken).
		Error

	assert.NoError(t, err)

	assert.NotEqual(
		t,
		"old-token",
		refreshToken.Token,
	)
}

func TestRefreshTokenSuccess(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	refreshToken := createRefreshToken(
		t,
		user,
	)

	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}

	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/refresh",
		bytes.NewBuffer(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// STATUS CODE
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	// -------------------------
	// RESPONSE DTO
	// -------------------------

	var response dto.RefreshTokenResponse

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.NotEmpty(
		t,
		response.AccessToken,
	)

	// -------------------------
	// DATABASE
	// -------------------------

	// Refresh token should still exist.
	var count int64

	testDB.
		Model(&models.RefreshToken{}).
		Where(
			"user_id = ?",
			user.ID,
		).
		Count(&count)

	assert.Equal(
		t,
		int64(1),
		count,
	)
}

func TestRefreshTokenValidation(t *testing.T) {

	cleanDatabase()

	reqBody := map[string]string{
		"refresh_token": "",
	}

	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/refresh",
		bytes.NewBuffer(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	assert.Equal(
		t,
		http.StatusBadRequest,
		recorder.Code,
	)

	var response map[string]string

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Contains(
		t,
		response["error"],
		"RefreshToken",
	)
}

func TestRefreshTokenNotFoundInDB(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	token,
		_,
		err :=
		utils.GenerateRefreshToken(
			user.ID,
		)

	assert.NoError(t, err)

	reqBody := map[string]string{
		"refresh_token": token,
	}

	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/refresh",
		bytes.NewBuffer(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	assert.Equal(
		t,
		http.StatusBadRequest,
		recorder.Code,
	)

	var response map[string]string

	err = json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		"invalid refresh token",
		response["error"],
	)
}

func TestRefreshTokenExpired(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	claims := utils.RefreshClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(
					-time.Hour,
				),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	tokenString, err := token.SignedString(
		[]byte(
			config.AppConfig.JWTSecret,
		),
	)

	assert.NoError(t, err)

	refreshToken := models.RefreshToken{
		UserID:    user.ID,
		Token:     tokenString,
		ExpiresAt: time.Now().Add(-time.Hour),
	}

	err = testDB.Create(
		&refreshToken,
	).Error

	assert.NoError(t, err)

	reqBody := map[string]string{
		"refresh_token": tokenString,
	}

	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/refresh",
		bytes.NewBuffer(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	assert.Equal(
		t,
		http.StatusBadRequest,
		recorder.Code,
	)

	var response map[string]string

	err = json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		"invalid refresh token",
		response["error"],
	)
}

func TestRefreshTokenInvalidSignature(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	claims := utils.RefreshClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(
					24 * time.Hour,
				),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	// Wrong secret.
	tokenString, err := token.SignedString(
		[]byte("fake-secret"),
	)

	assert.NoError(t, err)

	reqBody := map[string]string{
		"refresh_token": tokenString,
	}

	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/refresh",
		bytes.NewBuffer(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	assert.Equal(
		t,
		http.StatusBadRequest,
		recorder.Code,
	)

	var response map[string]string

	err = json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		"invalid refresh token",
		response["error"],
	)
}

func TestLogoutSuccess(t *testing.T) {

	cleanDatabase()

	// -------------------------
	// ARRANGE
	// -------------------------

	user := createTestUser(t)

	refreshToken := createRefreshToken(
		t,
		user,
	)

	// -------------------------
	// ACT
	// -------------------------

	reqBody := map[string]string{
		"refresh_token": refreshToken,
	}

	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/logout",
		bytes.NewBuffer(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// STATUS CODE
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	// -------------------------
	// RESPONSE DTO
	// -------------------------

	var response dto.MessageResponse

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		"logged out successfully",
		response.Message,
	)

	// -------------------------
	// DATABASE CHANGES
	// -------------------------

	var count int64

	testDB.
		Model(&models.RefreshToken{}).
		Where(
			"token = ?",
			refreshToken,
		).
		Count(&count)

	assert.Equal(
		t,
		int64(0),
		count,
	)
}

func TestLogoutValidation(t *testing.T) {

	cleanDatabase()

	reqBody := map[string]string{
		"refresh_token": "",
	}

	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/logout",
		bytes.NewBuffer(body),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// STATUS CODE
	// -------------------------

	assert.Equal(
		t,
		http.StatusBadRequest,
		recorder.Code,
	)

	// -------------------------
	// ERROR RESPONSE
	// -------------------------

	var response map[string]string

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Contains(
		t,
		response["error"],
		"RefreshToken",
	)
}
