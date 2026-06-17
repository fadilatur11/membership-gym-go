package routes

import "github.com/gin-gonic/gin"

func RegisterV2Routes(api *gin.RouterGroup, deps Dependencies) {
	admin := api.Group("/admin")
	admin.Use(deps.AuthMiddleware.RequireAuth())
	admin.Use(deps.Workout.RequirePlanFeature("workout.v2"))

	admin.GET("/muscle-groups", deps.Workout.V2ListMuscleGroups)
	admin.POST("/muscle-groups", deps.RoleMiddleware.Require("owner", "admin", "trainer"), deps.Workout.V2CreateMuscleGroup)
	admin.GET("/muscle-groups/:muscle_group_public_id", deps.Workout.V2DetailMuscleGroup)
	admin.PATCH("/muscle-groups/:muscle_group_public_id", deps.RoleMiddleware.Require("owner", "admin", "trainer"), deps.Workout.V2UpdateMuscleGroup)
	admin.DELETE("/muscle-groups/:muscle_group_public_id", deps.RoleMiddleware.Require("owner", "admin"), deps.Workout.V2DeleteMuscleGroup)

	admin.GET("/workout-templates", deps.Workout.V2ListWorkoutTemplates)
	admin.POST("/workout-templates", deps.RoleMiddleware.Require("owner", "admin", "trainer"), deps.Workout.V2CreateWorkoutTemplate)
	admin.GET("/workout-templates/:template_public_id", deps.Workout.V2DetailWorkoutTemplate)
	admin.PATCH("/workout-templates/:template_public_id", deps.RoleMiddleware.Require("owner", "admin", "trainer"), deps.Workout.V2UpdateWorkoutTemplate)
	admin.DELETE("/workout-templates/:template_public_id", deps.RoleMiddleware.Require("owner", "admin"), deps.Workout.V2DeleteWorkoutTemplate)
	admin.GET("/workout-templates/:template_public_id/days", deps.Workout.V2TemplateDays)
	admin.PUT("/workout-templates/:template_public_id/days", deps.RoleMiddleware.Require("owner", "admin", "trainer"), deps.Workout.V2ReplaceTemplateDays)

	admin.POST("/members/:member_public_id/workout-schedules/assign-template", deps.RoleMiddleware.Require("owner", "admin", "trainer"), deps.Workout.V2AssignTemplate)
	admin.GET("/members/:member_public_id/workout-schedules/active", deps.Workout.V2ActiveMemberSchedule)
	admin.PUT("/members/:member_public_id/workout-schedules/:schedule_public_id/days", deps.RoleMiddleware.Require("owner", "admin", "trainer"), deps.Workout.V2ReplaceMemberScheduleDays)

	admin.GET("/workout-sessions", deps.Workout.V2AdminWorkoutSessions)
	admin.GET("/members/:member_public_id/workout-sessions", deps.Workout.V2AdminMemberWorkoutSessions)
	admin.POST("/members/:member_public_id/workout-sessions/manual", deps.RoleMiddleware.Require("owner", "admin", "trainer"), deps.Workout.V2ManualWorkoutSession)

	admin.GET("/members/:member_public_id/weight-logs", deps.Workout.V2AdminWeightLogs)
	admin.POST("/members/:member_public_id/weight-logs", deps.RoleMiddleware.Require("owner", "admin", "trainer"), deps.Workout.V2CreateAdminWeightLog)
	admin.PATCH("/members/:member_public_id/weight-logs/:weight_log_public_id", deps.RoleMiddleware.Require("owner", "admin", "trainer"), deps.Workout.V2UpdateAdminWeightLog)
	admin.DELETE("/members/:member_public_id/weight-logs/:weight_log_public_id", deps.RoleMiddleware.Require("owner", "admin"), deps.Workout.V2DeleteAdminWeightLog)

	admin.GET("/reports/member-progress/:member_public_id", deps.Report.V2MemberProgressReport)
	admin.GET("/reports/workout-consistency", deps.Report.V2WorkoutConsistencyReport)
	admin.GET("/reports/weight-progress", deps.Report.V2WeightProgressReport)

	member := api.Group("/member")
	member.GET("/workout-schedule/:qr_token", deps.Workout.V2PublicWorkoutSchedule)
	member.GET("/workout-schedule/:qr_token/today", deps.Workout.V2PublicTodayWorkoutSchedule)
	member.POST("/workout-sessions/start", deps.Workout.V2PublicStartWorkout)
	member.POST("/workout-sessions/:session_public_id/finish", deps.Workout.V2PublicFinishWorkout)
	member.GET("/workout-sessions/:qr_token", deps.Workout.V2PublicWorkoutHistory)
	member.POST("/weight-logs", deps.Workout.V2PublicCreateWeightLog)
	member.GET("/weight-logs/:qr_token", deps.Workout.V2PublicWeightLogs)
}
