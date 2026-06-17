package shared

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"

	"membership-gym/internal/middleware"
	apperror "membership-gym/pkg/errors"
	"membership-gym/pkg/pagination"
	"membership-gym/pkg/response"
)

type ResourceConfig struct {
	Table        string
	Select       string
	SearchSQL    string
	Filters      map[string]string
	UpdateFields map[string]string
	StatusColumn string
}

type PageResult struct {
	Items []map[string]any
	Meta  pagination.Meta
}

type PlanService interface {
	PlanAllows(ctx context.Context, auth middleware.AuthUser, feature string) (bool, error)
}

type GenericService interface {
	List(ctx context.Context, auth middleware.AuthUser, cfg ResourceConfig, p pagination.Params, q map[string]string) (PageResult, error)
	Detail(ctx context.Context, auth middleware.AuthUser, cfg ResourceConfig, publicID string) (map[string]any, error)
	UpdateGeneric(ctx context.Context, auth middleware.AuthUser, cfg ResourceConfig, publicID string, req map[string]any) (map[string]any, error)
	ChangeStatus(ctx context.Context, auth middleware.AuthUser, cfg ResourceConfig, publicID string, status any, action string) (map[string]any, error)
	PlanService
}

func RequirePlanFeature(service PlanService, feature string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		allowed, err := service.PlanAllows(ctx.Request.Context(), middleware.GetAuthUser(ctx), feature)
		if err != nil {
			response.Error(ctx, err)
			ctx.Abort()
			return
		}
		if !allowed {
			response.Error(ctx, apperror.Forbidden("Fitur tidak tersedia pada subscription plan gym ini"))
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func QueryMap(ctx *gin.Context) map[string]string {
	return map[string]string{
		"search":         ctx.Query("search"),
		"role":           ctx.Query("role"),
		"is_active":      ctx.Query("is_active"),
		"status":         ctx.Query("status"),
		"payment_method": ctx.Query("payment_method"),
		"source":         ctx.Query("source"),
		"entity_type":    ctx.Query("entity_type"),
	}
}

func DTOToMap(dto any) map[string]any {
	raw, _ := json.Marshal(dto)
	result := map[string]any{}
	_ = json.Unmarshal(raw, &result)
	return result
}

func List(ctx *gin.Context, service GenericService, cfg ResourceConfig) {
	result, err := service.List(ctx.Request.Context(), middleware.GetAuthUser(ctx), cfg, pagination.Parse(ctx), QueryMap(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, result.Items, result.Meta)
}

func Detail(ctx *gin.Context, service GenericService, cfg ResourceConfig, param string) {
	result, err := service.Detail(ctx.Request.Context(), middleware.GetAuthUser(ctx), cfg, ctx.Param(param))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}

func UpdateDTO[T any](ctx *gin.Context, service GenericService, cfg ResourceConfig, param string) {
	var request T
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := service.UpdateGeneric(ctx.Request.Context(), middleware.GetAuthUser(ctx), cfg, ctx.Param(param), DTOToMap(request))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Data berhasil diperbarui", result)
}

func ChangeStatus(ctx *gin.Context, service GenericService, cfg ResourceConfig, param string, value any, action string) {
	result, err := service.ChangeStatus(ctx.Request.Context(), middleware.GetAuthUser(ctx), cfg, ctx.Param(param), value, action)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Data berhasil dinonaktifkan", result)
}
