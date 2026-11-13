package controller

import (
	"context"
	"errors"
	"net/http"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/middleware"
	"pinshaoshao/internal/service"

	"github.com/gin-gonic/gin"
)

// UserService is the user service surface the controller depends on.
type UserService interface {
	Login(ctx context.Context, input service.LoginInput) (string, error)
	Register(ctx context.Context, input service.RegisterUserInput) (string, error)
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
// @Param       request body dto.LoginRequest true "Login credentials"
// @Success     200 {object} dto.TokenResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Router      /login [post]
func (c *UserController) Login(ctx *gin.Context) {
	login(ctx, c.userService.Login, apperr.ErrUserBanned)
}

// Register creates a user and returns a session token.
//
// @Summary     User registration
// @Description Create an account and return a session token on success.
// @Tags        Authentication
// @Accept      json
// @Produce     json
// @Param       request body dto.RegisterRequest true "Registration details"
// @Success     201 {object} dto.TokenResponse
// @Failure     400 {object} dto.ErrorResponse
// @Failure     409 {object} dto.ErrorResponse "Username or email already exists"
// @Failure     500 {object} dto.ErrorResponse
// @Router      /register [post]
func (c *UserController) Register(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: apperr.ErrInvalidJsonFormat.Error(),
		})
		return
	}

	token, err := c.userService.Register(ctx.Request.Context(), service.RegisterUserInput{
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
// @Failure     401 {object} dto.ErrorResponse
// @Failure     403 {object} dto.ErrorResponse
// @Failure     500 {object} dto.ErrorResponse
// @Router      /logout [post]
func (c *UserController) Logout(ctx *gin.Context) {
	token := middleware.MustUserSessionToken(ctx)
	logout(ctx, token, c.userService.Logout)
}
