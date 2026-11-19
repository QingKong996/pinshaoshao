package controller

import (
	"context"
	"errors"
	"net/http"
	"pinshaoshao/internal/feature/shared/authctx"
	"pinshaoshao/internal/feature/shared/httpx"
	"pinshaoshao/internal/feature/user/apperr"
	"pinshaoshao/internal/feature/user/dto"
	userservice "pinshaoshao/internal/feature/user/service"

	"github.com/gin-gonic/gin"
)

// UserService is the user service surface the controller depends on.
type UserService interface {
	Login(ctx context.Context, input userservice.LoginInput) (string, error)
	Register(ctx context.Context, input userservice.RegisterUserInput) (string, error)
	Logout(ctx context.Context, token string) error
}

type UserController struct {
	userService UserService
}

func NewUserController(userService UserService) *UserController {
	return &UserController{userService: userService}
}

// Login authenticates a user and returns a session token.
//
// @Summary     User login
// @Description Verify the username and password, and return a session token on success.
// @Tags        Authentication
// @Accept      json
// @Produce     json
// @Param       request body httpx.LoginRequest true "Login credentials"
// @Success     200 {object} httpx.TokenResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Router      /login [post]
func (c *UserController) Login(ctx *gin.Context) {
	httpx.Login(ctx, func(ctx context.Context, username, password string) (string, error) {
		return c.userService.Login(ctx, userservice.LoginInput{
			Username: username,
			Password: password,
		})
	}, apperr.ErrInvalidCredentials, apperr.ErrUserBanned)
}

// Register creates a user and returns a session token.
//
// @Summary     User registration
// @Description Create an account and return a session token on success.
// @Tags        Authentication
// @Accept      json
// @Produce     json
// @Param       request body dto.RegisterRequest true "Registration details"
// @Success     201 {object} httpx.TokenResponse
// @Failure     400 {object} httpx.ErrorResponse
// @Failure     409 {object} httpx.ErrorResponse "Username or email already exists"
// @Failure     500 {object} httpx.ErrorResponse
// @Router      /register [post]
func (c *UserController) Register(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidJsonFormat.Error(),
		})
		return
	}

	token, err := c.userService.Register(ctx.Request.Context(), userservice.RegisterUserInput{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	})
	if err != nil {
		if errors.Is(err, apperr.ErrUsernameTaken) {
			ctx.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: apperr.ErrUsernameTaken.Error(),
			})
			return
		}
		if errors.Is(err, apperr.ErrEmailTaken) {
			ctx.JSON(http.StatusConflict, dto.ErrorResponse{
				Error: apperr.ErrEmailTaken.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: apperr.ErrInternalServer.Error(),
		})
		return
	}
	ctx.JSON(http.StatusCreated, dto.TokenResponse{
		Token: token,
	})
}

// Logout revokes the current session.
//
// @Summary     User logout
// @Description Revoke the current session token.
// @Tags        Authentication
// @Security    BearerAuth
// @Success     204
// @Failure     401 {object} httpx.ErrorResponse
// @Failure     403 {object} httpx.ErrorResponse
// @Failure     500 {object} httpx.ErrorResponse
// @Router      /logout [post]
func (c *UserController) Logout(ctx *gin.Context) {
	token := authctx.MustToken(ctx, authctx.UserTokenKey, "controller.UserController.Logout: route is not protected by UserAuth")
	httpx.Logout(ctx, token, c.userService.Logout)
}
