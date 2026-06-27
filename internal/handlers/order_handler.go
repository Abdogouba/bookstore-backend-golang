package handlers

import (
	"bookstore-backend/internal/dto"
	"bookstore-backend/internal/services"
	"bookstore-backend/internal/utils"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler(
	db *gorm.DB,
) *OrderHandler {

	return &OrderHandler{
		orderService: services.NewOrderService(db),
	}
}

// CreateOrder godoc
//
//	@Summary		Create order
//	@Description	Create a new order for the authenticated user
//	@Tags			Orders
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateOrderRequest	true	"Order Request"
//	@Success		201		{object}	dto.CreateOrderResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/orders [post]
func (h *OrderHandler) CreateOrder(
	c *gin.Context,
) {

	// -------------------------
	// Bind Request
	// -------------------------

	var request dto.CreateOrderRequest

	if err :=
		c.ShouldBindJSON(
			&request,
		); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	// -------------------------
	// Get User ID
	// -------------------------

	userIDValue, exists :=
		c.Get("userID")

	if !exists {

		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "unauthorized",
			},
		)

		return
	}

	userID :=
		userIDValue.(uint)

	// -------------------------
	// Create Order
	// -------------------------

	response, err :=
		h.orderService.CreateOrder(
			userID,
			request,
		)

	if err != nil {

		switch err.Error() {

		case "address is required",
			"order must contain at least one book",
			"quantity must be greater than 0",
			"duplicate books are not allowed",
			"not enough stock":

			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": err.Error(),
				},
			)

			return

		case "book not found":

			c.JSON(
				http.StatusNotFound,
				gin.H{
					"error": err.Error(),
				},
			)

			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)

		return
	}

	// -------------------------
	// Success
	// -------------------------

	c.JSON(
		http.StatusCreated,
		response,
	)
}

// GetMyOrders godoc
//
// @Summary View my orders
// @Description Returns paginated orders of the authenticated user
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.UserOrdersResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders [get]
func (h *OrderHandler) GetMyOrders(
	c *gin.Context,
) {

	// -------------------------
	// Query params
	// -------------------------

	var query dto.GetMyOrdersQuery

	if err :=
		c.ShouldBindQuery(
			&query,
		); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	// -------------------------
	// User ID
	// -------------------------

	userID :=
		utils.GetUserID(c)

	// -------------------------
	// Service
	// -------------------------

	response, err :=
		h.orderService.GetUserOrders(
			userID,
			query,
		)

	if err != nil {

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)

		return
	}

	// -------------------------
	// Success
	// -------------------------

	c.JSON(
		http.StatusOK,
		response,
	)
}

// GetMyOrder godoc
//
// @Summary View one of my orders
// @Description Returns a single order belonging to the authenticated user
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} dto.UserOrderDetailsResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders/{id} [get]
func (h *OrderHandler) GetMyOrder(
	c *gin.Context,
) {

	// -------------------------
	// Parse Order ID
	// -------------------------

	orderID, err :=
		strconv.ParseUint(
			c.Param("id"),
			10,
			64,
		)

	if err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid order id",
			},
		)

		return
	}

	// -------------------------
	// User ID
	// -------------------------

	userID :=
		utils.GetUserID(c)

	// -------------------------
	// Service
	// -------------------------

	response, err :=
		h.orderService.GetUserOrderByID(
			uint(orderID),
			userID,
		)

	if err != nil {

		if err.Error() ==
			"order not found" {

			c.JSON(
				http.StatusNotFound,
				gin.H{
					"error": err.Error(),
				},
			)

			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)

		return
	}

	// -------------------------
	// Success
	// -------------------------

	c.JSON(
		http.StatusOK,
		response,
	)
}

// GetAllOrders godoc
//
// @Summary View all orders
// @Description Returns paginated orders for the admin
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param user_name query string false "Search by user name"
// @Param status query string false "Filter by status"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} dto.AdminOrdersResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/orders [get]
func (h *OrderHandler) GetAllOrders(
	c *gin.Context,
) {

	// -------------------------
	// Query Params
	// -------------------------

	var query dto.AdminGetOrdersQuery

	if err :=
		c.ShouldBindQuery(
			&query,
		); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": err.Error(),
			},
		)

		return
	}

	// -------------------------
	// Service
	// -------------------------

	response, err :=
		h.orderService.GetAllOrders(
			query,
		)

	if err != nil {

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)

		return
	}

	// -------------------------
	// Success
	// -------------------------

	c.JSON(
		http.StatusOK,
		response,
	)
}

// GetOrder godoc
//
// @Summary View one order
// @Description Returns a single order for the admin
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} dto.AdminOrderResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/orders/{id} [get]
func (h *OrderHandler) GetOrder(
	c *gin.Context,
) {

	// -------------------------
	// Parse Order ID
	// -------------------------

	orderID, err :=
		strconv.ParseUint(
			c.Param("id"),
			10,
			64,
		)

	if err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid order id",
			},
		)

		return
	}

	// -------------------------
	// Service
	// -------------------------

	response, err :=
		h.orderService.GetOrder(
			uint(orderID),
		)

	if err != nil {

		if errors.Is(
			err,
			gorm.ErrRecordNotFound,
		) {

			c.JSON(
				http.StatusNotFound,
				gin.H{
					"error": "order not found",
				},
			)

			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)

		return
	}

	// -------------------------
	// Success
	// -------------------------

	c.JSON(
		http.StatusOK,
		response,
	)
}
