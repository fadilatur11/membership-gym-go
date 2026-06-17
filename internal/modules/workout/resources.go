package workout

import "membership-gym/internal/modules/shared"

var (
	MuscleGroupConfig = shared.ResourceConfig{
		Table:        "muscle_groups",
		Select:       "public_id, name, code, description, is_active, created_at, updated_at",
		SearchSQL:    "(LOWER(name) LIKE $2 OR LOWER(code) LIKE $2)",
		Filters:      map[string]string{"is_active": "is_active"},
		UpdateFields: map[string]string{"name": "name", "code": "code", "description": "description", "is_active": "is_active"},
		StatusColumn: "is_active",
	}
	WorkoutTemplateConfig = shared.ResourceConfig{
		Table:        "workout_templates",
		Select:       "public_id, name, description, is_active, created_at, updated_at",
		SearchSQL:    "LOWER(name) LIKE $2",
		Filters:      map[string]string{"is_active": "is_active"},
		UpdateFields: map[string]string{"name": "name", "description": "description", "is_active": "is_active"},
		StatusColumn: "is_active",
	}
	WorkoutSessionConfig = shared.ResourceConfig{
		Table:        "member_workout_sessions",
		Select:       "public_id, member_id, workout_date, title, notes, started_at, ended_at, duration_seconds, status, source, created_at, updated_at",
		SearchSQL:    "LOWER(title) LIKE $2",
		Filters:      map[string]string{"status": "status", "source": "source"},
		StatusColumn: "status",
	}
)
