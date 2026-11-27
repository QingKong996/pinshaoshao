package controller

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"pinshaoshao/internal/feature/address/apperr"
	"pinshaoshao/internal/feature/address/dto"
	"pinshaoshao/internal/feature/address/model"
	"pinshaoshao/internal/feature/address/service"
	"pinshaoshao/internal/http/authctx"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AddressService interface {
	List(ctx context.Context, userID int) ([]model.Address, error)
	Create(ctx context.Context, userID int, input service.AddressInput) (*model.Address, error)
	Update(ctx context.Context, userID int, addressID int, input service.AddressInput) (*model.Address, error)
	Delete(ctx context.Context, userID int, addressID int) error
	SetDefault(ctx context.Context, userID int, addressID int) (*model.Address, error)
}

type AddressController struct {
	addressService AddressService
}

func NewAddressController(addressService AddressService) *AddressController {
	return &AddressController{addressService: addressService}
}

// List returns the authenticated buyer's saved addresses.
//
// @Summary     List saved addresses
// @Description List all delivery addresses saved by the authenticated buyer.
// @Tags        Addresses
// @Produce     json
// @Success     200 {array} dto.AddressResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /addresses [get]
func (c *AddressController) List(ctx *gin.Context) {
	addresses, err := c.addressService.List(
		ctx.Request.Context(),
		authctx.MustUserID(ctx),
	)
	if err != nil {
		c.writeError(ctx, "list addresses", err)
		return
	}

	response := make([]dto.AddressResponse, 0, len(addresses))
	for i := range addresses {
		response = append(response, newAddressResponse(&addresses[i]))
	}
	ctx.JSON(http.StatusOK, response)
}

// Create saves an address for the authenticated buyer.
//
// @Summary     Create a saved address
// @Description Save a delivery address. The first saved address becomes the default.
// @Tags        Addresses
// @Accept      json
// @Produce     json
// @Param       request body dto.SaveAddressRequest true "Address details"
// @Success     201 {object} dto.AddressResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /addresses [post]
func (c *AddressController) Create(ctx *gin.Context) {
	var req dto.SaveAddressRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: apperr.ErrInvalidAddress.Error()})
		return
	}

	address, err := c.addressService.Create(
		ctx.Request.Context(),
		authctx.MustUserID(ctx),
		newAddressInput(&req),
	)
	if err != nil {
		c.writeError(ctx, "create address", err)
		return
	}
	ctx.JSON(http.StatusCreated, newAddressResponse(address))
}

// Update replaces a saved address owned by the authenticated buyer.
//
// @Summary     Update a saved address
// @Description Update an owned delivery address and optionally make it the default.
// @Tags        Addresses
// @Accept      json
// @Produce     json
// @Param       addressID path int true "Address ID"
// @Param       request body dto.SaveAddressRequest true "Address details"
// @Success     200 {object} dto.AddressResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /addresses/{addressID} [put]
func (c *AddressController) Update(ctx *gin.Context) {
	addressID, ok := parseAddressID(ctx)
	if !ok {
		return
	}

	var req dto.SaveAddressRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: apperr.ErrInvalidAddress.Error()})
		return
	}

	address, err := c.addressService.Update(
		ctx.Request.Context(),
		authctx.MustUserID(ctx),
		addressID,
		newAddressInput(&req),
	)
	if err != nil {
		c.writeError(ctx, "update address", err)
		return
	}
	ctx.JSON(http.StatusOK, newAddressResponse(address))
}

// Delete removes an owned saved address.
//
// @Summary     Delete a saved address
// @Description Remove an owned address. If it was the default, the newest remaining address becomes the default.
// @Tags        Addresses
// @Param       addressID path int true "Address ID"
// @Success     204
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /addresses/{addressID} [delete]
func (c *AddressController) Delete(ctx *gin.Context) {
	addressID, ok := parseAddressID(ctx)
	if !ok {
		return
	}

	if err := c.addressService.Delete(
		ctx.Request.Context(),
		authctx.MustUserID(ctx),
		addressID,
	); err != nil {
		c.writeError(ctx, "delete address", err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// SetDefault makes an owned address the default.
//
// @Summary     Set the default address
// @Description Make an owned saved address the buyer's default delivery address.
// @Tags        Addresses
// @Produce     json
// @Param       addressID path int true "Address ID"
// @Success     200 {object} dto.AddressResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     404 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Security    BearerAuth
// @Router      /addresses/{addressID}/default [post]
func (c *AddressController) SetDefault(ctx *gin.Context) {
	addressID, ok := parseAddressID(ctx)
	if !ok {
		return
	}

	address, err := c.addressService.SetDefault(
		ctx.Request.Context(),
		authctx.MustUserID(ctx),
		addressID,
	)
	if err != nil {
		c.writeError(ctx, "set default address", err)
		return
	}
	ctx.JSON(http.StatusOK, newAddressResponse(address))
}

func parseAddressID(ctx *gin.Context) (int, bool) {
	addressID, err := strconv.Atoi(ctx.Param("addressID"))
	if err != nil || addressID <= 0 {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: apperr.ErrInvalidAddressID.Error()})
		return 0, false
	}
	return addressID, true
}

func (c *AddressController) writeError(ctx *gin.Context, operation string, err error) {
	switch {
	case errors.Is(err, apperr.ErrInvalidAddress):
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: apperr.ErrInvalidAddress.Error()})
	case errors.Is(err, apperr.ErrInvalidAddressID):
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: apperr.ErrInvalidAddressID.Error()})
	case errors.Is(err, apperr.ErrAddressNotFound):
		ctx.JSON(http.StatusNotFound, dto.ErrorResponse{Error: apperr.ErrAddressNotFound.Error()})
	default:
		slog.ErrorContext(ctx.Request.Context(), operation, "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: apperr.ErrInternalServer.Error()})
	}
}

func newAddressInput(req *dto.SaveAddressRequest) service.AddressInput {
	return service.AddressInput{
		ReceiverName:    req.ReceiverName,
		ReceiverPhone:   req.ReceiverPhone,
		ShippingAddress: req.ShippingAddress,
		IsDefault:       req.IsDefault,
	}
}

func newAddressResponse(address *model.Address) dto.AddressResponse {
	return dto.AddressResponse{
		ID:              address.ID,
		ReceiverName:    address.ReceiverName,
		ReceiverPhone:   address.ReceiverPhone,
		ShippingAddress: address.ShippingAddress,
		IsDefault:       address.IsDefault,
		CreatedAt:       address.CreatedAt,
		UpdatedAt:       address.UpdatedAt,
	}
}
