package cart

import (
	model "cart-order-service/repository/models"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type store struct {
	db     *sql.DB
	logger zerolog.Logger
}

// NewStore is a constructor function that returns a new store instance.
func NewStore(db *sql.DB, logger zerolog.Logger) *store {
	return &store{
		db:     db,
		logger: logger,
	}
}

// GetCartByUserID is a method that retrieves the cart for a given user.
// It returns a slice of cart and an error if any occurs during the retrieval process.
func (s *store) GetCartByUserID(bReq model.GetCartRequest) (*[]model.Cart, error) {
	logMsgStr := "Repository:Cart - GetCartByUserID:"

	querySelect := `
		SELECT
			*
		FROM cart_items
		WHERE deleted_at IS NULL
	`

	var queryConditions []string

	if bReq.UserID != uuid.Nil {
		queryConditions = append(queryConditions, fmt.Sprintf("user_id = '%s'", bReq.UserID))
	}

	if len(bReq.ProductID) > 0 {
		var productIDs []string
		for _, pid := range bReq.ProductID {
			productIDs = append(productIDs, fmt.Sprintf("'%s'", pid))
		}
		queryConditions = append(queryConditions, fmt.Sprintf("product_id IN (%s)", strings.Join(productIDs, ",")))
	}

	if len(queryConditions) > 0 {
		querySelect += " AND " + strings.Join(queryConditions, " AND ")
	}

	rows, err := s.db.Query(querySelect)
	if err != nil {
		s.logger.Error().Any("Err", err.Error()).Msg(fmt.Sprintf("%v Failed to Query querySelect", logMsgStr))
		return nil, err
	}
	defer rows.Close()

	var carts []model.Cart
	for rows.Next() {
		var cart model.Cart
		if err := rows.Scan(
			&cart.ID,
			&cart.UserID,
			&cart.ProductID,
			&cart.Qty,
			&cart.CreatedAt,
			&cart.UpdatedAt,
			&cart.DeletedAt,
		); err != nil {
			s.logger.Error().Any("Err", err.Error()).Msg(fmt.Sprintf("%v Failed to rows.Scan", logMsgStr))
			return nil, err
		}
		carts = append(carts, cart)
	}

	if err := rows.Err(); err != nil {
		s.logger.Error().Any("Err", err.Error()).Msg(fmt.Sprintf("%v Failed to rows.Err", logMsgStr))
		return nil, err
	}

	return &carts, nil
}

func (s *store) AddCart(bReq model.Cart) (*uuid.UUID, error) {
	logMsgStr := "Repository:Cart - AddCart:"

	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to Begin tx", logMsgStr))
		return nil, err
	}

	var id uuid.UUID
	queryCreate := `
		INSERT INTO cart_items (
			user_id,
			product_id,
			qty,
			created_at
		) VALUES (
			$1,
			$2,
			$3,
			NOW()
		) RETURNING id
	`
	if err := tx.QueryRow(
		queryCreate,
		bReq.UserID,
		bReq.ProductID,
		bReq.Qty,
	).Scan(&id); err != nil {
		tx.Rollback()
		s.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to scan id", logMsgStr))
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		s.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to commit", logMsgStr))
		return nil, err
	}

	return &id, nil
}

func (s *store) UpdateQty(userID, productID uuid.UUID, qty int) error {
	logMsgStr := "Repository:Cart - UpdateQty:"

	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to Begin tx", logMsgStr))
		return err
	}

	queryLock := `
		SELECT 1
		FROM cart_items
		WHERE user_id = $1
		FOR UPDATE
	`
	if _, err := tx.Exec(queryLock, userID); err != nil {
		tx.Rollback()
		s.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to lock data", logMsgStr))
		return errors.New("failed to lock data")
	}

	queryUpdate := `
		UPDATE cart_items
		SET qty = $1
		WHERE deleted_at IS NULL AND user_id = $2 AND product_id = $3
	`
	result, err := tx.Exec(queryUpdate, qty, userID, productID)
	if err != nil {
		tx.Rollback()
		s.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to update data", logMsgStr))
		return errors.New("failed to update data")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		s.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to get rows affected", logMsgStr))
		return errors.New("failed to get rows affected")
	}

	if rowsAffected == 0 {
		tx.Rollback()
		s.logger.Warn().Msg(fmt.Sprintf("%v No rows affected", logMsgStr))
		return errors.New("no rows affected")
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		s.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to commit transaction", logMsgStr))
		return err
	}

	return nil
}

func (s *store) DeleteProduct(bReq model.DeleteCartRequest) error {
	logMsgStr := "Repository:Cart - DeleteProduct:"

	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to Begin tx", logMsgStr))
		return err
	}

	queryLock := `
		SELECT 1
		FROM cart_items	
		WHERE user_id = $1
		FOR UPDATE
	`
	if _, err := tx.Exec(queryLock, bReq.UserID); err != nil {
		tx.Rollback()
		s.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to lock data", logMsgStr))
		return errors.New("failed to lock data")
	}

	queryUpdate := `
		UPDATE cart_items
		SET deleted_at = NOW()
		WHERE deleted_at IS NULL AND user_id = $1 AND product_id = $2
	`
	result, err := tx.Exec(queryUpdate, bReq.UserID, bReq.ProductID)
	if err != nil {
		tx.Rollback()
		s.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to delete data", logMsgStr))
		return errors.New("failed to delete data")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		s.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to get rows affected", logMsgStr))
		return errors.New("failed to get rows affected")
	}

	if rowsAffected == 0 {
		tx.Rollback()
		s.logger.Warn().Msg(fmt.Sprintf("%v No rows affected", logMsgStr))
		return errors.New("no rows affected")
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		s.logger.Error().Any("Err", err).Msg(fmt.Sprintf("%v Failed to commit transaction", logMsgStr))
		return err
	}

	return nil
}
