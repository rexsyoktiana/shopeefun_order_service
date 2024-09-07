package order

import (
	model "cart-order-service/repository/models"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type store struct {
	db     *sql.DB
	logger zerolog.Logger
}

func NewStore(db *sql.DB, logger zerolog.Logger) *store {
	return &store{
		db:     db,
		logger: logger,
	}
}

// CreateOrder is a method that creates a new order and returns the order ID.
// It returns an error if any occurs during the creation process.
func (o *store) CreateOrder(bReq model.Order) (*uuid.UUID, *string, error) {
	logMsgStr := "Repository:Order - CreateOrder:"

	tx, err := o.db.Begin()
	if err != nil {
		o.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to Begin tx", logMsgStr))
		return nil, nil, err
	}

	queryCreate := `
		INSERT INTO orders (
			user_id,
			payment_type_id,
			order_number,
			total_price,
			product_order,
			status,
			is_paid,
			ref_code,
			created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, NOW()
		) RETURNING id, ref_code
	`

	var orderID uuid.UUID
	var refCode string
	if err := tx.QueryRow(
		queryCreate,
		bReq.UserID,
		bReq.PaymentTypeID,
		bReq.OrderNumber,
		bReq.TotalPrice,
		bReq.ProductOrder,
		bReq.Status,
		bReq.IsPaid,
		bReq.RefCode,
	).Scan(&orderID, &refCode); err != nil {
		tx.Rollback()
		o.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to Scan orderId, refCode", logMsgStr))
		return nil, nil, err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		o.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to Commit tx", logMsgStr))
		return nil, nil, err
	}

	return &orderID, &refCode, nil
}

// createOrderItemsLogs is a method that creates a new order items log.
// It returns an error if any occurs during the creation process.
func (o *store) CreateOrderItemsLogs(bReq model.OrderItemsLogs) (*string, error) {
	logMsgStr := "Repository:Order - CreateOrderItemsLogs:"

	tx, err := o.db.Begin()
	if err != nil {
		o.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to Begin tx", logMsgStr))
		return nil, err
	}

	queryCreate := `
		INSERT INTO order_status_logs (
			order_id,
			ref_code,
			from_status,
			to_status,
			notes,
			created_at
		) VALUES (
			$1, $2, $3, $4, $5, NOW()
		) RETURNING ref_code
	`

	var refCode string
	if err := tx.QueryRow(
		queryCreate,
		bReq.OrderID,
		bReq.RefCode,
		bReq.FromStatus,
		bReq.ToStatus,
		bReq.Notes,
	).Scan(&refCode); err != nil {
		tx.Rollback()
		o.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to Scan refCode", logMsgStr))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		o.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to Commit tx", logMsgStr))
		return nil, err
	}

	return &refCode, nil
}
