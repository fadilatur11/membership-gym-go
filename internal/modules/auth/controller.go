package auth

import (
	"github.com/gin-gonic/gin"

	"membership-gym/internal/middleware"
	"membership-gym/pkg/response"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// Login godoc
// @Summary Login admin/staff
// @Tags Admin Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login payload"
// @Router /v1/admin/auth/login [post]
func (c *Controller) Login(ctx *gin.Context) {
	var request LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}

	result, err := c.service.Login(ctx.Request.Context(), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.OK(ctx, "Login berhasil", result)
}

// Refresh godoc
// @Summary Refresh access token
// @Tags Admin Auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token payload"
// @Router /v1/admin/auth/refresh [post]
func (c *Controller) Refresh(ctx *gin.Context) {
	var request RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}

	result, err := c.service.Refresh(ctx.Request.Context(), request.RefreshToken)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.OK(ctx, "Token berhasil diperbarui", result)
}

// Logout godoc
// @Summary Logout admin/staff
// @Tags Admin Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body LogoutRequest true "Logout payload"
// @Router /v1/admin/auth/logout [post]
func (c *Controller) Logout(ctx *gin.Context) {
	var request LogoutRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}

	if err := c.service.Logout(ctx.Request.Context(), request.RefreshToken); err != nil {
		response.Error(ctx, err)
		return
	}

	response.OK(ctx, "Logout berhasil", nil)
}

// RegisterOwner godoc
// @Summary Register new gym owner
// @Tags Public Auth
// @Accept json
// @Produce json
// @Param request body RegisterOwnerRequest true "Register owner payload"
// @Router /v1/auth/register-owner [post]
func (c *Controller) RegisterOwner(ctx *gin.Context) {
	var request RegisterOwnerRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}

	result, err := c.service.RegisterOwner(ctx.Request.Context(), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Created(ctx, "Owner berhasil didaftarkan", result)
}

// GoogleAuthURL godoc
// @Summary Get Google OAuth URL
// @Tags Public Auth
// @Produce json
// @Router /v1/auth/google [get]
func (c *Controller) GoogleAuthURL(ctx *gin.Context) {
	result, err := c.service.GoogleAuthURL(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.OK(ctx, "Berhasil", result)
}

// GoogleCallback godoc
// @Summary Google OAuth callback
// @Tags Public Auth
// @Produce json
// @Param code query string true "Google authorization code"
// @Param state query string true "OAuth state"
// @Router /v1/auth/google/callback [get]
func (c *Controller) GoogleCallback(ctx *gin.Context) {
	result, err := c.service.GoogleCallback(ctx.Request.Context(), ctx.Query("code"), ctx.Query("state"))
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.OK(ctx, "Login Google berhasil", result)
}

// Profile godoc
// @Summary Get authenticated profile
// @Tags Admin Auth
// @Produce json
// @Security BearerAuth
// @Router /v1/admin/auth/profile [get]
func (c *Controller) Profile(ctx *gin.Context) {
	result, err := c.service.Profile(ctx.Request.Context(), middleware.GetAuthUser(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.OK(ctx, "Berhasil", result)
}
