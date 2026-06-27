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

func (s *OrderService) GetUserOrders(
	userID uint,
	query dto.GetMyOrdersQuery,
) (
	*dto.UserOrdersResponse,
	error,
) {

	// -------------------------
	// Default pagination
	// -------------------------

	if query.Page <= 0 {
		query.Page = 1
	}

	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// -------------------------
	// Query database
	// -------------------------

	orders,
		total,
		err :=
		s.orderRepo.GetUserOrders(
			userID,
			query.Page,
			query.PageSize,
		)

	if err != nil {
		return nil, err
	}

	// -------------------------
	// Build response
	// -------------------------

	responseOrders :=
		make(
			[]dto.UserOrderListItemResponse,
			0,
			len(orders),
		)

	for _, order := range orders {

		responseOrders =
			append(
				responseOrders,
				dto.UserOrderListItemResponse{
					ID: order.ID,

					Status: order.Status,

					TotalPrice: order.TotalPrice,

					ItemsCount: calculateTotalItems(order.OrderItems),

					CreatedAt: order.CreatedAt,
				},
			)
	}

	response :=
		dto.UserOrdersResponse{
			Orders: responseOrders,

			Page: query.Page,

			PageSize: query.PageSize,

			Total: total,
		}

	return &response, nil
}

func calculateTotalItems(orderItems []models.OrderItem) int {
	count := 0
	for _, orderItem := range orderItems {
		count = count + orderItem.Quantity
	}
	return count
}

func (s *OrderService) GetUserOrderByID(
	orderID uint,
	userID uint,
) (
	*dto.UserOrderDetailsResponse,
	error,
) {

	// -------------------------
	// Query Order
	// -------------------------

	order, err :=
		s.orderRepo.GetUserOrderByID(
			orderID,
			userID,
		)

	if err != nil {

		if errors.Is(
			err,
			gorm.ErrRecordNotFound,
		) {

			return nil,
				errors.New(
					"order not found",
				)
		}

		return nil, err
	}

	// -------------------------
	// Build Items
	// -------------------------

	items :=
		make(
			[]dto.UserOrderDetailsItemResponse,
			0,
			len(order.OrderItems),
		)

	for _, item := range order.OrderItems {

		items =
			append(
				items,
				dto.UserOrderDetailsItemResponse{
					ID: item.ID,

					BookID: item.BookID,

					Quantity: item.Quantity,

					Price: item.Price,

					Title: item.Title,

					Author: item.Author,

					Publisher: item.Publisher,

					ImagePath: item.ImagePath,
				},
			)
	}

	// -------------------------
	// Response
	// -------------------------

	response :=
		dto.UserOrderDetailsResponse{
			ID: order.ID,

			Status: order.Status,

			Address: order.ShippingAddress,

			TotalPrice: order.TotalPrice,

			CreatedAt: order.CreatedAt,

			UpdatedAt: order.UpdatedAt,

			Items: items,
		}

	return &response, nil
}

func (s *OrderService) GetAllOrders(
	query dto.AdminGetOrdersQuery,
) (
	*dto.AdminOrdersResponse,
	error,
) {

	// -------------------------
	// Default Pagination
	// -------------------------

	if query.Page <= 0 {

		query.Page = 1
	}

	if query.PageSize <= 0 {

		query.PageSize = 10
	}

	if query.PageSize > 100 {

		query.PageSize = 100
	}

	// -------------------------
	// Repository
	// -------------------------

	orders,
		total,
		err :=
		s.orderRepo.GetAllOrders(
			query,
		)

	if err != nil {

		return nil,
			err
	}

	// -------------------------
	// Build Response
	// -------------------------

	responseOrders :=
		make(
			[]dto.AdminOrderListItemResponse,
			0,
			len(orders),
		)

	for _, order := range orders {

		responseOrders =
			append(
				responseOrders,
				dto.AdminOrderListItemResponse{
					ID: order.ID,

					UserName: order.User.Name,

					Status: order.Status,

					Address: order.ShippingAddress,

					TotalPrice: order.TotalPrice,

					ItemsCount: calculateTotalItems(
						order.OrderItems,
					),

					CreatedAt: order.CreatedAt,
				},
			)
	}

	response :=
		dto.AdminOrdersResponse{
			Orders: responseOrders,

			Page: query.Page,

			PageSize: query.PageSize,

			Total: total,
		}

	return &response,
		nil
}

