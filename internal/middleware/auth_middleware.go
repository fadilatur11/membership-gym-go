package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	jwthelper "membership-gym/pkg/auth"
	apperror "membership-gym/pkg/errors"
	"membership-gym/pkg/response"
)

const authUserKey = "auth_user"

type AuthUser struct {
	UserID       int64
	UserPublicID uuid.UUID
	GymID        int64
	GymPublicID  uuid.UUID
	Role         string
}

type AuthMiddleware struct {
	secret string
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{secret: secret}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Error(ctx, apperror.Unauthorized("Token tidak ditemukan"))
			ctx.Abort()
			return
		}
		claims, err := jwthelper.ParseAccessToken(strings.TrimPrefix(header, "Bearer "), m.secret)
		if err != nil {
			response.Error(ctx, apperror.Unauthorized("Token tidak valid"))
			ctx.Abort()
			return
		}
		ctx.Set(authUserKey, AuthUser{
			UserID: claims.UserID, UserPublicID: claims.UserPublicID,
			GymID: claims.GymID, GymPublicID: claims.GymPublicID, Role: claims.Role,
		})
		ctx.Next()
	}
}

func GetAuthUser(ctx *gin.Context) AuthUser {
	value, ok := ctx.Get(authUserKey)
	if !ok {
		return AuthUser{}
	}
	authUser, _ := value.(AuthUser)
	return authUser
}
