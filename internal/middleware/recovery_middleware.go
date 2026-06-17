package middleware

import "github.com/gin-gonic/gin"

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(ctx *gin.Context, recovered any) {
		ctx.JSON(500, gin.H{"success": false, "message": "Terjadi kesalahan server", "errors": nil})
	})
}
