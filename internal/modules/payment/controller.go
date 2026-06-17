package payment

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

func (c *Controller) ListPayments(ctx *gin.Context) {
	result, err := c.service.ListPayments(ctx.Request.Context(), middleware.GetAuthUser(ctx), pagination.Parse(ctx), shared.QueryMap(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, result.Items, result.Meta)
}
func (c *Controller) DetailPayment(ctx *gin.Context) {
	result, err := c.service.DetailPayment(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("payment_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}
func (c *Controller) CreatePayment(ctx *gin.Context) {
	var request CreatePaymentRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.CreatePayment(ctx.Request.Context(), middleware.GetAuthUser(ctx), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "Pembayaran berhasil", result)
}
func (c *Controller) CancelPayment(ctx *gin.Context) {
	result, err := c.service.ChangePayment(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("payment_public_id"), "cancelled")
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Pembayaran dibatalkan", result)
}
func (c *Controller) RefundPayment(ctx *gin.Context) {
	result, err := c.service.ChangePayment(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("payment_public_id"), "refunded")
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Pembayaran direfund", result)
}
