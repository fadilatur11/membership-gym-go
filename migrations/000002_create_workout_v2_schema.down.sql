DROP TABLE IF EXISTS member_weight_logs;
DROP TABLE IF EXISTS member_workout_sessions;
DROP TABLE IF EXISTS member_workout_schedule_days;
DROP TABLE IF EXISTS member_workout_schedules;
DROP TABLE IF EXISTS workout_template_days;
DROP TABLE IF EXISTS workout_templates;
DROP TABLE IF EXISTS muscle_groups;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_id_fkey;
DROP INDEX IF EXISTS idx_users_gym_role_id;
ALTER TABLE users DROP COLUMN IF EXISTS role_id;
DROP TABLE IF EXISTS roles;
