package workout

type CreateMuscleGroupRequest struct {
	Name        string `json:"name" binding:"required" example:"Chest"`
	Code        string `json:"code" binding:"required" example:"chest"`
	Description string `json:"description,omitempty" example:"Otot dada"`
}

type UpdateMuscleGroupRequest struct {
	Name        string `json:"name,omitempty" example:"Chest"`
	Code        string `json:"code,omitempty" example:"chest"`
	Description string `json:"description,omitempty" example:"Fokus otot dada"`
	IsActive    *bool  `json:"is_active,omitempty" example:"true"`
}

type CreateWorkoutTemplateRequest struct {
	Name        string `json:"name" binding:"required" example:"Beginner 5 Days"`
	Description string `json:"description,omitempty" example:"Jadwal latihan pemula 5 hari"`
}

type UpdateWorkoutTemplateRequest struct {
	Name        string `json:"name,omitempty" example:"Beginner 6 Days"`
	Description string `json:"description,omitempty" example:"Updated description"`
	IsActive    *bool  `json:"is_active,omitempty" example:"true"`
}

type ReplaceWorkoutDaysRequest struct {
	Days []WorkoutDayRequest `json:"days" binding:"required"`
}

type WorkoutDayRequest struct {
	DayOfWeek           int     `json:"day_of_week" binding:"required,min=1,max=7" example:"1"`
	MuscleGroupPublicID *string `json:"muscle_group_public_id,omitempty" example:"muscle-group-public-id"`
	Title               string  `json:"title" binding:"required" example:"Chest Day"`
	Description         string  `json:"description,omitempty" example:"Fokus latihan dada"`
	IsRestDay           bool    `json:"is_rest_day" example:"false"`
}

type AssignWorkoutTemplateRequest struct {
	WorkoutTemplatePublicID string `json:"workout_template_public_id" binding:"required" example:"template-public-id"`
	Name                    string `json:"name,omitempty" example:"Jadwal Budi - Beginner 5 Days"`
}

type ManualWorkoutSessionRequest struct {
	WorkoutDate         string `json:"workout_date" binding:"required" example:"2026-06-16"`
	MuscleGroupPublicID string `json:"muscle_group_public_id,omitempty" example:"chest-public-id"`
	Title               string `json:"title" binding:"required" example:"Chest Day"`
	DurationSeconds     int    `json:"duration_seconds" binding:"required,min=1" example:"3600"`
	Notes               string `json:"notes,omitempty" example:"Input manual oleh trainer"`
}

type StartWorkoutSessionRequest struct {
	QRToken             string `json:"qr_token" binding:"required" example:"qr-token-random"`
	ScheduleDayPublicID string `json:"schedule_day_public_id,omitempty" example:"schedule-day-public-id"`
	MuscleGroupPublicID string `json:"muscle_group_public_id,omitempty" example:"chest-public-id"`
	Title               string `json:"title,omitempty" example:"Chest Manual Workout"`
}

type FinishWorkoutSessionRequest struct {
	QRToken string `json:"qr_token" binding:"required" example:"qr-token-random"`
	Notes   string `json:"notes,omitempty" example:"Latihan selesai, cukup berat"`
}

type CreateWeightLogRequest struct {
	MeasuredDate string  `json:"measured_date" binding:"required" example:"2026-06-16"`
	WeightKG     float64 `json:"weight_kg" binding:"required,gt=0" example:"68.5"`
	Notes        string  `json:"notes,omitempty" example:"Ditimbang setelah latihan"`
}

type PublicCreateWeightLogRequest struct {
	QRToken      string  `json:"qr_token" binding:"required" example:"qr-token-random"`
	MeasuredDate string  `json:"measured_date" binding:"required" example:"2026-06-16"`
	WeightKG     float64 `json:"weight_kg" binding:"required,gt=0" example:"68.5"`
	Notes        string  `json:"notes,omitempty" example:"Setelah latihan"`
}
