package routes

import (
	"github.com/gin-gonic/gin"

	"membership-gym/internal/middleware"
	"membership-gym/internal/modules"
)

type Dependencies struct {
	modules.Controllers
	AuthMiddleware *middleware.AuthMiddleware
	RoleMiddleware *middleware.RoleMiddleware
}

func Register(router *gin.Engine, deps Dependencies) {
	v1 := router.Group("/api/v1")
	auth := v1.Group("/auth")
	auth.POST("/register-owner", deps.Auth.RegisterOwner)
	auth.GET("/google", deps.Auth.GoogleAuthURL)
	auth.GET("/google/callback", deps.Auth.GoogleCallback)
	RegisterAdminRoutes(v1.Group("/admin"), deps)
	RegisterMemberRoutes(v1.Group("/member"), deps)
	RegisterV2Routes(router.Group("/api/v2"), deps)
}
