package workout

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

func (c *Controller) V2ListMuscleGroups(ctx *gin.Context) {
	shared.List(ctx, c.service, MuscleGroupConfig)
}
func (c *Controller) V2DetailMuscleGroup(ctx *gin.Context) {
	shared.Detail(ctx, c.service, MuscleGroupConfig, "muscle_group_public_id")
}
func (c *Controller) V2CreateMuscleGroup(ctx *gin.Context) {
	var request CreateMuscleGroupRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.CreateMuscleGroup(ctx.Request.Context(), middleware.GetAuthUser(ctx), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "Muscle group berhasil dibuat", result)
}
func (c *Controller) V2UpdateMuscleGroup(ctx *gin.Context) {
	shared.UpdateDTO[UpdateMuscleGroupRequest](ctx, c.service, MuscleGroupConfig, "muscle_group_public_id")
}
func (c *Controller) V2DeleteMuscleGroup(ctx *gin.Context) {
	shared.ChangeStatus(ctx, c.service, MuscleGroupConfig, "muscle_group_public_id", false, "muscle_group.deactivated")
}

func (c *Controller) V2ListWorkoutTemplates(ctx *gin.Context) {
	shared.List(ctx, c.service, WorkoutTemplateConfig)
}
func (c *Controller) V2DetailWorkoutTemplate(ctx *gin.Context) {
	shared.Detail(ctx, c.service, WorkoutTemplateConfig, "template_public_id")
}
func (c *Controller) V2CreateWorkoutTemplate(ctx *gin.Context) {
	var request CreateWorkoutTemplateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.CreateWorkoutTemplate(ctx.Request.Context(), middleware.GetAuthUser(ctx), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "Workout template berhasil dibuat", result)
}
func (c *Controller) V2UpdateWorkoutTemplate(ctx *gin.Context) {
	shared.UpdateDTO[UpdateWorkoutTemplateRequest](ctx, c.service, WorkoutTemplateConfig, "template_public_id")
}
func (c *Controller) V2DeleteWorkoutTemplate(ctx *gin.Context) {
	shared.ChangeStatus(ctx, c.service, WorkoutTemplateConfig, "template_public_id", false, "workout_template.deactivated")
}
func (c *Controller) V2TemplateDays(ctx *gin.Context) {
	result, err := c.service.TemplateDays(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("template_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}
func (c *Controller) V2ReplaceTemplateDays(ctx *gin.Context) {
	var request ReplaceWorkoutDaysRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.ReplaceTemplateDays(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("template_public_id"), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Workout template days berhasil diperbarui", result)
}

func (c *Controller) V2AssignTemplate(ctx *gin.Context) {
	var request AssignWorkoutTemplateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.AssignTemplateToMember(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "Workout schedule berhasil diberikan ke member", result)
}
func (c *Controller) V2ActiveMemberSchedule(ctx *gin.Context) {
	result, err := c.service.ActiveMemberSchedule(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}
func (c *Controller) V2ReplaceMemberScheduleDays(ctx *gin.Context) {
	var request ReplaceWorkoutDaysRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.ReplaceMemberScheduleDays(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"), ctx.Param("schedule_public_id"), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Jadwal member berhasil diperbarui", result)
}

func (c *Controller) V2AdminWorkoutSessions(ctx *gin.Context) {
	shared.List(ctx, c.service, WorkoutSessionConfig)
}
func (c *Controller) V2AdminMemberWorkoutSessions(ctx *gin.Context) {
	result, err := c.service.AdminMemberWorkoutHistory(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"), pagination.Parse(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, result.Items, result.Meta)
}
func (c *Controller) V2ManualWorkoutSession(ctx *gin.Context) {
	var request ManualWorkoutSessionRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.CreateManualWorkoutSession(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "Workout session manual berhasil dibuat", result)
}

func (c *Controller) V2AdminWeightLogs(ctx *gin.Context) {
	result, err := c.service.AdminWeightLogs(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}
func (c *Controller) V2CreateAdminWeightLog(ctx *gin.Context) {
	var request CreateWeightLogRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.CreateAdminWeightLog(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "Berat badan berhasil dicatat", result)
}
func (c *Controller) V2UpdateAdminWeightLog(ctx *gin.Context) {
	var request CreateWeightLogRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.UpdateAdminWeightLog(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"), ctx.Param("weight_log_public_id"), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Weight log berhasil diperbarui", result)
}
func (c *Controller) V2DeleteAdminWeightLog(ctx *gin.Context) {
	if err := c.service.DeleteAdminWeightLog(ctx.Request.Context(), middleware.GetAuthUser(ctx), ctx.Param("member_public_id"), ctx.Param("weight_log_public_id")); err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Weight log berhasil dihapus", nil)
}

func (c *Controller) V2PublicWorkoutSchedule(ctx *gin.Context) {
	result, err := c.service.PublicWorkoutSchedule(ctx.Request.Context(), ctx.Param("qr_token"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}
func (c *Controller) V2PublicTodayWorkoutSchedule(ctx *gin.Context) {
	result, err := c.service.PublicTodayWorkoutSchedule(ctx.Request.Context(), ctx.Param("qr_token"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}
func (c *Controller) V2PublicStartWorkout(ctx *gin.Context) {
	var request StartWorkoutSessionRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.StartWorkoutSession(ctx.Request.Context(), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "Workout dimulai", result)
}
func (c *Controller) V2PublicFinishWorkout(ctx *gin.Context) {
	var request FinishWorkoutSessionRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.FinishWorkoutSession(ctx.Request.Context(), ctx.Param("session_public_id"), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Workout selesai", result)
}
func (c *Controller) V2PublicWorkoutHistory(ctx *gin.Context) {
	result, err := c.service.PublicWorkoutHistory(ctx.Request.Context(), ctx.Param("qr_token"), pagination.Parse(ctx))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, result.Items, result.Meta)
}
func (c *Controller) V2PublicCreateWeightLog(ctx *gin.Context) {
	var request PublicCreateWeightLogRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		response.ValidationError(ctx, err)
		return
	}
	result, err := c.service.CreatePublicWeightLog(ctx.Request.Context(), request)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "Berat badan berhasil dicatat", result)
}
func (c *Controller) V2PublicWeightLogs(ctx *gin.Context) {
	result, err := c.service.PublicWeightLogs(ctx.Request.Context(), ctx.Param("qr_token"))
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.OK(ctx, "Berhasil", result)
}
