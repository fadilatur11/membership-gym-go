package member

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

func (c *Controller) ListMembers(ctx *gin.Context) {
	result, err := c.service.ListMembers(ctx.Request.Context(), middleware.GetAuthUser(ctx), pagination.Parse(ctx), shared.QueryMap(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, result.Items, result.Meta)
}

func (c *Controller) DetailMember(ctx *gin.Context) {
	result, err := c.service.DetailMember(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}

func (c *Controller) UpdateMemberDTO(ctx *gin.Context) {
	var request UpdateMemberRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.UpdateMember(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Data berhasil diperbarui", result)
}

func (c *Controller) DeleteMember(ctx *gin.Context) {
	result, err := c.service.DeactivateMember(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Data berhasil dinonaktifkan", result)
}

func (c *Controller) CreateMember(ctx *gin.Context) {
	var request CreateMemberRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.CreateMember(ctx.Request.Context(), middleware.GetAuthUser(ctx), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "Member berhasil dibuat", result)
}

func (c *Controller) PublicQRCode(ctx *gin.Context) {
	result, err := c.service.PublicIdentity(ctx.Request.Context(), ctx.Param("qr_token"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}

func (c *Controller) PublicStatus(ctx *gin.Context) {
	result, err := c.service.PublicStatus(ctx.Request.Context(), ctx.Param("qr_token"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}
