package order

import (
	"cart-order-service/helper"
	model "cart-order-service/repository/models"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type orderDto interface {
	CreateOrder(bReq model.Order) (*uuid.UUID, error)
}

type Handler struct {
	order     orderDto
	validator *validator.Validate
	logger    zerolog.Logger
}

func NewHandler(order orderDto, validator *validator.Validate, logger zerolog.Logger) *Handler {
	return &Handler{
		order:     order,
		validator: validator,
		logger:    logger,
	}
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	logMsgStr := "Handler:Order - CreateOrder:"

	var bReq model.Order
	if err := helper.ParseRequestBody(r, &bReq, h.logger); err != nil {
		h.logger.Error().Any("Err", err.Error()).Msg(fmt.Sprintf("%v failed to decode request body", logMsgStr))
		helper.HandleResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	bReq.RefCode = helper.GenerateRefCode()

	if bReq.ProductOrder == nil {
		bReq.ProductOrder = json.RawMessage("[]")
	}

	if err := h.validator.Struct(bReq); err != nil {
		h.logger.Error().AnErr("Err", err).Msg(fmt.Sprintf("%v failed to validate request body", logMsgStr))
		helper.HandleResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	bRes, err := h.order.CreateOrder(bReq)
	if err != nil {
		h.logger.Error().AnErr("Err", err).Msg(fmt.Sprintf("%v failed to create order", logMsgStr))
		helper.HandleResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	helper.HandleResponse(w, http.StatusCreated, bRes)
}
