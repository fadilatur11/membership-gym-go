package subscription

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

// ListSaasPlans godoc
// @Summary List SaaS subscription plans
// @Tags SaaS Subscription
// @Produce json
// @Security BearerAuth
// @Router /v1/admin/saas-plans [get]
func (c *Controller) ListSaasPlans(ctx *gin.Context) {
	result, err := c.service.ListSaasPlans(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}

// CurrentGymSubscription godoc
// @Summary Get current gym SaaS subscription
// @Tags SaaS Subscription
// @Produce json
// @Security BearerAuth
// @Router /v1/admin/gym/subscription [get]
func (c *Controller) CurrentGymSubscription(ctx *gin.Context) {
	result, err := c.service.CurrentGymSubscription(ctx.Request.Context(), middleware.GetAuthUser(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}

// CreateGymSubscription godoc
// @Summary Subscribe gym to a SaaS plan
// @Tags SaaS Subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateGymSubscriptionRequest true "Gym subscription payload"
// @Router /v1/admin/gym/subscription [post]
func (c *Controller) CreateGymSubscription(ctx *gin.Context) {
	var request CreateGymSubscriptionRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.CreateGymSubscription(ctx.Request.Context(), middleware.GetAuthUser(ctx), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "Subscription gym berhasil dibuat", result)
}

// ListSubscriptions godoc
// @Summary List subscriptions
// @Tags Subscriptions
// @Security BearerAuth
// @Router /v1/admin/subscriptions [get]
func (c *Controller) ListSubscriptions(ctx *gin.Context) {
	result, err := c.service.ListSubscriptions(ctx.Request.Context(), middleware.GetAuthUser(ctx), pagination.Parse(ctx), shared.QueryMap(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, result.Items, result.Meta)
}

// DetailSubscription godoc
// @Summary Detail subscription
// @Tags Subscriptions
// @Security BearerAuth
// @Router /v1/admin/subscriptions/{subscription_public_id} [get]
func (c *Controller) DetailSubscription(ctx *gin.Context) {
	result, err := c.service.DetailSubscription(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("subscription_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}

// CreateSubscription godoc
// @Summary Create manual subscription
// @Tags Subscriptions
// @Security BearerAuth
// @Param request body CreateSubscriptionRequest true "Create subscription payload"
// @Router /v1/admin/subscriptions [post]
func (c *Controller) CreateSubscription(ctx *gin.Context) {
	var request CreateSubscriptionRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.CreateSubscription(ctx.Request.Context(), middleware.GetAuthUser(ctx), request, "manual")
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "Subscription berhasil dibuat", result)
}

// CancelSubscription godoc
// @Summary Cancel subscription
// @Tags Subscriptions
// @Security BearerAuth
// @Router /v1/admin/subscriptions/{subscription_public_id}/cancel [patch]
func (c *Controller) CancelSubscription(ctx *gin.Context) {
	c.changeStatus(ctx, "cancelled", "subscription.cancelled")
}

// ExpireSubscription godoc
// @Summary Expire subscription
// @Tags Subscriptions
// @Security BearerAuth
// @Router /v1/admin/subscriptions/{subscription_public_id}/expire [patch]
func (c *Controller) ExpireSubscription(ctx *gin.Context) {
	c.changeStatus(ctx, "expired", "subscription.expired")
}

func (c *Controller) changeStatus(ctx *gin.Context, status string, action string) {
	result, err := c.service.ChangeSubscriptionStatus(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("subscription_public_id"), status, action)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Data berhasil dinonaktifkan", result)
}
