package dto

import "time"

type CreateOrderItemRequest struct {
	BookID uint `json:"book_id" binding:"required"`

	Quantity int `json:"quantity" binding:"required"`
}

type CreateOrderRequest struct {
	Address string `json:"address" binding:"required"`

	Items []CreateOrderItemRequest `json:"items" binding:"required"`
}

type OrderItemResponse struct {
	BookID uint `json:"book_id"`

	Quantity int `json:"quantity"`

	Price float64 `json:"price"`

	Title string `json:"title"`

	Author string `json:"author"`

	Publisher string `json:"publisher"`

	ImageURL string `json:"image_url"`
}

type CreateOrderResponse struct {
	ID uint `json:"id"`

	Status string `json:"status"`

	Address string `json:"address"`

	TotalPrice float64 `json:"total_price"`

	Items []OrderItemResponse `json:"items"`

	CreatedAt time.Time `json:"created_at"`
}

type GetMyOrdersQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

type UserOrderListItemResponse struct {
	ID uint `json:"id"`

	Status string `json:"status"`

	TotalPrice float64 `json:"total_price"`

	ItemsCount int `json:"items_count"`

	CreatedAt time.Time `json:"created_at"`
}

type UserOrdersResponse struct {
	Orders []UserOrderListItemResponse `json:"orders"`

	Page int `json:"page"`

	PageSize int `json:"page_size"`

	Total int64 `json:"total"`
}

type UserOrderDetailsItemResponse struct {
	ID uint `json:"id"`

	BookID uint `json:"book_id"`

	Quantity int `json:"quantity"`

	Price float64 `json:"price"`

	Title string `json:"title"`

	Author string `json:"author"`

	Publisher string `json:"publisher"`

	ImagePath string `json:"image_path"`
}

type UserOrderDetailsResponse struct {
	ID uint `json:"id"`

	Status string `json:"status"`

	Address string `json:"address"`

	TotalPrice float64 `json:"total_price"`

	CreatedAt time.Time `json:"created_at"`

	UpdatedAt time.Time `json:"updated_at"`

	Items []UserOrderDetailsItemResponse `json:"items"`
}

type AdminGetOrdersQuery struct {
	UserName string `form:"user_name"`

	Status string `form:"status"`

	Page int `form:"page"`

	PageSize int `form:"page_size"`
}

type AdminOrderListItemResponse struct {
	ID uint `json:"id"`

	UserName string `json:"user_name"`

	Status string `json:"status"`

	Address string `json:"address"`

	TotalPrice float64 `json:"total_price"`

	ItemsCount int `json:"items_count"`

	CreatedAt time.Time `json:"created_at"`
}

type AdminOrdersResponse struct {
	Orders []AdminOrderListItemResponse `json:"orders"`

	Page int `json:"page"`

	PageSize int `json:"page_size"`

	Total int64 `json:"total"`
}

type AdminOrderItemResponse struct {
	BookID uint `json:"book_id"`

	Quantity int `json:"quantity"`

	Price float64 `json:"price"`

	Title string `json:"title"`

	Author string `json:"author"`

	Publisher string `json:"publisher"`

	ImagePath string `json:"image_path"`
}

type AdminOrderResponse struct {
	ID uint `json:"id"`

	Status string `json:"status"`

	Address string `json:"address"`

	TotalPrice float64 `json:"total_price"`

	UserID uint `json:"user_id"`

	UserName string `json:"user_name"`

	UserEmail string `json:"user_email"`

	UserPhoneNumber string `json:"user_phone_number"`

	Items []AdminOrderItemResponse `json:"items"`

	CreatedAt time.Time `json:"created_at"`

	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}