package user

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

func (c *Controller) ListUsers(ctx *gin.Context) {
	result, err := c.service.ListUsers(ctx.Request.Context(), middleware.GetAuthUser(ctx), pagination.Parse(ctx), shared.QueryMap(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, result.Items, result.Meta)
}
func (c *Controller) DetailUser(ctx *gin.Context) {
	result, err := c.service.DetailUser(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("user_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}
func (c *Controller) UpdateUserDTO(ctx *gin.Context) {
	var request UpdateUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.UpdateUser(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("user_public_id"), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Data berhasil diperbarui", result)
}
func (c *Controller) DeleteUser(ctx *gin.Context) {
	result, err := c.service.DeactivateUser(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("user_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Data berhasil dinonaktifkan", result)
}
func (c *Controller) CreateUser(ctx *gin.Context) {
	var request CreateUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.CreateUser(ctx.Request.Context(), middleware.GetAuthUser(ctx), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "User berhasil dibuat", result)
}
