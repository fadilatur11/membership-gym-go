package middleware

import (
	"github.com/gin-gonic/gin"

	apperror "membership-gym/pkg/errors"
	"membership-gym/pkg/response"
)

type RoleMiddleware struct{}

func NewRoleMiddleware() *RoleMiddleware {
	return &RoleMiddleware{}
}

func (m *RoleMiddleware) Require(roles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, role := range roles {
		allowed[role] = true
	}
	return func(ctx *gin.Context) {
		authUser := GetAuthUser(ctx)
		if !allowed[authUser.Role] {
			response.Error(ctx, apperror.Forbidden("Akses tidak diizinkan"))
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
