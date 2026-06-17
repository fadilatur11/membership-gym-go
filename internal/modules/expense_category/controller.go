package expense_category

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

func (c *Controller) ListExpenseCategories(ctx *gin.Context) {
	result, err := c.service.ListCategories(ctx.Request.Context(), middleware.GetAuthUser(ctx), pagination.Parse(ctx), shared.QueryMap(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, result.Items, result.Meta)
}
func (c *Controller) CreateExpenseCategory(ctx *gin.Context) {
	var request CreateExpenseCategoryRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.CreateCategory(ctx.Request.Context(), middleware.GetAuthUser(ctx), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "Kategori expense berhasil dibuat", result)
}
func (c *Controller) UpdateExpenseCategoryDTO(ctx *gin.Context) {
	var request UpdateExpenseCategoryRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.UpdateCategory(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("category_public_id"), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Data berhasil diperbarui", result)
}
func (c *Controller) DeleteExpenseCategory(ctx *gin.Context) {
	result, err := c.service.DeactivateCategory(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("category_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Data berhasil dinonaktifkan", result)
}
