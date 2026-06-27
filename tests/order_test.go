package tests

import (
	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/models"
	"bookstore-backend/internal/seeder"
	"bookstore-backend/internal/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateOrder_NoToken_Fail(t *testing.T) {

	cleanDatabase()
	err := seeder.SeedAdmin(testDB)
	assert.NoError(t, err)

	book := seedBook("Book 1", 10, 100)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo",
		Items: []dto.CreateOrderItemRequest{
			{BookID: book.ID, Quantity: 1},
		},
	}, "")

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCreateOrder_Admin_Fail(t *testing.T) {

	cleanDatabase()
	err := seeder.SeedAdmin(testDB)
	assert.NoError(t, err)

	adminToken := getAdminToken(t)

	book := seedBook("Book 1", 10, 100)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo",
		Items: []dto.CreateOrderItemRequest{
			{BookID: book.ID, Quantity: 1},
		},
	}, adminToken)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCreateOrder_EmptyAddress_Fail(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	token := createTestUser(t)
	userToken, _ := utils.GenerateAccessToken(token.ID, token.Role)

	book := seedBook("Book 1", 10, 100)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "",
		Items: []dto.CreateOrderItemRequest{
			{BookID: book.ID, Quantity: 1},
		},
	}, userToken)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateOrder_NoItems_Fail(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	user := createTestUser(t)
	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo",
		Items:   []dto.CreateOrderItemRequest{},
	}, token)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateOrder_InvalidQuantity_Fail(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	user := createTestUser(t)
	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	book := seedBook("Book 1", 10, 100)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo",
		Items: []dto.CreateOrderItemRequest{
			{BookID: book.ID, Quantity: 0},
		},
	}, token)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateOrder_DuplicateBooks_Fail(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	user := createTestUser(t)
	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	book := seedBook("Book 1", 10, 100)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo",
		Items: []dto.CreateOrderItemRequest{
			{BookID: book.ID, Quantity: 1},
			{BookID: book.ID, Quantity: 2},
		},
	}, token)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateOrder_BookNotFound_Fail(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	user := createTestUser(t)
	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo",
		Items: []dto.CreateOrderItemRequest{
			{BookID: 99999, Quantity: 1},
		},
	}, token)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCreateOrder_InsufficientStock_Fail(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	user := createTestUser(t)
	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	book := seedBook("Book 1", 2, 100)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo",
		Items: []dto.CreateOrderItemRequest{
			{BookID: book.ID, Quantity: 10},
		},
	}, token)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateOrder_Success(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	user := createTestUser(t)
	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	book1 := seedBook("Book 1", 10, 100)
	book2 := seedBook("Book 2", 5, 200)

	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo, Egypt",
		Items: []dto.CreateOrderItemRequest{
			{BookID: book1.ID, Quantity: 2},
			{BookID: book2.ID, Quantity: 1},
		},
	}, token)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// =========================
	// ASSERT RESPONSE
	// =========================
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response dto.CreateOrderResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "pending", response.Status)
	assert.Equal(t, "Cairo, Egypt", response.Address)
	assert.Equal(t, float64(400), response.TotalPrice) // (2*100 + 1*200)

	assert.Len(t, response.Items, 2)

	// check DTO fields
	assert.NotZero(t, response.ID)
	assert.NotZero(t, response.CreatedAt)

	// =========================
	// ASSERT DATABASE (ORDER)
	// =========================
	var order models.Order
	err = testDB.Preload("OrderItems").
		Where("id = ?", response.ID).
		First(&order).Error

	assert.NoError(t, err)
	assert.Equal(t, "pending", order.Status)
	assert.Equal(t, float64(400), order.TotalPrice)

	// =========================
	// ASSERT ORDER ITEMS
	// =========================
	assert.Len(t, order.OrderItems, 2)

	// =========================
	// ASSERT STOCK REDUCTION
	// =========================
	var updatedBook1 models.Book
	testDB.First(&updatedBook1, book1.ID)
	assert.Equal(t, 8, updatedBook1.Stock)

	var updatedBook2 models.Book
	testDB.First(&updatedBook2, book2.ID)
	assert.Equal(t, 4, updatedBook2.Stock)
}

func TestCreateOrder_TransactionRollback(t *testing.T) {

	cleanDatabase()
	seeder.SeedAdmin(testDB)

	user := createTestUser(t)
	token, _ := utils.GenerateAccessToken(user.ID, user.Role)

	book := seedBook("Book 1", 10, 100)

	// simulate failure by requesting invalid second item
	req := createOrderRequest(dto.CreateOrderRequest{
		Address: "Cairo",
		Items: []dto.CreateOrderItemRequest{
			{BookID: book.ID, Quantity: 1},
			{BookID: 99999, Quantity: 1}, // invalid book triggers failure
		},
	}, token)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	// ORDER SHOULD NOT EXIST
	var count int64
	testDB.Model(&models.Order{}).Count(&count)
	assert.Equal(t, int64(0), count)

	// STOCK SHOULD NOT CHANGE
	var updatedBook models.Book
	testDB.First(&updatedBook, book.ID)
	assert.Equal(t, 10, updatedBook.Stock)
}

