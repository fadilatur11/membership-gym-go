package membership_package

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

func (c *Controller) ListPackages(ctx *gin.Context) {
	result, err := c.service.ListPackages(ctx.Request.Context(), middleware.GetAuthUser(ctx), pagination.Parse(ctx), shared.QueryMap(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, result.Items, result.Meta)
}
func (c *Controller) DetailPackage(ctx *gin.Context) {
	result, err := c.service.DetailPackage(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("package_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}
func (c *Controller) UpdatePackageDTO(ctx *gin.Context) {
	var request UpdateMembershipPackageRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.UpdatePackage(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("package_public_id"), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Data berhasil diperbarui", result)
}
func (c *Controller) DeletePackage(ctx *gin.Context) {
	result, err := c.service.DeactivatePackage(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("package_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Data berhasil dinonaktifkan", result)
}
func (c *Controller) CreatePackage(ctx *gin.Context) {
	var request CreateMembershipPackageRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.CreatePackage(ctx.Request.Context(), middleware.GetAuthUser(ctx), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "Package berhasil dibuat", result)
}
func (c *Controller) PublicPackages(ctx *gin.Context) {
	items, err := c.service.PublicPackages(ctx.Request.Context(), ctx.Param("qr_token"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", items)
}
