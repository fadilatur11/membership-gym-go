package report

import (
	"context"

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

func (c *Controller) Dashboard(ctx *gin.Context) {
	result, err := c.service.Dashboard(ctx.Request.Context(), middleware.GetAuthUser(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}
func (c *Controller) ReportPayments(ctx *gin.Context) { c.list(ctx, c.service.ReportPayments) }
func (c *Controller) ReportExpenses(ctx *gin.Context) { c.list(ctx, c.service.ReportExpenses) }
func (c *Controller) ReportProfitLoss(ctx *gin.Context) {
	result, err := c.service.Dashboard(ctx.Request.Context(), middleware.GetAuthUser(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", gin.H{"total_income": result["income_this_month"], "total_expense": result["expense_this_month"], "net_profit": result["net_profit_this_month"]})
}
func (c *Controller) ReportCheckins(ctx *gin.Context) { c.list(ctx, c.service.ReportCheckins) }
func (c *Controller) ReportExpiredMembers(ctx *gin.Context) {
	c.list(ctx, c.service.ReportExpiredMembers)
}
func (c *Controller) V2MemberProgressReport(ctx *gin.Context) {
	result, err := c.service.MemberProgressReport(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}
func (c *Controller) V2WorkoutConsistencyReport(ctx *gin.Context) {
	c.list(ctx, c.service.WorkoutConsistency)
}
func (c *Controller) V2WeightProgressReport(ctx *gin.Context) {
	response.OK(ctx, "Berhasil", gin.H{"message": "Gunakan endpoint member progress untuk detail per member"})
}
func (c *Controller) list(ctx *gin.Context, fn func(context.Context, middleware.AuthUser, pagination.Params, map[string]string) (PageResult, error)) {
	result, err := fn(ctx.Request.Context(), middleware.GetAuthUser(ctx), pagination.Parse(ctx), shared.QueryMap(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, result.Items, result.Meta)
}