func (s *OrderService) GetOrder(
	orderID uint,
) (
	*dto.AdminOrderResponse,
	error,
) {

	// -------------------------
	// Get Order
	// -------------------------

	order,
		err :=
		s.orderRepo.GetByID(
			orderID,
		)

	if err != nil {

		return nil,
			err
	}

	// -------------------------
	// Build Items
	// -------------------------

	items :=
		make(
			[]dto.AdminOrderItemResponse,
			0,
			len(order.OrderItems),
		)

	for _, item := range order.OrderItems {

		items =
			append(
				items,
				dto.AdminOrderItemResponse{
					BookID: item.BookID,

					Quantity: item.Quantity,

					Price: item.Price,

					Title: item.Title,

					Author: item.Author,

					Publisher: item.Publisher,

					ImagePath: item.ImagePath,
				},
			)
	}

	// -------------------------
	// Build Response
	// -------------------------

	response :=
		dto.AdminOrderResponse{
			ID: order.ID,

			Status: order.Status,

			Address: order.ShippingAddress,

			TotalPrice: order.TotalPrice,

			UserID: order.User.ID,

			UserName: order.User.Name,

			UserEmail: order.User.Email,

			UserPhoneNumber: order.User.PhoneNumber,

			Items: items,

			CreatedAt: order.CreatedAt,

			UpdatedAt: order.UpdatedAt,
		}

	return &response,
		nil
}

func (s *OrderService) UpdateOrderStatus(
	orderID uint,
	request dto.UpdateOrderStatusRequest,
) error {

	// -------------------------
	// Allowed Statuses
	// -------------------------

	allowedStatuses :=
		map[string]bool{
			"pending":          true,
			"confirmed":        true,
			"out_for_delivery": true,
			"delivered":        true,
			"cancelled":        true,
		}

	// -------------------------
	// Validate Status
	// -------------------------

	if !allowedStatuses[
		request.Status,
	] {

		return errors.New(
			"invalid status",
		)
	}

	// -------------------------
	// Get Order
	// -------------------------

	order,
		err :=
		s.orderRepo.GetByID(
			orderID,
		)

	if err != nil {

		return err
	}

	// -------------------------
	// Cannot change cancelled
	// -------------------------

	if order.Status ==
		"cancelled" {

		return errors.New(
			"order is already cancelled",
		)
	}

	// -------------------------
	// Same status
	// -------------------------

	if order.Status ==
		request.Status {

		return errors.New(
			"order already has this status",
		)
	}

	// -------------------------
	// Transaction
	// -------------------------

	tx :=
		s.db.Begin()

	if tx.Error != nil {

		return tx.Error
	}

	defer func() {

		if r := recover(); r != nil {

			tx.Rollback()
		}
	}()

	// -------------------------
	// Restore stock
	// if new status cancelled
	// -------------------------

	if request.Status ==
		"cancelled" {

		for _, item :=
			range order.OrderItems {

			book,
				err :=
				s.bookRepo.GetByID(
					item.BookID,
				)

			if err != nil {

				tx.Rollback()

				return err
			}

			book.Stock +=
				item.Quantity

			err =
				s.bookRepo.Update(
					tx,
					book,
				)

			if err != nil {

				tx.Rollback()

				return err
			}
		}
	}

	// -------------------------
	// Update Status
	// -------------------------

	order.Status =
		request.Status

	err =
		s.orderRepo.Save(
			tx,
			order,
		)

	if err != nil {

		tx.Rollback()

		return err
	}

	// -------------------------
	// Commit
	// -------------------------

	err =
		tx.Commit().
			Error

	if err != nil {

		return err
	}

	return nil
}