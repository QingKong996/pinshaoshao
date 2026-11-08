package controller

import (
	"errors"
	"net/http"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/middleware"
	"pinshaoshao/internal/service"

	"github.com/gin-gonic/gin"
)

type ShopController struct {
	shopService *service.ShopService
}

func NewShopController(shopService *service.ShopService) *ShopController {
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
