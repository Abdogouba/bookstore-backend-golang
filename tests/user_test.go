package tests

import (
	"bookstore-backend/config"
	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/models"
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

func TestGetProfileSuccess(t *testing.T) {

	cleanDatabase()

	// -------------------------
	// Arrange
	// -------------------------

	user := createTestUser(t)

	accessToken, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(t, err)

	req := authenticatedRequest(
		t,
		http.MethodGet,
		"/users/profile",
		accessToken,
	)

	recorder := httptest.NewRecorder()

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Status Code
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	// -------------------------
	// Response DTO
	// -------------------------

	var response dto.UserResponse

	err = json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		user.ID,
		response.ID,
	)

	assert.Equal(
		t,
		user.Name,
		response.Name,
	)

	assert.Equal(
		t,
		user.Email,
		response.Email,
	)

	assert.Equal(
		t,
		user.PhoneNumber,
		response.PhoneNumber,
	)

	assert.Equal(
		t,
		user.Role,
		response.Role,
	)

	assert.WithinDuration(
		t,
		user.CreatedAt,
		response.CreatedAt,
		time.Second,
	)
}

func TestGetProfileUserNotFound(
	t *testing.T,
) {

	cleanDatabase()

	// -------------------------
	// Arrange
	// -------------------------

	user := createTestUser(t)

	accessToken, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(t, err)

	err = testDB.
		Delete(
			&models.User{},
			user.ID,
		).
		Error

	assert.NoError(t, err)

	req := authenticatedRequest(
		t,
		http.MethodGet,
		"/users/profile",
		accessToken,
	)

	recorder := httptest.NewRecorder()

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Status Code
	// -------------------------

	assert.Equal(
		t,
		http.StatusNotFound,
		recorder.Code,
	)

	// -------------------------
	// Error Response
	// -------------------------

	var response map[string]string

	err = json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		"user not found",
		response["error"],
	)
}