func TestGetMyOrdersNoToken(t *testing.T) {

	cleanDatabase()

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders",
		nil,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusUnauthorized,
		rec.Code,
	)

	var response map[string]string

	err := json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		"authorization header is required",
		response["error"],
	)
}

func TestGetMyOrdersAdminForbidden(t *testing.T) {

	cleanDatabase()

	err := seeder.SeedAdmin(testDB)
	assert.NoError(t, err)

	token := getAdminToken(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusForbidden,
		rec.Code,
	)

	var response map[string]string

	err = json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Equal(
		t,
		"forbidden",
		response["error"],
	)
}

func TestGetMyOrdersEmpty(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	token, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(t, err)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusOK,
		rec.Code,
	)

	var response dto.UserOrdersResponse

	err = json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Empty(
		t,
		response.Orders,
	)

	assert.Equal(
		t,
		int64(0),
		response.Total,
	)

	assert.Equal(
		t,
		1,
		response.Page,
	)

	assert.Equal(
		t,
		10,
		response.PageSize,
	)
}

func TestGetMyOrdersSortedLatestFirst(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	token, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(t, err)

	oldOrder :=
		createUserOrder(
			t,
			user.ID,
			"pending",
			100,
			1,
		)

	time.Sleep(time.Second)

	newOrder :=
		createUserOrder(
			t,
			user.ID,
			"confirmed",
			200,
			2,
		)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusOK,
		rec.Code,
	)

	var response dto.UserOrdersResponse

	err = json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Len(
		t,
		response.Orders,
		2,
	)

	assert.Equal(
		t,
		newOrder.ID,
		response.Orders[0].ID,
	)

	assert.Equal(
		t,
		oldOrder.ID,
		response.Orders[1].ID,
	)
}

func TestGetMyOrdersPagination(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	token, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(t, err)

	for i := 0; i < 15; i++ {

		createUserOrder(
			t,
			user.ID,
			"pending",
			100,
			1,
		)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders?page=2&page_size=5",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(
		t,
		http.StatusOK,
		rec.Code,
	)

	var response dto.UserOrdersResponse

	err = json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Len(
		t,
		response.Orders,
		5,
	)

	assert.Equal(
		t,
		2,
		response.Page,
	)

	assert.Equal(
		t,
		5,
		response.PageSize,
	)

	assert.Equal(
		t,
		int64(15),
		response.Total,
	)
}

func TestGetMyOrdersSuccess(t *testing.T) {

	cleanDatabase()

	user := createTestUser(t)

	token, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(t, err)

	order :=
		createUserOrder(
			t,
			user.ID,
			"pending",
			300,
			3,
		)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+token,
	)

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// -------------------------
	// Status
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		rec.Code,
	)

	// -------------------------
	// Response DTO
	// -------------------------

	var response dto.UserOrdersResponse

	err = json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	assert.Len(
		t,
		response.Orders,
		1,
	)

	assert.Equal(
		t,
		int64(1),
		response.Total,
	)

	assert.Equal(
		t,
		1,
		response.Page,
	)

	assert.Equal(
		t,
		10,
		response.PageSize,
	)

	orderResponse :=
		response.Orders[0]

	assert.Equal(
		t,
		order.ID,
		orderResponse.ID,
	)

	assert.Equal(
		t,
		"pending",
		orderResponse.Status,
	)

	assert.Equal(
		t,
		300.0,
		orderResponse.TotalPrice,
	)

	// total quantity
	assert.Equal(
		t,
		3,
		orderResponse.ItemsCount,
	)

	assert.False(
		t,
		orderResponse.CreatedAt.IsZero(),
	)
}

func TestGetMyOrder_NoToken(
	t *testing.T,
) {

	cleanDatabase()

	// -------------------------
	// Arrange
	// -------------------------

	user := createTestUser(t)

	order := createUserOrder(
		t,
		user.ID,
		"pending",
		100,
		1,
	)

	req := createGetOrderRequest(
		"",
		order.ID,
	)

	// -------------------------
	// Act
	// -------------------------

	rec := httptest.NewRecorder()

	router.ServeHTTP(
		rec,
		req,
	)

	// -------------------------
	// Assert
	// -------------------------

	assert.Equal(
		t,
		http.StatusUnauthorized,
		rec.Code,
	)
}

