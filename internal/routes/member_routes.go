package routes

import "github.com/gin-gonic/gin"

func RegisterMemberRoutes(r *gin.RouterGroup, deps Dependencies) {
	r.GET("/qrcode/:qr_token", deps.Member.PublicQRCode)
	r.GET("/status/:qr_token", deps.Member.PublicStatus)
	r.POST("/checkins/scan", deps.Checkin.PublicScan)
	r.GET("/checkins/:qr_token", deps.Checkin.PublicCheckins)
	r.GET("/packages/:qr_token", deps.MembershipPackage.PublicPackages)
	r.GET("/gym/:qr_token", deps.Gym.PublicGym)
}
