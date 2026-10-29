package controller

import (
	"errors"
	"net/http"
	"pinshaoshao/internal/apperr"
	"pinshaoshao/internal/dto"
	"pinshaoshao/internal/service"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{userService: userService}
}

func (c *UserController) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": apperr.ErrInvalidJsonFormat.Error(),
		})
		return
	}

	token, err := c.userService.Login(ctx.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, apperr.ErrInvalidCredentials) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": apperr.ErrInvalidCredentials.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": apperr.ErrInternalServer.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (c *UserController) Register(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": apperr.ErrInvalidJsonFormat.Error(),
		})
		return
	}

	token, err := c.userService.Register(ctx.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, apperr.ErrUsernameTaken) {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": apperr.ErrUsernameTaken.Error(),
			})
			return
		}
		if errors.Is(err, apperr.ErrEmailTaken) {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": apperr.ErrEmailTaken.Error(),
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": apperr.ErrInternalServer.Error(),
		})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"token": token,
	})
}