func TestGetMyOrder_AdminForbidden(
	t *testing.T,
) {

	cleanDatabase()

	// -------------------------
	// Arrange
	// -------------------------

	seeder.SeedAdmin(testDB)

	user := createTestUser(t)

	order := createUserOrder(
		t,
		user.ID,
		"pending",
		100,
		1,
	)

	adminToken :=
		getAdminToken(t)

	req := createGetOrderRequest(
		adminToken,
		order.ID,
	)

	// -------------------------
	// Act
	// -------------------------

	rec := httptest.NewRecorder()

	router.ServeHTTP(
		rec,
		req,
	)

	// -------------------------
	// Assert
	// -------------------------

	assert.Equal(
		t,
		http.StatusForbidden,
		rec.Code,
	)
}

func TestGetMyOrder_OrderNotFound(
	t *testing.T,
) {

	cleanDatabase()

	// -------------------------
	// Arrange
	// -------------------------

	user := createTestUser(t)

	token, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(
		t,
		err,
	)

	req := createGetOrderRequest(
		token,
		999999,
	)

	// -------------------------
	// Act
	// -------------------------

	rec := httptest.NewRecorder()

	router.ServeHTTP(
		rec,
		req,
	)

	// -------------------------
	// Assert
	// -------------------------

	assert.Equal(
		t,
		http.StatusNotFound,
		rec.Code,
	)
}

func TestGetMyOrder_OrderBelongsToAnotherUser(
	t *testing.T,
) {

	cleanDatabase()

	// -------------------------
	// Arrange
	// -------------------------

	user1 := createTestUser(t)

	user2 := models.User{
		Name:         "Omar",
		Email:        "omar@test.com",
		PasswordHash: user1.PasswordHash,
		PhoneNumber:  "01000000000",
		Role:         "user",
	}

	err := testDB.Create(&user2).Error
	assert.NoError(t, err)

	order := createUserOrder(
		t,
		user2.ID,
		"pending",
		100,
		1,
	)

	token, err := utils.GenerateAccessToken(
		user1.ID,
		user1.Role,
	)

	assert.NoError(t, err)

	req := createGetOrderRequest(
		token,
		order.ID,
	)

	// -------------------------
	// Act
	// -------------------------

	rec := httptest.NewRecorder()

	router.ServeHTTP(
		rec,
		req,
	)

	// -------------------------
	// Assert
	// -------------------------

	assert.Equal(
		t,
		http.StatusNotFound,
		rec.Code,
	)
}

func TestGetMyOrder_Success(
	t *testing.T,
) {

	cleanDatabase()

	// -------------------------
	// Arrange
	// -------------------------

	user := createTestUser(t)

	book := models.Book{
		Title:     "Clean Code",
		Author:    "Robert Martin",
		Publisher: "Prentice Hall",
		Category:  "Programming",
		Price:     100,
		Stock:     20,
		ImagePath: "/uploads/books/clean-code.jpg",
	}

	err := testDB.Create(&book).Error
	assert.NoError(t, err)

	order := models.Order{
		UserID:          user.ID,
		Status:          "pending",
		ShippingAddress: "Cairo",
		TotalPrice:      200,
	}

	err = testDB.Create(&order).Error
	assert.NoError(t, err)

	item := models.OrderItem{
		OrderID:   order.ID,
		BookID:    book.ID,
		Quantity:  2,
		Price:     book.Price,
		Title:     book.Title,
		Author:    book.Author,
		Publisher: book.Publisher,
		ImagePath: book.ImagePath,
	}

	err = testDB.Create(&item).Error
	assert.NoError(t, err)

	token, err := utils.GenerateAccessToken(
		user.ID,
		user.Role,
	)

	assert.NoError(t, err)

	req := createGetOrderRequest(
		token,
		order.ID,
	)

	// -------------------------
	// Act
	// -------------------------

	rec := httptest.NewRecorder()

	router.ServeHTTP(
		rec,
		req,
	)

	// -------------------------
	// Assert Status
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		rec.Code,
	)

	// -------------------------
	// Parse Response
	// -------------------------

	var response dto.UserOrderDetailsResponse

	err = json.Unmarshal(
		rec.Body.Bytes(),
		&response,
	)

	assert.NoError(t, err)

	// -------------------------
	// Order Fields
	// -------------------------

	assert.Equal(
		t,
		order.ID,
		response.ID,
	)

	assert.Equal(
		t,
		order.Status,
		response.Status,
	)

	assert.Equal(
		t,
		order.ShippingAddress,
		response.Address,
	)

	assert.Equal(
		t,
		order.TotalPrice,
		response.TotalPrice,
	)

	assert.False(
		t,
		response.CreatedAt.IsZero(),
	)

	assert.False(
		t,
		response.UpdatedAt.IsZero(),
	)

	// -------------------------
	// Order Items
	// -------------------------

	assert.Len(
		t,
		response.Items,
		1,
	)

	responseItem := response.Items[0]

	assert.Equal(
		t,
		item.ID,
		responseItem.ID,
	)

	assert.Equal(
		t,
		item.BookID,
		responseItem.BookID,
	)

	assert.Equal(
		t,
		item.Quantity,
		responseItem.Quantity,
	)

	assert.Equal(
		t,
		item.Price,
		responseItem.Price,
	)

	assert.Equal(
		t,
		item.Title,
		responseItem.Title,
	)

	assert.Equal(
		t,
		item.Author,
		responseItem.Author,
	)

	assert.Equal(
		t,
		item.Publisher,
		responseItem.Publisher,
	)

	assert.Equal(
		t,
		item.ImagePath,
		responseItem.ImagePath,
	)
}

