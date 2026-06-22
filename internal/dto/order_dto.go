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
