package reminder_rule

import (
	"github.com/gin-gonic/gin"

	"membership-gym/internal/middleware"
	"membership-gym/internal/modules/shared"
	"membership-gym/pkg/pagination"
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

func (c *Controller) ListReminderRules(ctx *gin.Context) {
	result, err := c.service.ListRules(ctx.Request.Context(), middleware.GetAuthUser(ctx), pagination.Parse(ctx), shared.QueryMap(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, result.Items, result.Meta)
}
func (c *Controller) CreateReminderRule(ctx *gin.Context) {
	var request CreateReminderRuleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.CreateRule(ctx.Request.Context(), middleware.GetAuthUser(ctx), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "Reminder rule berhasil dibuat", result)
}
func (c *Controller) DetailReminderRule(ctx *gin.Context) {
	result, err := c.service.DetailRule(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("rule_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}
func (c *Controller) UpdateReminderRuleDTO(ctx *gin.Context) {
	var request UpdateReminderRuleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.UpdateRule(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("rule_public_id"), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Data berhasil diperbarui", result)
}
func (c *Controller) DeleteReminderRule(ctx *gin.Context) {
	result, err := c.service.DeactivateRule(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("rule_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Data berhasil dinonaktifkan", result)
}
