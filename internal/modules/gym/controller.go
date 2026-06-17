package gym

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

func (c *Controller) GymProfile(ctx *gin.Context) {
	result, err := c.service.GymProfile(ctx.Request.Context(), middleware.GetAuthUser(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}

func (c *Controller) UpdateGym(ctx *gin.Context) {
	var request UpdateGymRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.UpdateGym(ctx.Request.Context(), middleware.GetAuthUser(ctx), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Gym berhasil diperbarui", result)
}

func (c *Controller) PublicGym(ctx *gin.Context) {
	result, err := c.service.PublicGymContact(ctx.Request.Context(), ctx.Param("qr_token"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}
