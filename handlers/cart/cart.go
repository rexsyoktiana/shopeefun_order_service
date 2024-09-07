package cart

import (
	"cart-order-service/helper"
	model "cart-order-service/repository/models"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// cartDto is an interface that defines the methods that our Handler struct depends on.
type cartDto interface {
	GetCartByUserID(bReq model.GetCartRequest) (*[]model.Cart, error)
	AddCart(bReq model.Cart) (*uuid.UUID, error)
	UpdateQty(bReq model.Cart) (string, error)
	DeleteCart(bReq model.DeleteCartRequest) (string, error)
}

// Handler is a struct that holds a cartDto.
type Handler struct {
	cart   cartDto
	logger zerolog.Logger
}

// NewHandler is a constructor function that returns a new Handler.
func NewHandler(cart cartDto, logger zerolog.Logger) *Handler {
	return &Handler{
		cart:   cart,
		logger: logger,
	}
}

func (h *Handler) GetCartByUserID(w http.ResponseWriter, r *http.Request) {
	logMsgStr := "Handler:Cart - GetCartByUserID:"

	userID := r.PathValue("user_id")
	if userID == "" {
		h.logger.Error().Msg(fmt.Sprintf("%v User ID is required", logMsgStr))
		helper.HandleResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Error().AnErr("Err", err).Msg(fmt.Sprintf("%v error parse uuid: %v", logMsgStr, userID))
		helper.HandleResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var bReq model.GetCartRequest
	if err := helper.ParseRequestBody(r, &bReq, h.logger); err != nil {
		h.logger.Error().AnErr("Err", err).Msg(fmt.Sprintf("%v Failed to decode request body", logMsgStr))
		helper.HandleResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var pidSlice []uuid.UUID
	pidSlice = append(pidSlice, bReq.ProductID...)

	bReq = model.GetCartRequest{
		UserID:    uid,
		ProductID: pidSlice,
	}

	bResp, err := h.cart.GetCartByUserID(bReq)
	if err != nil {
		h.logger.Error().AnErr("Err", err).Msg(fmt.Sprintf("%v Failed to GetCartByUserID", logMsgStr))
		helper.HandleResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	helper.HandleResponse(w, http.StatusOK, bResp)
}

func (h *Handler) AddCart(w http.ResponseWriter, r *http.Request) {
	logMsgStr := "Handler:Cart - AddCart:"

	var bReq model.Cart
	if err := helper.ParseRequestBody(r, &bReq, h.logger); err != nil {
		h.logger.Error().AnErr("Err", err).Msg(fmt.Sprintf("%v Failed to decode request body", logMsgStr))
		helper.HandleResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if bReq.Qty == 0 {
		h.logger.Error().Msg(fmt.Sprintf("%v Qty must be greater than 0", logMsgStr))
		helper.HandleResponse(w, http.StatusBadRequest, "Qty must be greater than 0")
		return
	}

	bResp, err := h.cart.AddCart(bReq)
	if err != nil {
		h.logger.Error().AnErr("Err", err).Msg(fmt.Sprintf("%v Failed to AddCart", logMsgStr))
		helper.HandleResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	helper.HandleResponse(w, http.StatusOK, bResp)
}

func (h *Handler) UpdateCart(w http.ResponseWriter, r *http.Request) {
	logMsgStr := "Hanlder:Cart - UpdateCart:"

	userID := r.PathValue("user_id")
	if userID == "" {
		h.logger.Error().Msg(fmt.Sprintf("%v User ID is required", logMsgStr))
		helper.HandleResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Error().AnErr("Err", err).Msg(fmt.Sprintf("%v error parse uuid: %v", logMsgStr, userID))
		helper.HandleResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var bReq model.Cart
	if err := helper.ParseRequestBody(r, &bReq, h.logger); err != nil {
		h.logger.Error().Any("Err", err.Error()).Msg(fmt.Sprintf("%v failed to decode request body", logMsgStr))
		helper.HandleResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	bReq.UserID = uid

	bResp, err := h.cart.UpdateQty(bReq)
	if err != nil {
		h.logger.Error().Any("Err", err.Error()).Msg(fmt.Sprintf("%v failed to UpdateQty", logMsgStr))
		helper.HandleResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	helper.HandleResponse(w, http.StatusOK, bResp)
}

func (h *Handler) DeleteCart(w http.ResponseWriter, r *http.Request) {
	logMsgStr := "Hanlder:Cart - DeleteCart:"

	userID := r.PathValue("user_id")
	if userID == "" {
		h.logger.Error().Msg(fmt.Sprintf("%v User ID is required", logMsgStr))
		helper.HandleResponse(w, http.StatusBadRequest, "User ID is required")
		return
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Error().AnErr("Err", err).Msg(fmt.Sprintf("%v error parse uuid: %v", logMsgStr, userID))
		helper.HandleResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var bReq model.DeleteCartRequest
	if err := helper.ParseRequestBody(r, &bReq, h.logger); err != nil {
		h.logger.Error().Any("Err", err.Error()).Msg(fmt.Sprintf("%v failed to decode request body", logMsgStr))
		helper.HandleResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	bReq.UserID = uid

	bResp, err := h.cart.DeleteCart(bReq)
	if err != nil {
		h.logger.Error().Any("Err", err.Error()).Msg(fmt.Sprintf("%v failed to DeleteCart", logMsgStr))
		helper.HandleResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	helper.HandleResponse(w, http.StatusOK, bResp)
}