func TestAdminGetOrders_NoToken(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	// -------------------------
	// Request without token
	// -------------------------

	req := httptest.NewRequest(
		http.MethodGet,
		"/admin/orders",
		nil,
	)

	// -------------------------
	// Execute request
	// -------------------------

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

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

func TestAdminGetOrders_NotAdmin(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	// -------------------------
	// Create normal user
	// -------------------------

	user :=
		createTestUser(t)

	// -------------------------
	// Generate access token
	// -------------------------

	token, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(
		t,
		err,
	)

	// -------------------------
	// Build request
	// -------------------------

	req :=
		authenticatedRequest(
			t,
			http.MethodGet,
			"/admin/orders",
			token,
		)

	// -------------------------
	// Execute request
	// -------------------------

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

	assert.Equal(
		t,
		http.StatusForbidden,
		recorder.Code,
	)

	var response map[string]string

	err = json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(
		t,
		err,
	)

	assert.Equal(
		t,
		"forbidden",
		response["error"],
	)
}

func TestAdminGetOrders_NoOrders(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	// -------------------------
	// Admin token
	// -------------------------

	seeder.SeedAdmin(testDB)

	token :=
		getAdminToken(t)

	// -------------------------
	// Build request
	// -------------------------

	req :=
		authenticatedRequest(
			t,
			http.MethodGet,
			"/admin/orders",
			token,
		)

	// -------------------------
	// Execute request
	// -------------------------

	recorder :=
		httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	var response dto.AdminOrdersResponse

	err :=
		json.Unmarshal(
			recorder.Body.Bytes(),
			&response,
		)

	assert.NoError(
		t,
		err,
	)

	assert.Empty(
		t,
		response.Orders,
	)

	assert.Equal(
		t,
		int64(0),
		response.Total,
	)

	assert.Equal(
		t,
		1,
		response.Page,
	)

	assert.Equal(
		t,
		10,
		response.PageSize,
	)
}

func TestAdminGetOrders_SearchByUserName(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	// -------------------------
	// Admin token
	// -------------------------

	seeder.SeedAdmin(testDB)

	token :=
		getAdminToken(t)

	// -------------------------
	// Create Users
	// -------------------------

	password, err :=
		utils.HashPassword(
			"password123",
		)

	assert.NoError(
		t,
		err,
	)

	user1 := models.User{
		Name:         "Ahmed Mohamed",
		Email:        "ahmed@test.com",
		PasswordHash: password,
		Role:         "user",
		PhoneNumber: "01289566682",
	}

	user2 := models.User{
		Name:         "Omar Ali",
		Email:        "omar@test.com",
		PhoneNumber: "01289566682",
		PasswordHash: password,
		Role:         "user",
	}

	assert.NoError(
		t,
		testDB.Create(&user1).Error,
	)

	assert.NoError(
		t,
		testDB.Create(&user2).Error,
	)

	// -------------------------
	// Create Orders
	// -------------------------

	createUserOrder(
		t,
		user1.ID,
		"pending",
		200,
		2,
	)

	createUserOrder(
		t,
		user2.ID,
		"pending",
		200,
		2,
	)

	// -------------------------
	// Request
	// -------------------------

	req :=
		authenticatedRequest(
			t,
			http.MethodGet,
			"/admin/orders?user_name=Ahmed",
			token,
		)

	// -------------------------
	// Execute
	// -------------------------

	recorder :=
		httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	var response dto.AdminOrdersResponse

	err =
		json.Unmarshal(
			recorder.Body.Bytes(),
			&response,
		)

	assert.NoError(
		t,
		err,
	)

	assert.Len(
		t,
		response.Orders,
		1,
	)

	order :=
		response.Orders[0]

	assert.Equal(
		t,
		"Ahmed Mohamed",
		order.UserName,
	)

	assert.Equal(
		t,
		"pending",
		order.Status,
	)

	assert.Equal(
		t,
		"Cairo",
		order.Address,
	)

	assert.Equal(
		t,
		200.0,
		order.TotalPrice,
	)

	assert.Equal(
		t,
		2,
		order.ItemsCount,
	)

	assert.NotZero(
		t,
		order.ID,
	)

	assert.False(
		t,
		order.CreatedAt.IsZero(),
	)

	assert.Equal(
		t,
		int64(1),
		response.Total,
	)
}

func TestAdminGetOrders_Pagination(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	// -------------------------
	// Admin token
	// -------------------------

	seeder.SeedAdmin(testDB)

	token :=
		getAdminToken(t)

	// -------------------------
	// Create user
	// -------------------------

	user :=
		createTestUser(t)

	// -------------------------
	// Create 15 orders
	// -------------------------

	for i := 0; i < 15; i++ {

		createUserOrder(
			t,
			user.ID,
			"pending",
			100,
			1,
		)
	}

	// -------------------------
	// Build request
	// -------------------------

	req :=
		authenticatedRequest(
			t,
			http.MethodGet,
			"/admin/orders?page=2&page_size=5",
			token,
		)

	// -------------------------
	// Execute request
	// -------------------------

	recorder :=
		httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	var response dto.AdminOrdersResponse

	err :=
		json.Unmarshal(
			recorder.Body.Bytes(),
			&response,
		)

	assert.NoError(
		t,
		err,
	)

	assert.Len(
		t,
		response.Orders,
		5,
	)

	assert.Equal(
		t,
		int64(15),
		response.Total,
	)

	assert.Equal(
		t,
		2,
		response.Page,
	)

	assert.Equal(
		t,
		5,
		response.PageSize,
	)
}

func TestAdminGetOrders_FilterByStatus(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	// -------------------------
	// Admin token
	// -------------------------

	seeder.SeedAdmin(testDB)

	token :=
		getAdminToken(t)

	// -------------------------
	// Create user
	// -------------------------

	user :=
		createTestUser(t)

	// -------------------------
	// Create orders
	// -------------------------

	createUserOrder(
		t,
		user.ID,
		"pending",
		100,
		1,
	)

	createUserOrder(
		t,
		user.ID,
		"confirmed",
		200,
		2,
	)

	createUserOrder(
		t,
		user.ID,
		"delivered",
		300,
		3,
	)

	// -------------------------
	// Build request
	// -------------------------

	req :=
		authenticatedRequest(
			t,
			http.MethodGet,
			"/admin/orders?status=confirmed",
			token,
		)

	// -------------------------
	// Execute request
	// -------------------------

	recorder :=
		httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	var response dto.AdminOrdersResponse

	err :=
		json.Unmarshal(
			recorder.Body.Bytes(),
			&response,
		)

	assert.NoError(
		t,
		err,
	)

	assert.Len(
		t,
		response.Orders,
		1,
	)

	assert.Equal(
		t,
		"confirmed",
		response.Orders[0].Status,
	)

	assert.Equal(
		t,
		int64(1),
		response.Total,
	)
}

func TestAdminGetOrders_SortedLatestFirst(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	// -------------------------
	// Admin token
	// -------------------------

	seeder.SeedAdmin(testDB)

	token := getAdminToken(t)

	// -------------------------
	// Create user
	// -------------------------

	user := createTestUser(t)

	// -------------------------
	// Create older order
	// -------------------------

	oldOrder := createUserOrder(
		t,
		user.ID,
		"pending",
		100,
		1,
	)

	// Ensure different timestamps
	time.Sleep(time.Second)

	// -------------------------
	// Create newer order
	// -------------------------

	newOrder := createUserOrder(
		t,
		user.ID,
		"confirmed",
		200,
		2,
	)

	// -------------------------
	// Build request
	// -------------------------

	req := authenticatedRequest(
		t,
		http.MethodGet,
		"/admin/orders",
		token,
	)

	// -------------------------
	// Execute request
	// -------------------------

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	var response dto.AdminOrdersResponse

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(
		t,
		err,
	)

	assert.Len(
		t,
		response.Orders,
		2,
	)

	// Latest order should come first.
	assert.Equal(
		t,
		newOrder.ID,
		response.Orders[0].ID,
	)

	assert.Equal(
		t,
		oldOrder.ID,
		response.Orders[1].ID,
	)
}

func TestAdminGetOrders_Success(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	// -------------------------
	// Admin token
	// -------------------------
	seeder.SeedAdmin(testDB)

	token := getAdminToken(t)

	// -------------------------
	// Create user
	// -------------------------

	user := createTestUser(t)

	// -------------------------
	// Create order
	// -------------------------

	order := createUserOrder(
		t,
		user.ID,
		"confirmed",
		250,
		3,
	)

	// -------------------------
	// Build request
	// -------------------------

	req := authenticatedRequest(
		t,
		http.MethodGet,
		"/admin/orders",
		token,
	)

	// -------------------------
	// Execute request
	// -------------------------

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	var response dto.AdminOrdersResponse

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(
		t,
		err,
	)

	assert.Len(
		t,
		response.Orders,
		1,
	)

	assert.Equal(
		t,
		int64(1),
		response.Total,
	)

	assert.Equal(
		t,
		1,
		response.Page,
	)

	assert.Equal(
		t,
		10,
		response.PageSize,
	)

	result := response.Orders[0]

	assert.Equal(
		t,
		order.ID,
		result.ID,
	)

	assert.Equal(
		t,
		user.Name,
		result.UserName,
	)

	assert.Equal(
		t,
		"confirmed",
		result.Status,
	)

	assert.Equal(
		t,
		"Cairo",
		result.Address,
	)

	assert.Equal(
		t,
		250.0,
		result.TotalPrice,
	)

	// Quantity was 3, so ItemsCount should be 3.
	assert.Equal(
		t,
		3,
		result.ItemsCount,
	)

	assert.False(
		t,
		result.CreatedAt.IsZero(),
	)
}

func TestAdminGetOrder_NoToken(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	// -------------------------
	// Create order
	// -------------------------

	user := createTestUser(t)

	order := createUserOrder(
		t,
		user.ID,
		"pending",
		100,
		1,
	)

	// -------------------------
	// Build request
	// -------------------------

	req := createAdminViewOrderRequest(
		"",
		order.ID,
	)

	// -------------------------
	// Execute request
	// -------------------------

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

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

	assert.NoError(
		t,
		err,
	)

	assert.Equal(
		t,
		"authorization header is required",
		response["error"],
	)
}

func TestAdminGetOrder_NotAdmin(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	// -------------------------
	// Create user
	// -------------------------

	user := createTestUser(t)

	// -------------------------
	// User access token
	// -------------------------

	token, err := utils.GenerateAccessToken(
		user.ID,
		user.Role,
	)

	assert.NoError(
		t,
		err,
	)

	// -------------------------
	// Create order
	// -------------------------

	order := createUserOrder(
		t,
		user.ID,
		"pending",
		100,
		1,
	)

	// -------------------------
	// Build request
	// -------------------------

	req := createAdminViewOrderRequest(
		token,
		order.ID,
	)

	// -------------------------
	// Execute request
	// -------------------------

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

	assert.Equal(
		t,
		http.StatusForbidden,
		recorder.Code,
	)

	var response map[string]string

	err = json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(
		t,
		err,
	)

	assert.Equal(
		t,
		"forbidden",
		response["error"],
	)
}

func TestAdminGetOrder_OrderNotFound(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	// -------------------------
	// Admin token
	// -------------------------

	seeder.SeedAdmin(testDB)

	token := getAdminToken(t)

	// -------------------------
	// Build request
	// -------------------------

	req := createAdminViewOrderRequest(
		token,
		999,
	)

	// -------------------------
	// Execute request
	// -------------------------

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

	assert.Equal(
		t,
		http.StatusNotFound,
		recorder.Code,
	)

	var response map[string]string

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(
		t,
		err,
	)

	assert.Equal(
		t,
		"order not found",
		response["error"],
	)
}

func TestAdminGetOrder_Success(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	// -------------------------
	// Admin token
	// -------------------------

	seeder.SeedAdmin(testDB)

	token := getAdminToken(t)

	// -------------------------
	// Create user
	// -------------------------

	user := createTestUser(t)

	// -------------------------
	// Create order
	// -------------------------

	order := createUserOrder(
		t,
		user.ID,
		"confirmed",
		300,
		3,
	)

	// -------------------------
	// Load complete order
	// (includes user & items)
	// -------------------------

	var dbOrder models.Order

	err := testDB.
		Preload("User").
		Preload("OrderItems").
		First(&dbOrder, order.ID).
		Error

	assert.NoError(
		t,
		err,
	)

	// -------------------------
	// Build request
	// -------------------------

	req := createAdminViewOrderRequest(
		token,
		order.ID,
	)

	// -------------------------
	// Execute request
	// -------------------------

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	var response dto.AdminOrderResponse

	err = json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(
		t,
		err,
	)

	// -------------------------
	// Order fields
	// -------------------------

	assert.Equal(
		t,
		dbOrder.ID,
		response.ID,
	)

	assert.Equal(
		t,
		dbOrder.Status,
		response.Status,
	)

	assert.Equal(
		t,
		dbOrder.ShippingAddress,
		response.Address,
	)

	assert.Equal(
		t,
		dbOrder.TotalPrice,
		response.TotalPrice,
	)

	assert.Equal(
		t,
		dbOrder.User.ID,
		response.UserID,
	)

	assert.Equal(
		t,
		dbOrder.User.Name,
		response.UserName,
	)

	assert.Equal(
		t,
		dbOrder.User.Email,
		response.UserEmail,
	)

	assert.Equal(
		t,
		dbOrder.User.PhoneNumber,
		response.UserPhoneNumber,
	)

	assert.False(
		t,
		response.CreatedAt.IsZero(),
	)

	assert.False(
		t,
		response.UpdatedAt.IsZero(),
	)

	// -------------------------
	// Order Items
	// -------------------------

	assert.Len(
		t,
		response.Items,
		1,
	)

	expectedItem := dbOrder.OrderItems[0]
	actualItem := response.Items[0]

	assert.Equal(
		t,
		expectedItem.BookID,
		actualItem.BookID,
	)

	assert.Equal(
		t,
		expectedItem.Quantity,
		actualItem.Quantity,
	)

	assert.Equal(
		t,
		expectedItem.Price,
		actualItem.Price,
	)

	assert.Equal(
		t,
		expectedItem.Title,
		actualItem.Title,
	)

	assert.Equal(
		t,
		expectedItem.Author,
		actualItem.Author,
	)

	assert.Equal(
		t,
		expectedItem.Publisher,
		actualItem.Publisher,
	)

	assert.Equal(
		t,
		expectedItem.ImagePath,
		actualItem.ImagePath,
	)
}

func TestAdminUpdateOrderStatus_NoToken(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	seeder.SeedAdmin(testDB)

	// -------------------------
	// Create user + order
	// -------------------------

	user := createTestUser(t)

	order := createUserOrder(
		t,
		user.ID,
		"pending",
		100,
		1,
	)

	// -------------------------
	// Request
	// -------------------------

	req := createUpdateOrderStatusRequest(
		"",
		order.ID,
		"confirmed",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

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

	assert.NoError(
		t,
		err,
	)

	assert.Equal(
		t,
		"authorization header is required",
		response["error"],
	)
}

func TestAdminUpdateOrderStatus_NotAdmin(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	seeder.SeedAdmin(testDB)

	// -------------------------
	// Create user
	// -------------------------

	user := createTestUser(t)

	token, err :=
		utils.GenerateAccessToken(
			user.ID,
			user.Role,
		)

	assert.NoError(
		t,
		err,
	)

	// -------------------------
	// Create order
	// -------------------------

	order := createUserOrder(
		t,
		user.ID,
		"pending",
		100,
		1,
	)

	// -------------------------
	// Request
	// -------------------------

	req := createUpdateOrderStatusRequest(
		token,
		order.ID,
		"confirmed",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

	assert.Equal(
		t,
		http.StatusForbidden,
		recorder.Code,
	)

	var response map[string]string

	err = json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(
		t,
		err,
	)

	assert.Equal(
		t,
		"forbidden",
		response["error"],
	)
}

func TestAdminUpdateOrderStatus_OrderNotFound(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	seeder.SeedAdmin(testDB)

	// -------------------------
	// Admin token
	// -------------------------

	token := getAdminToken(t)

	// -------------------------
	// Request
	// -------------------------

	req := createUpdateOrderStatusRequest(
		token,
		999,
		"confirmed",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

	assert.Equal(
		t,
		http.StatusNotFound,
		recorder.Code,
	)

	var response map[string]string

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(
		t,
		err,
	)

	assert.Equal(
		t,
		"order not found",
		response["error"],
	)
}

func TestAdminUpdateOrderStatus_InvalidStatus(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	seeder.SeedAdmin(testDB)

	// -------------------------
	// Admin token
	// -------------------------

	token := getAdminToken(t)

	// -------------------------
	// Create user + order
	// -------------------------

	user := createTestUser(t)

	order := createUserOrder(
		t,
		user.ID,
		"pending",
		100,
		1,
	)

	// -------------------------
	// Request
	// -------------------------

	req := createUpdateOrderStatusRequest(
		token,
		order.ID,
		"something_invalid",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

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

	assert.NoError(
		t,
		err,
	)

	assert.Equal(
		t,
		"invalid status",
		response["error"],
	)
}

func TestAdminUpdateOrderStatus_AlreadyCancelled(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	seeder.SeedAdmin(testDB)

	// -------------------------
	// Admin token
	// -------------------------

	token := getAdminToken(t)

	// -------------------------
	// Create user + cancelled order
	// -------------------------

	user := createTestUser(t)

	order := createUserOrder(
		t,
		user.ID,
		"cancelled",
		100,
		1,
	)

	// -------------------------
	// Request
	// -------------------------

	req := createUpdateOrderStatusRequest(
		token,
		order.ID,
		"confirmed",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

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

	assert.NoError(
		t,
		err,
	)

	assert.Equal(
		t,
		"order is already cancelled",
		response["error"],
	)
}

func TestAdminUpdateOrderStatus_SameStatus(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	seeder.SeedAdmin(testDB)

	// -------------------------
	// Admin token
	// -------------------------

	token := getAdminToken(t)

	// -------------------------
	// Create user + order
	// -------------------------

	user := createTestUser(t)

	order := createUserOrder(
		t,
		user.ID,
		"confirmed",
		100,
		1,
	)

	// -------------------------
	// Request
	// -------------------------

	req := createUpdateOrderStatusRequest(
		token,
		order.ID,
		"confirmed",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

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

	assert.NoError(
		t,
		err,
	)

	assert.Equal(
		t,
		"order already has this status",
		response["error"],
	)
}

func TestAdminUpdateOrderStatus_CancelledUpdatesStock(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	seeder.SeedAdmin(testDB)

	// -------------------------
	// Admin token
	// -------------------------

	token := getAdminToken(t)

	// -------------------------
	// Create user
	// -------------------------

	user := createTestUser(t)

	// -------------------------
	// Create book
	// -------------------------

	book := models.Book{
		Title:     "Clean Code",
		Author:    "Robert Martin",
		Publisher: "Prentice Hall",
		Category:  "Programming",
		Price:     100,
		Stock:     5,
	}

	err := testDB.Create(&book).Error

	assert.NoError(
		t,
		err,
	)

	// -------------------------
	// Create order
	// -------------------------

	order := models.Order{
		UserID:          user.ID,
		Status:          "confirmed",
		ShippingAddress: "Cairo",
		TotalPrice:      300,
	}

	err = testDB.Create(&order).Error

	assert.NoError(
		t,
		err,
	)

	// -------------------------
	// Create order item
	// -------------------------

	orderItem := models.OrderItem{
		OrderID:   order.ID,
		BookID:    book.ID,
		Quantity:  3,
		Price:     book.Price,
		Title:     book.Title,
		Author:    book.Author,
		Publisher: book.Publisher,
	}

	err = testDB.Create(&orderItem).Error

	assert.NoError(
		t,
		err,
	)

	// -------------------------
	// Simulate stock deduction
	// -------------------------

	book.Stock = 2

	err = testDB.Save(&book).Error

	assert.NoError(
		t,
		err,
	)

	// -------------------------
	// Request
	// -------------------------

	req := createUpdateOrderStatusRequest(
		token,
		order.ID,
		"cancelled",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Response
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	var response dto.MessageResponse

	err = json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(
		t,
		err,
	)

	assert.Equal(
		t,
		"order status updated successfully",
		response.Message,
	)

	// -------------------------
	// Verify order status
	// -------------------------

	var updatedOrder models.Order

	err = testDB.
		First(
			&updatedOrder,
			order.ID,
		).
		Error

	assert.NoError(
		t,
		err,
	)

	assert.Equal(
		t,
		"cancelled",
		updatedOrder.Status,
	)

	// -------------------------
	// Verify stock restored
	// -------------------------

	var updatedBook models.Book

	err = testDB.
		First(
			&updatedBook,
			book.ID,
		).
		Error

	assert.NoError(
		t,
		err,
	)

	assert.Equal(
		t,
		5,
		updatedBook.Stock,
	)
}

func TestAdminUpdateOrderStatus_Success(
	t *testing.T,
) {

	// -------------------------
	// Clean database
	// -------------------------

	cleanDatabase()

	seeder.SeedAdmin(testDB)

	// -------------------------
	// Admin token
	// -------------------------

	token := getAdminToken(t)

	// -------------------------
	// Create user + order
	// -------------------------

	user := createTestUser(t)

	order := createUserOrder(
		t,
		user.ID,
		"pending",
		100,
		1,
	)

	// -------------------------
	// Request
	// -------------------------

	req := createUpdateOrderStatusRequest(
		token,
		order.ID,
		"confirmed",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	// -------------------------
	// Assertions
	// -------------------------

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	var response dto.MessageResponse

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	assert.NoError(
		t,
		err,
	)

	assert.Equal(
		t,
		"order status updated successfully",
		response.Message,
	)

	// -------------------------
	// Verify database
	// -------------------------

	var updatedOrder models.Order

	err = testDB.
		First(
			&updatedOrder,
			order.ID,
		).
		Error

	assert.NoError(
		t,
		err,
	)

	assert.Equal(
		t,
		"confirmed",
		updatedOrder.Status,
	)
}