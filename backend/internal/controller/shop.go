package controller

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/middleware"
	"pinshaoshao/internal/model"

	"github.com/gin-gonic/gin"
)

// ShopService is the shop service surface the controller depends on.
type ShopService interface {
	Login(ctx context.Context, req *dto.LoginRequest) (string, error)
	Register(ctx context.Context, req *dto.ShopRegisterRequest) (string, error)
	Logout(ctx context.Context, token string) error
	GetProfile(ctx context.Context, shopID int) (*model.Shop, error)
	UpdateProfile(ctx context.Context, shopID int, req *dto.UpdateShopProfileRequest) (*model.Shop, error)
}

type ShopController struct {
	shopService ShopService
}

func NewShopController(shopService ShopService) *ShopController {
	return &ShopController{shopService: shopService}
}

// Login authenticates a shop and returns a session token.
//
// @Summary     Merchant login
// @Description Verify the shop username and password, and return a session token on success.
// @Tags        Merchant Authentication
// @Accept      json
// @Produce     json
// @Param       request body dto.LoginRequest true "Login credentials"
// @Success     200 {object} dto.TokenResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Router      /merchant/login [post]
func (c *ShopController) Login(ctx *gin.Context) {
	login(ctx, c.shopService.Login, apperr.ErrShopBanned)
}

// Register creates a shop and returns a session token.
//
// @Summary     Merchant registration
// @Description Create a shop account and return a session token on success.
// @Tags        Merchant Authentication
// @Accept      json
// @Produce     json
// @Param       request body dto.ShopRegisterRequest true "Registration details"
// @Success     201 {object} dto.TokenResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse "Shop username or email already exists"
// @Failure     500 {object} dto.ErrorResponse
// @Router      /merchant/register [post]
func (c *ShopController) Register(ctx *gin.Context) {
	var req dto.ShopRegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidJsonFormat.Error(),
		})
		return
	}

	token, err := c.shopService.Register(ctx.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrShopUsernameTaken):
			ctx.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: apperr.ErrShopUsernameTaken.Error(),
			})
		case errors.Is(err, apperr.ErrShopEmailTaken):
			ctx.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: apperr.ErrShopEmailTaken.Error(),
			})
		default:
			ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Error: apperr.ErrInternalServer.Error(),
			})
		}
		return
	}
	ctx.JSON(http.StatusCreated, dto.TokenResponse{
		Token: token,
	})
}

// Logout revokes the current shop session.
//
// @Summary     Merchant logout
// @Description Revoke the current shop session token.
// @Tags        Merchant Authentication
// @Security    BearerAuth
// @Success     204
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Router      /merchant/logout [post]
func (c *ShopController) Logout(ctx *gin.Context) {
	token := middleware.MustShopSessionToken(ctx)
	logout(ctx, token, c.shopService.Logout)
}

// GetProfile returns the profile of the authenticated shop.
//
// @Summary     Get merchant shop profile
// @Description Return the profile of the authenticated shop.
// @Tags        Merchant Shop
// @Produce     json
// @Success     200 {object} dto.ShopProfileResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/shop [get]
func (c *ShopController) GetProfile(ctx *gin.Context) {
	shop, err := c.shopService.GetProfile(ctx.Request.Context(), middleware.MustCurrentShop(ctx).ID)
	if err != nil {
		if errors.Is(err, apperr.ErrShopNotFound) {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: apperr.ErrShopNotFound.Error(),
			})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), "get shop profile", "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: apperr.ErrInternalServer.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, newShopProfileResponse(shop))
}

// UpdateProfile updates the editable profile fields of the authenticated shop.
//
// @Summary     Update merchant shop profile
// @Description Update the display name and description of the authenticated shop.
// @Tags        Merchant Shop
// @Accept      json
// @Produce     json
// @Param       request body dto.UpdateShopProfileRequest true "Profile fields"
// @Success     200 {object} dto.ShopProfileResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     404 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Security    BearerAuth
// @Router      /merchant/shop [patch]
func (c *ShopController) UpdateProfile(ctx *gin.Context) {
	var req dto.UpdateShopProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidJsonFormat.Error(),
		})
		return
	}

	shop, err := c.shopService.UpdateProfile(ctx.Request.Context(), middleware.MustCurrentShop(ctx).ID, &req)
	if err != nil {
		if errors.Is(err, apperr.ErrShopNotFound) {
			ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: apperr.ErrShopNotFound.Error(),
			})
			return
		}
		slog.ErrorContext(ctx.Request.Context(), "update shop profile", "error", err)
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: apperr.ErrInternalServer.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, newShopProfileResponse(shop))
}

func newShopProfileResponse(shop *model.Shop) dto.ShopProfileResponse {
	return dto.ShopProfileResponse{
		ID:          shop.ID,
		Username:    shop.Username,
		Name:        shop.Name,
		Description: shop.Description,
		Email:       shop.Email,
		Status:      string(shop.Status),
		CreatedAt:   shop.CreatedAt,
		UpdatedAt:   shop.UpdatedAt,
	}
}
