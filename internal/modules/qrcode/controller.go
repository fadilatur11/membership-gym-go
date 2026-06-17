package qrcode

import (
	"github.com/gin-gonic/gin"

	"membership-gym/internal/middleware"
	"membership-gym/internal/modules/shared"
	"membership-gym/pkg/response"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) RequirePlanFeature(feature string) gin.HandlerFunc {
	return shared.RequirePlanFeature(c.service, feature)
}

func (c *Controller) GetQR(ctx *gin.Context) {
	result, err := c.service.GetOrCreateQR(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"), false)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}
func (c *Controller) GenerateQR(ctx *gin.Context) {
	result, err := c.service.GetOrCreateQR(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"), false)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "QR berhasil dibuat", result)
}
func (c *Controller) RegenerateQR(ctx *gin.Context) {
	result, err := c.service.GetOrCreateQR(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"), true)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "QR berhasil dibuat ulang", result)
}
func (c *Controller) RevokeQR(ctx *gin.Context) {
	result, err := c.service.RevokeQR(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "QR berhasil dicabut", result)
}
