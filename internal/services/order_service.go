package services

import (
	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/models"
	"bookstore-backend/internal/repositories"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type OrderService struct {
	db *gorm.DB

	bookRepo *repositories.BookRepository

	orderRepo *repositories.OrderRepository

	orderItemRepo *repositories.OrderItemRepository
}

func NewOrderService(
	db *gorm.DB,
) *OrderService {

	return &OrderService{
		db: db,

		bookRepo: repositories.NewBookRepository(db),

		orderRepo: repositories.NewOrderRepository(db),

		orderItemRepo: repositories.NewOrderItemRepository(db),
	}
}

func (s *OrderService) CreateOrder(
	userID uint,
	request dto.CreateOrderRequest,
) (
	dto.CreateOrderResponse,
	error,
) {

	// -------------------------
	// Validate Address
	// -------------------------

	if strings.TrimSpace(
		request.Address,
	) == "" {

		return dto.CreateOrderResponse{},
			errors.New(
				"address is required",
			)
	}

	// -------------------------
	// Validate Items
	// -------------------------

	if len(request.Items) == 0 {

		return dto.CreateOrderResponse{},
			errors.New(
				"order must contain at least one book",
			)
	}

	// -------------------------
	// Validate Quantities
	// -------------------------

	bookIDs :=
		make(
			[]uint,
			0,
			len(request.Items),
		)

	seen :=
		make(map[uint]bool)

	for _, item := range request.Items {

		if item.Quantity <= 0 {

			return dto.CreateOrderResponse{},
				errors.New(
					"quantity must be greater than 0",
				)
		}

		if seen[item.BookID] {

			return dto.CreateOrderResponse{},
				errors.New(
					"duplicate books are not allowed",
				)
		}

		seen[item.BookID] = true

		bookIDs = append(
			bookIDs,
			item.BookID,
		)
	}

	// -------------------------
	// Get Books
	// -------------------------

	books, err :=
		s.bookRepo.GetByIDs(
			bookIDs,
		)

	if err != nil {

		return dto.CreateOrderResponse{},
			err
	}

	// Ensure all books exist

	if len(books) != len(bookIDs) {

		return dto.CreateOrderResponse{},
			errors.New(
				"book not found",
			)
	}

	bookMap :=
		make(
			map[uint]models.Book,
		)

	for _, book := range books {

		bookMap[book.ID] = book
	}

	// -------------------------
	// Validate Stock
	// Calculate Total
	// Build Order Items
	// -------------------------

	totalPrice := 0.0

	orderItems :=
		make(
			[]models.OrderItem,
			0,
			len(request.Items),
		)

	responseItems :=
		make(
			[]dto.OrderItemResponse,
			0,
			len(request.Items),
		)

	for _, item := range request.Items {

		book :=
			bookMap[item.BookID]

		if item.Quantity >
			book.Stock {

			return dto.CreateOrderResponse{},
				errors.New(
					"not enough stock",
				)
		}

		itemPrice :=
			book.Price *
				float64(
					item.Quantity,
				)

		totalPrice += itemPrice

		orderItems =
			append(
				orderItems,
				models.OrderItem{
					BookID: item.BookID,

					Quantity: item.Quantity,

					Price: book.Price,

					Title: book.Title,

					Author: book.Author,

					Publisher: book.Publisher,

					ImagePath: book.ImagePath,
				},
			)

		responseItems =
			append(
				responseItems,
				dto.OrderItemResponse{
					BookID: item.BookID,

					Quantity: item.Quantity,

					Price: book.Price,

					Title: book.Title,

					Author: book.Author,

					Publisher: book.Publisher,

					ImageURL: book.ImagePath,
				},
			)
	}

	// -------------------------
	// Build Order
	// -------------------------

	order :=
		models.Order{
			UserID: userID,

			Status: "pending",

			ShippingAddress: request.Address,

			TotalPrice: totalPrice,
		}

	tx :=
		s.db.Begin()

	if tx.Error != nil {

		return dto.CreateOrderResponse{},
			tx.Error
	}

	// Rollback on panic

	defer func() {

		if r := recover(); r != nil {

			tx.Rollback()
		}
	}()

	// -------------------------
	// Create Order
	// -------------------------

	err =
		s.orderRepo.Create(
			tx,
			&order,
		)

	if err != nil {

		tx.Rollback()

		return dto.CreateOrderResponse{},
			err
	}

	// -------------------------
	// Assign OrderID
	// -------------------------

	for i := range orderItems {

		orderItems[i].OrderID =
			order.ID
	}

	// -------------------------
	// Create Order Items
	// -------------------------

	err =
		s.orderItemRepo.CreateMany(
			tx,
			orderItems,
		)

	if err != nil {

		tx.Rollback()

		return dto.CreateOrderResponse{},
			err
	}

	// -------------------------
	// Update Stock
	// -------------------------

	for _, item := range request.Items {

		book :=
			bookMap[item.BookID]

		book.Stock -=
			item.Quantity

		err =
			s.bookRepo.Update(
				tx,
				&book,
			)

		if err != nil {

			tx.Rollback()

			return dto.CreateOrderResponse{},
				err
		}
	}

	// -------------------------
	// Commit
	// -------------------------

	err = tx.Commit().Error

	if err != nil {

		return dto.CreateOrderResponse{},
			err
	}

	return dto.CreateOrderResponse{
		ID: order.ID,

		Status: order.Status,

		Address: order.ShippingAddress,

		TotalPrice: order.TotalPrice,

		Items: responseItems,

		CreatedAt: order.CreatedAt,
	}, nil
}
