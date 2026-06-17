package models

const (
	QueryV2ResolveQR = `
		SELECT q.gym_id, q.member_id, m.public_id AS member_public_id, m.full_name AS member_name,
		       m.member_code, m.status AS member_status, g.name AS gym_name, g.timezone
		FROM member_qrcodes q
		JOIN members m ON m.id=q.member_id AND m.gym_id=q.gym_id
		JOIN gyms g ON g.id=q.gym_id
		WHERE q.qr_token=$1 AND q.status='active'
		LIMIT 1`

	QueryTemplateDays = `
		SELECT d.public_id, d.day_of_week, d.title, d.description, d.is_rest_day,
		       mg.public_id AS muscle_group_public_id, mg.name AS muscle_group_name, mg.code AS muscle_group_code
		FROM workout_template_days d
		LEFT JOIN muscle_groups mg ON mg.id=d.muscle_group_id AND mg.gym_id=d.gym_id
		JOIN workout_templates wt ON wt.id=d.workout_template_id AND wt.gym_id=d.gym_id
		WHERE d.gym_id=$1 AND wt.public_id=$2
		ORDER BY d.day_of_week`

	QueryWorkoutTemplateID  = "SELECT id FROM workout_templates WHERE gym_id=$1 AND public_id=$2"
	QueryDeleteTemplateDays = "DELETE FROM workout_template_days WHERE gym_id=$1 AND workout_template_id=$2"

	QueryInsertTemplateDay = `
		INSERT INTO workout_template_days(public_id,gym_id,workout_template_id,muscle_group_id,day_of_week,title,description,is_rest_day,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,NOW(),NOW())`

	QueryActiveMuscleGroupID             = "SELECT id FROM muscle_groups WHERE gym_id=$1 AND public_id=$2 AND is_active=true"
	QueryActiveMemberForWorkout          = "SELECT id, full_name FROM members WHERE gym_id=$1 AND public_id=$2 AND status='active'"
	QueryActiveWorkoutTemplate           = "SELECT id, name FROM workout_templates WHERE gym_id=$1 AND public_id=$2 AND is_active=true"
	QueryDeactivateActiveMemberSchedules = "UPDATE member_workout_schedules SET is_active=false, updated_at=NOW() WHERE gym_id=$1 AND member_id=$2 AND is_active=true"

	QueryInsertMemberWorkoutSchedule = `
		INSERT INTO member_workout_schedules(public_id,gym_id,member_id,workout_template_id,name,is_active,is_custom,created_by,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,true,false,$6,NOW(),NOW()) RETURNING id`

	QueryCopyTemplateDaysToMemberSchedule = `
		INSERT INTO member_workout_schedule_days(public_id,gym_id,member_workout_schedule_id,member_id,muscle_group_id,day_of_week,title,description,is_rest_day,created_at,updated_at)
		SELECT gen_random_uuid(), gym_id, $1, $2, muscle_group_id, day_of_week, title, description, is_rest_day, NOW(), NOW()
		FROM workout_template_days WHERE gym_id=$3 AND workout_template_id=$4`

	QueryActiveMemberSchedule = `
		SELECT s.public_id, m.public_id AS member_public_id, m.full_name AS member_name, s.name, s.description, s.is_active, s.is_custom
		FROM member_workout_schedules s
		JOIN members m ON m.id=s.member_id AND m.gym_id=s.gym_id
		WHERE s.gym_id=$1 AND m.public_id=$2 AND s.is_active=true`

	QueryMemberScheduleDays = `
		SELECT d.public_id AS schedule_day_public_id, d.day_of_week, d.title, d.description, d.is_rest_day,
		       mg.public_id AS muscle_group_public_id, mg.name AS muscle_group_name, mg.code AS muscle_group_code
		FROM member_workout_schedule_days d
		JOIN member_workout_schedules s ON s.id=d.member_workout_schedule_id AND s.gym_id=d.gym_id
		LEFT JOIN muscle_groups mg ON mg.id=d.muscle_group_id AND mg.gym_id=d.gym_id
		WHERE d.gym_id=$1 AND s.public_id=$2
		ORDER BY d.day_of_week`

	QueryMemberAndScheduleID = `
		SELECT m.id, s.id
		FROM members m JOIN member_workout_schedules s ON s.member_id=m.id AND s.gym_id=m.gym_id
		WHERE m.gym_id=$1 AND m.public_id=$2 AND s.public_id=$3`

	QueryDeleteMemberScheduleDays = "DELETE FROM member_workout_schedule_days WHERE gym_id=$1 AND member_workout_schedule_id=$2"

	QueryInsertMemberScheduleDay = `
		INSERT INTO member_workout_schedule_days(public_id,gym_id,member_workout_schedule_id,member_id,muscle_group_id,day_of_week,title,description,is_rest_day,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),NOW())`

	QueryMarkMemberScheduleCustom = "UPDATE member_workout_schedules SET is_custom=true, updated_at=NOW() WHERE gym_id=$1 AND id=$2"
	QueryPublicActiveSchedule     = "SELECT public_id, name FROM member_workout_schedules WHERE gym_id=$1 AND member_id=$2 AND is_active=true LIMIT 1"

	QueryPublicTodaySchedule = `
		SELECT d.public_id AS schedule_day_public_id, d.day_of_week, d.title, d.description, d.is_rest_day,
		       mg.public_id AS muscle_group_public_id, mg.name AS muscle_group_name, mg.code AS muscle_group_code
		FROM member_workout_schedule_days d
		JOIN member_workout_schedules s ON s.id=d.member_workout_schedule_id AND s.gym_id=d.gym_id AND s.is_active=true
		LEFT JOIN muscle_groups mg ON mg.id=d.muscle_group_id AND mg.gym_id=d.gym_id
		WHERE d.gym_id=$1 AND d.member_id=$2 AND d.day_of_week=$3
		LIMIT 1`

	QueryScheduleDayForWorkout = `
		SELECT d.id, d.title, d.muscle_group_id
		FROM member_workout_schedule_days d
		WHERE d.gym_id=$1 AND d.member_id=$2 AND d.public_id=$3`

	QueryValidCheckinIDForWorkout = "SELECT id FROM member_checkins WHERE gym_id=$1 AND member_id=$2 AND checkin_date=$3 AND status='valid' LIMIT 1"

	QueryInsertWorkoutSessionStart = `
		INSERT INTO member_workout_sessions(public_id,gym_id,member_id,member_checkin_id,member_workout_schedule_day_id,muscle_group_id,workout_date,title,started_at,duration_seconds,status,source,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,0,'started','member',NOW(),NOW())
		RETURNING public_id AS workout_session_public_id, title, started_at, status`

	QueryStartedWorkoutSession = "SELECT started_at FROM member_workout_sessions WHERE gym_id=$1 AND member_id=$2 AND public_id=$3 AND status='started'"

	QueryFinishWorkoutSession = `
		UPDATE member_workout_sessions
		SET ended_at=$4, duration_seconds=$5, status='completed', notes=$6, updated_at=NOW()
		WHERE gym_id=$1 AND member_id=$2 AND public_id=$3
		RETURNING public_id AS workout_session_public_id, title, started_at, ended_at, duration_seconds, status`

	QueryCountWorkoutSessions = "SELECT COUNT(*) FROM member_workout_sessions WHERE gym_id=$1 AND member_id=$2"

	QueryWorkoutHistory = `
		SELECT ws.public_id AS workout_session_public_id, ws.workout_date, ws.title, mg.name AS muscle_group,
		       ws.duration_seconds, ws.status, ws.started_at, ws.ended_at
		FROM member_workout_sessions ws
		LEFT JOIN muscle_groups mg ON mg.id=ws.muscle_group_id AND mg.gym_id=ws.gym_id
		WHERE ws.gym_id=$1 AND ws.member_id=$2
		ORDER BY ws.workout_date DESC, ws.created_at DESC LIMIT $3 OFFSET $4`

	QueryInsertManualWorkoutSession = `
		INSERT INTO member_workout_sessions(public_id,gym_id,member_id,muscle_group_id,workout_date,title,notes,duration_seconds,status,source,created_by,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,'manual',$9,$10,NOW(),NOW())
		RETURNING public_id, workout_date, title, duration_seconds, status`

	QueryWeightLogs = `
		SELECT public_id, measured_date, weight_kg, notes, source, created_at
		FROM member_weight_logs
		WHERE gym_id=$1 AND member_id=$2
		ORDER BY measured_date ASC, created_at ASC`

	QueryInsertAdminWeightLog = `
		INSERT INTO member_weight_logs(public_id,gym_id,member_id,measured_date,weight_kg,notes,source,created_by,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,NOW(),NOW())
		RETURNING public_id, measured_date, weight_kg, notes, source, created_at`

	QueryUpdateAdminWeightLog = `
		UPDATE member_weight_logs SET measured_date=$4, weight_kg=$5, notes=$6, updated_at=NOW()
		WHERE gym_id=$1 AND member_id=$2 AND public_id=$3
		RETURNING public_id, measured_date, weight_kg, notes, source, created_at`

	QueryDeleteWeightLog = "DELETE FROM member_weight_logs WHERE gym_id=$1 AND member_id=$2 AND public_id=$3"

	QueryInsertPublicWeightLog = `
		INSERT INTO member_weight_logs(public_id,gym_id,member_id,measured_date,weight_kg,notes,source,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,'member',NOW(),NOW())
		RETURNING public_id, measured_date, weight_kg, notes`

	QueryMemberProgressReport = `
		SELECT m.full_name AS member_name,
		       (SELECT COUNT(*) FROM member_workout_sessions ws WHERE ws.gym_id=m.gym_id AND ws.member_id=m.id) AS total_workout_sessions,
		       (SELECT COALESCE(SUM(duration_seconds),0) FROM member_workout_sessions ws WHERE ws.gym_id=m.gym_id AND ws.member_id=m.id) AS total_duration_seconds,
		       (SELECT COALESCE(AVG(duration_seconds),0)::bigint FROM member_workout_sessions ws WHERE ws.gym_id=m.gym_id AND ws.member_id=m.id AND duration_seconds > 0) AS average_duration_seconds,
		       (SELECT weight_kg FROM member_weight_logs wl WHERE wl.gym_id=m.gym_id AND wl.member_id=m.id ORDER BY measured_date ASC LIMIT 1) AS first_weight,
		       (SELECT weight_kg FROM member_weight_logs wl WHERE wl.gym_id=m.gym_id AND wl.member_id=m.id ORDER BY measured_date DESC LIMIT 1) AS latest_weight,
		       (SELECT mg.name FROM member_workout_sessions ws JOIN muscle_groups mg ON mg.id=ws.muscle_group_id WHERE ws.gym_id=m.gym_id AND ws.member_id=m.id GROUP BY mg.name ORDER BY COUNT(*) DESC LIMIT 1) AS most_trained_muscle_group,
		       (SELECT MAX(workout_date) FROM member_workout_sessions ws WHERE ws.gym_id=m.gym_id AND ws.member_id=m.id) AS last_workout_date
		FROM members m WHERE m.gym_id=$1 AND m.public_id=$2 LIMIT 1`
)
