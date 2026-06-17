package checkin

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

// ScanCheckin godoc
// @Summary Admin QR check-in
// @Tags Check-ins
// @Security BearerAuth
// @Param request body ScanCheckinRequest true "QR scan payload"
// @Router /v1/admin/checkins/scan [post]
func (c *Controller) ScanCheckin(ctx *gin.Context) {
	var request ScanCheckinRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}

	auth := middleware.GetAuthUser(ctx)
	result, err := c.service.CheckinByQR(ctx.Request.Context(), &auth, request.QRToken)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Created(ctx, "Check-in berhasil", result)
}

// ManualCheckin godoc
// @Summary Manual check-in
// @Tags Check-ins
// @Security BearerAuth
// @Param request body ManualCheckinRequest true "Manual check-in payload"
// @Router /v1/admin/checkins/manual [post]
func (c *Controller) ManualCheckin(ctx *gin.Context) {
	var request ManualCheckinRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}

	result, err := c.service.ManualCheckin(ctx.Request.Context(), middleware.GetAuthUser(ctx), request.MemberPublicID, request.Notes)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Created(ctx, "Check-in berhasil", result)
}

func (c *Controller) ListCheckins(ctx *gin.Context) {
	c.list(ctx)
}

func (c *Controller) TodayCheckins(ctx *gin.Context) {
	c.list(ctx)
}

func (c *Controller) MemberCheckins(ctx *gin.Context) {
	c.list(ctx)
}

// PublicScan godoc
// @Summary Public QR check-in
// @Tags Public Member
// @Param request body ScanCheckinRequest true "Public QR scan payload"
// @Router /v1/member/checkins/scan [post]
func (c *Controller) PublicScan(ctx *gin.Context) {
	var request ScanCheckinRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}

	result, err := c.service.CheckinByQR(ctx.Request.Context(), nil, request.QRToken)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Created(ctx, "Check-in berhasil", result)
}

func (c *Controller) PublicCheckins(ctx *gin.Context) {
	result, err := c.service.PublicCheckins(ctx.Request.Context(), ctx.Param("qr_token"), pagination.Parse(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Paginated(ctx, result.Items, result.Meta)
}

func (c *Controller) list(ctx *gin.Context) {
	result, err := c.service.ListCheckins(ctx.Request.Context(), middleware.GetAuthUser(ctx), pagination.Parse(ctx), shared.QueryMap(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Paginated(ctx, result.Items, result.Meta)
}
