package models

import "time"

type WorkoutSessionStatus string
type WorkoutSource string

const (
	WorkoutSessionStatusStarted   WorkoutSessionStatus = "started"
	WorkoutSessionStatusPaused    WorkoutSessionStatus = "paused"
	WorkoutSessionStatusCompleted WorkoutSessionStatus = "completed"
	WorkoutSessionStatusCancelled WorkoutSessionStatus = "cancelled"
	WorkoutSessionStatusManual    WorkoutSessionStatus = "manual"

	WorkoutSourceMember  WorkoutSource = "member"
	WorkoutSourceAdmin   WorkoutSource = "admin"
	WorkoutSourceTrainer WorkoutSource = "trainer"
)

type MuscleGroup struct {
	BaseTenantModel
	Name, Code, Description string
	IsActive                bool
}

type WorkoutTemplate struct {
	BaseTenantModel
	Name, Description string
	IsActive          bool
	CreatedBy         *int64
}

type WorkoutTemplateDay struct {
	BaseTenantModel
	WorkoutTemplateID  int64
	MuscleGroupID      *int64
	DayOfWeek          int
	Title, Description string
	IsRestDay          bool
}

type MemberWorkoutSchedule struct {
	BaseTenantModel
	MemberID           int64
	WorkoutTemplateID  *int64
	Name, Description  string
	IsActive, IsCustom bool
	CreatedBy          *int64
}

type MemberWorkoutScheduleDay struct {
	BaseTenantModel
	MemberWorkoutScheduleID int64
	MemberID                int64
	MuscleGroupID           *int64
	DayOfWeek               int
	Title, Description      string
	IsRestDay               bool
}

type MemberWorkoutSession struct {
	BaseTenantModel
	MemberID                   int64
	MemberCheckinID            *int64
	MemberWorkoutScheduleDayID *int64
	MuscleGroupID              *int64
	WorkoutDate                time.Time
	Title, Notes               string
	StartedAt, EndedAt         *time.Time
	DurationSeconds            int
	Status, Source             string
	CreatedBy                  *int64
}

type MemberWeightLog struct {
	BaseTenantModel
	MemberID      int64
	MeasuredDate  time.Time
	WeightKG      float64
	Notes, Source string
	CreatedBy     *int64
}
