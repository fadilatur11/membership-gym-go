package expense

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

func (c *Controller) ListExpenses(ctx *gin.Context) {
	result, err := c.service.ListExpenses(ctx.Request.Context(), middleware.GetAuthUser(ctx), pagination.Parse(ctx), shared.QueryMap(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, result.Items, result.Meta)
}
func (c *Controller) DetailExpense(ctx *gin.Context) {
	result, err := c.service.DetailExpense(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("expense_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}
func (c *Controller) UpdateExpenseDTO(ctx *gin.Context) {
	var request UpdateExpenseRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.UpdateExpense(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("expense_public_id"), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Data berhasil diperbarui", result)
}
func (c *Controller) CreateExpense(ctx *gin.Context) {
	var request CreateExpenseRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.CreateExpense(ctx.Request.Context(), middleware.GetAuthUser(ctx), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "Expense berhasil dibuat", result)
}
func (c *Controller) ApproveExpense(ctx *gin.Context) {
	result, err := c.service.ChangeExpenseStatus(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("expense_public_id"), "approved", "expense.approved")
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Data berhasil diperbarui", result)
}
func (c *Controller) RejectExpense(ctx *gin.Context) {
	result, err := c.service.ChangeExpenseStatus(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("expense_public_id"), "rejected", "expense.rejected")
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Data berhasil diperbarui", result)
}