func TestGetProfileNoToken(
	t *testing.T,
) {

	req := httptest.NewRequest(
		http.MethodGet,
		"/users/profile",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	assert.Equal(
		t,
		http.StatusUnauthorized,
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
		"authorization header is required",
		response["error"],
	)
}

func TestGetProfileFakeToken(
	t *testing.T,
) {

	req := authenticatedRequest(
		t,
		http.MethodGet,
		"/users/profile",
		"fake-token",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	assert.Equal(
		t,
		http.StatusUnauthorized,
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
		"invalid access token",
		response["error"],
	)
}

func TestGetProfileExpiredToken(
	t *testing.T,
) {

	cleanDatabase()

	user := createTestUser(t)

	claims := utils.AccessClaims{
		UserID: user.ID,
		Role:   user.Role,
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

	expiredToken, err :=
		token.SignedString(
			[]byte(
				config.AppConfig.JWTSecret,
			),
		)

	assert.NoError(t, err)

	req := authenticatedRequest(
		t,
		http.MethodGet,
		"/users/profile",
		expiredToken,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	assert.Equal(
		t,
		http.StatusUnauthorized,
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
		"invalid access token",
		response["error"],
	)
}

func TestUpdateProfileUserNotFound(
	t *testing.T,
) {

	cleanDatabase()

	user := createTestUser(t)

	accessToken, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(t, err)

	err = testDB.
		Delete(
			&models.User{},
			user.ID,
		).
		Error

	assert.NoError(t, err)

	requestBody := dto.UpdateProfileRequest{
		Name:        "Updated Name",
		Email:       "updated@test.com",
		PhoneNumber: "01099999999",
	}

	jsonBody, _ :=
		json.Marshal(requestBody)

	req := httptest.NewRequest(
		http.MethodPut,
		"/users/profile",
		bytes.NewBuffer(jsonBody),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+accessToken,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	assert.Equal(
		t,
		http.StatusNotFound,
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
		"user not found",
		response["error"],
	)
}

func TestUpdateProfileNoToken(
	t *testing.T,
) {

	requestBody := dto.UpdateProfileRequest{
		Name:        "Updated Name",
		Email:       "updated@test.com",
		PhoneNumber: "01099999999",
	}

	jsonBody, _ :=
		json.Marshal(requestBody)

	req := httptest.NewRequest(
		http.MethodPut,
		"/users/profile",
		bytes.NewBuffer(jsonBody),
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
		http.StatusUnauthorized,
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
		"authorization header is required",
		response["error"],
	)
}

func TestUpdateProfileValidationFail(
	t *testing.T,
) {

	cleanDatabase()

	user := createTestUser(t)

	accessToken, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(t, err)

	requestBody := map[string]string{
		"name":         "",
		"email":        "updated@test.com",
		"phone_number": "01099999999",
	}

	jsonBody, _ :=
		json.Marshal(requestBody)

	req := httptest.NewRequest(
		http.MethodPut,
		"/users/profile",
		bytes.NewBuffer(jsonBody),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+accessToken,
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

	assert.NotEmpty(
		t,
		response["error"],
	)
}

func TestUpdateProfileEmailAlreadyExists(
	t *testing.T,
) {

	cleanDatabase()

	user1 := createTestUser(t)

	user2 := &models.User{
		Name:         "Second User",
		Email:        "second@test.com",
		PasswordHash: "hashed",
		Role:         "user",
	}

	err := testDB.
		Create(user2).
		Error

	assert.NoError(t, err)

	accessToken, err :=
		utils.GenerateAccessToken(
			user1.ID,
			user1.Role,
		)

	assert.NoError(t, err)

	requestBody := dto.UpdateProfileRequest{
		Name:        "Updated Name",
		Email:       user2.Email,
		PhoneNumber: "01099999999",
	}

	jsonBody, _ :=
		json.Marshal(requestBody)

	req := httptest.NewRequest(
		http.MethodPut,
		"/users/profile",
		bytes.NewBuffer(jsonBody),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+accessToken,
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
		"email already exists",
		response["error"],
	)

	// -------------------------
	// Verify DB unchanged
	// -------------------------

	var dbUser models.User

	err = testDB.
		First(
			&dbUser,
			user1.ID,
		).
		Error

	assert.NoError(t, err)

	assert.Equal(
		t,
		user1.Email,
		dbUser.Email,
	)
}

func TestUpdateProfileSuccess(t *testing.T) {

	cleanDatabase()

	// -------------------------
	// Arrange
	// -------------------------

	user := createTestUser(t)

	time.Sleep(10 * time.Millisecond)

	accessToken, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(t, err)

	requestBody := dto.UpdateProfileRequest{
		Name:        "Updated Name",
		Email:       "updated@test.com",
		PhoneNumber: "01099999999",
	}

	jsonBody, err :=
		json.Marshal(requestBody)

	assert.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPut,
		"/users/profile",
		bytes.NewBuffer(jsonBody),
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+accessToken,
	)

	recorder := httptest.NewRecorder()

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Status Code
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	// -------------------------
	// Response DTO
	// -------------------------

	var response dto.UserResponse

	err = json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		user.ID,
		response.ID,
	)

	assert.Equal(
		t,
		requestBody.Name,
		response.Name,
	)

	assert.Equal(
		t,
		requestBody.Email,
		response.Email,
	)

	assert.Equal(
		t,
		requestBody.PhoneNumber,
		response.PhoneNumber,
	)

	assert.WithinDuration(
		t,
		user.CreatedAt,
		response.CreatedAt,
		time.Second,
	)

	// -------------------------
	// Database Checks
	// -------------------------

	var updatedUser models.User

	err = testDB.
		First(
			&updatedUser,
			user.ID,
		).
		Error

	assert.NoError(t, err)

	// Updated fields

	assert.Equal(
		t,
		requestBody.Name,
		updatedUser.Name,
	)

	assert.Equal(
		t,
		requestBody.Email,
		updatedUser.Email,
	)

	assert.Equal(
		t,
		requestBody.PhoneNumber,
		updatedUser.PhoneNumber,
	)

	// Unchanged fields

	assert.Equal(
		t,
		user.ID,
		updatedUser.ID,
	)

	assert.Equal(
		t,
		user.Role,
		updatedUser.Role,
	)

	assert.Equal(
		t,
		user.PasswordHash,
		updatedUser.PasswordHash,
	)

	assert.WithinDuration(
		t,
		user.CreatedAt,
		updatedUser.CreatedAt,
		time.Second,
	)

	// UpdatedAt should change

	assert.True(
		t,
		updatedUser.UpdatedAt.After(
			user.UpdatedAt,
		),
	)
}

func TestChangePasswordSuccess(t *testing.T) {

	cleanDatabase()

	// -------------------------
	// Arrange
	// -------------------------

	user := createTestUser(t)

	// Create a refresh token that should be deleted.
	refreshToken, _, err := utils.GenerateRefreshToken(user.ID)

	assert.NoError(t, err)

	err = testDB.Create(&models.RefreshToken{
		UserID: user.ID,
		Token:  refreshToken,
	}).Error

	assert.NoError(t, err)

	accessToken, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(t, err)

	requestBody := dto.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newpassword123",
	}

	jsonBody, err := json.Marshal(requestBody)

	assert.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodPut,
		"/users/change-password",
		bytes.NewBuffer(jsonBody),
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	recorder := httptest.NewRecorder()

	// -------------------------
	// Act
	// -------------------------

	router.ServeHTTP(recorder, req)

	// -------------------------
	// Status Code
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	// -------------------------
	// Response
	// -------------------------

	var response map[string]string

	err = json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		"password changed successfully",
		response["message"],
	)

	// -------------------------
	// Database User Checks
	// -------------------------

	var updatedUser models.User

	err = testDB.First(
		&updatedUser,
		user.ID,
	).Error

	assert.NoError(t, err)

	// New password should succeed.

	err = utils.CheckPassword(
		requestBody.NewPassword,
		updatedUser.PasswordHash,
	)

	assert.NoError(t, err)

	// Other fields unchanged.

	assert.Equal(
		t,
		user.ID,
		updatedUser.ID,
	)

	assert.Equal(
		t,
		user.Name,
		updatedUser.Name,
	)

	assert.Equal(
		t,
		user.Email,
		updatedUser.Email,
	)

	assert.Equal(
		t,
		user.PhoneNumber,
		updatedUser.PhoneNumber,
	)

	assert.Equal(
		t,
		user.Role,
		updatedUser.Role,
	)

	// -------------------------
	// Refresh Tokens Deleted
	// -------------------------

	var count int64

	err = testDB.
		Model(&models.RefreshToken{}).
		Where("user_id = ?", user.ID).
		Count(&count).
		Error

	assert.NoError(t, err)

	assert.Equal(
		t,
		int64(0),
		count,
	)
}

func TestChangePasswordUserNotFound(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	accessToken, _ :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	testDB.Delete(
		&models.User{},
		user.ID,
	)

	requestBody := dto.ChangePasswordRequest{
		CurrentPassword: "password123",
		NewPassword:     "newpassword123",
	}

	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(
		http.MethodPut,
		"/users/change-password",
		bytes.NewBuffer(body),
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+accessToken,
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
		http.StatusNotFound,
		recorder.Code,
	)

	var response map[string]string

	json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.Equal(
		t,
		"user not found",
		response["error"],
	)
}

func TestChangePasswordNoToken(t *testing.T) {

	requestBody := dto.ChangePasswordRequest{
		CurrentPassword: "testPassword",
		NewPassword:     "newpassword123",
	}

	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(
		http.MethodPut,
		"/users/change-password",
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
		http.StatusUnauthorized,
		recorder.Code,
	)
}

func TestChangePasswordValidationFail(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	token, _ :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	requestBody := map[string]string{
		"current_password": "",
		"new_password":     "newpassword123",
	}

	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(
		http.MethodPut,
		"/users/change-password",
		bytes.NewBuffer(body),
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
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

	var updatedUser models.User

	err := testDB.First(
		&updatedUser,
		user.ID,
	).Error

	assert.NoError(t, err)

	err = utils.CheckPassword(
		"password123",
		updatedUser.PasswordHash,
	)

	assert.NoError(t, err)
}

func TestChangePasswordWrongPassword(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	token, _ :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	requestBody := dto.ChangePasswordRequest{
		CurrentPassword: "wrongpassword",
		NewPassword:     "newpassword123",
	}

	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(
		http.MethodPut,
		"/users/change-password",
		bytes.NewBuffer(body),
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
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

	json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.Equal(
		t,
		"current password is incorrect",
		response["error"],
	)

	// Password should remain unchanged.

	var dbUser models.User

	testDB.First(
		&dbUser,
		user.ID,
	)

	err := utils.CheckPassword(
		"password123",
		dbUser.PasswordHash,
	)

	assert.NoError(t, err)
}

func TestDeleteProfileSuccess(t *testing.T) {

	cleanDatabase()

	// -------------------------
	// Arrange
	// -------------------------
	user := createTestUser(t)

	err := testDB.Create(&models.RefreshToken{
		UserID: user.ID,
		Token:  "dummy-token",
	}).Error
	assert.NoError(t, err)

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Role)
	assert.NoError(t, err)

	requestBody := dto.DeleteProfileRequest{
		Password: "password123",
	}

	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/users/profile",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	recorder := httptest.NewRecorder()

	// -------------------------
	// Act
	// -------------------------
	router.ServeHTTP(recorder, req)

	// -------------------------
	// Assert status
	// -------------------------
	assert.Equal(t, http.StatusOK, recorder.Code)

	// -------------------------
	// Response
	// -------------------------
	var response map[string]string
	json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Equal(t, "profile deleted successfully", response["message"])

	// -------------------------
	// DB checks - user soft deleted
	// -------------------------
	var count int64
	err = testDB.Model(&models.User{}).
		Where("id = ? AND deleted_at IS NULL", user.ID).
		Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// -------------------------
	// DB checks - refresh tokens deleted
	// -------------------------
	err = testDB.Model(&models.RefreshToken{}).
		Where("user_id = ?", user.ID).
		Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestDeleteProfileNoToken(t *testing.T) {

	cleanDatabase()

	req := httptest.NewRequest(
		http.MethodDelete,
		"/users/profile",
		bytes.NewBuffer([]byte(`{"password":"password123"}`)),
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestDeleteProfileAdminForbidden(t *testing.T) {

	cleanDatabase()

	// create admin user
	hashed, _ := utils.HashPassword("password123")

	admin := models.User{
		Name:         "Admin",
		Email:        "admin@test.com",
		PasswordHash: hashed,
		Role:         "admin",
	}

	testDB.Create(&admin)

	token, _ := utils.GenerateAccessToken(admin.ID, admin.Role)

	body, _ := json.Marshal(dto.DeleteProfileRequest{
		Password: "password123",
	})

	req := httptest.NewRequest(
		http.MethodDelete,
		"/users/profile",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusForbidden, recorder.Code)

	// -------------------------
	// Database Check
	// -------------------------

	var dbAdmin models.User

	err := testDB.First(
		&dbAdmin,
		admin.ID,
	).Error

	assert.NoError(t, err)

	assert.Equal(
		t,
		admin.ID,
		dbAdmin.ID,
	)
}

func TestDeleteProfileInvalidDTO(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	// missing password
	body := []byte(`{"password":""}`)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/users/profile",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// user should still exist
	var count int64
	testDB.Model(&models.User{}).
		Where("id = ?", user.ID).
		Count(&count)

	assert.Equal(t, int64(1), count)
}

func TestDeleteProfileUserNotFound(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	// delete user before request
	testDB.Delete(&models.User{}, user.ID)

	body, _ := json.Marshal(dto.DeleteProfileRequest{
		Password: "password123",
	})

	req := httptest.NewRequest(
		http.MethodDelete,
		"/users/profile",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestDeleteProfileWrongPassword(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	body, _ := json.Marshal(dto.DeleteProfileRequest{
		Password: "wrongpassword",
	})

	req := httptest.NewRequest(
		http.MethodDelete,
		"/users/profile",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// user should still exist
	var dbUser models.User
	testDB.First(&dbUser, user.ID)

	assert.Equal(t, user.Email, dbUser.Email)
}

func TestDeleteProfileHasActiveOrders(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	createTestOrder(t, user.ID, "pending")

	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	body, _ := json.Marshal(dto.DeleteProfileRequest{
		Password: "password123",
	})

	req := httptest.NewRequest(
		http.MethodDelete,
		"/users/profile",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response map[string]string
	json.Unmarshal(recorder.Body.Bytes(), &response)

	assert.Equal(
		t,
		"cannot delete account with active orders",
		response["error"],
	)

	// user must still exist
	var count int64
	testDB.Model(&models.User{}).
		Where("id = ?", user.ID).
		Count(&count)

	assert.Equal(t, int64(1), count)
}