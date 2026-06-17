package models

const (
	QueryGenericListCountFormat = "SELECT COUNT(*) FROM %s WHERE %s"
	QueryGenericListFormat      = "SELECT %s FROM %s WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d"
	QueryGenericDetailFormat    = "SELECT %s FROM %s WHERE gym_id=$1 AND public_id=$2 LIMIT 1"
	QueryGenericChangeStatus    = "UPDATE %s SET %s=$3, updated_at=NOW() WHERE gym_id=$1 AND public_id=$2 RETURNING %s"
	QueryGenericUpdate          = "UPDATE %s SET %s WHERE gym_id=$1 AND public_id=$2 RETURNING %s"
	QueryAuditLogInsert         = `
		INSERT INTO audit_logs (public_id, gym_id, user_id, action, entity_type, entity_id, payload, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())`
	QueryResolveIDFormat = "SELECT id FROM %s WHERE gym_id=$1 AND public_id=$2"

	QueryLoginUserBase = `
		SELECT u.id AS user_id, u.public_id AS user_public_id, u.name, u.email, u.password_hash, COALESCE(r.code, u.role) AS role, u.is_active,
		       g.id AS gym_id, g.public_id AS gym_public_id, g.name AS gym_name, g.status AS gym_status
		FROM users u
		JOIN gyms g ON g.id=u.gym_id
		LEFT JOIN roles r ON r.id=u.role_id AND r.gym_id=u.gym_id
		WHERE %s LIMIT 1`

	QueryTouchLastLogin = "UPDATE users SET last_login_at=NOW(), updated_at=NOW() WHERE id=$1"

	QueryRefreshUser = `
		SELECT u.id AS user_id, u.public_id AS user_public_id, COALESCE(r.code, u.role) AS role, g.id AS gym_id, g.public_id AS gym_public_id
		FROM users u
		JOIN gyms g ON g.id=u.gym_id
		LEFT JOIN roles r ON r.id=u.role_id AND r.gym_id=u.gym_id
		WHERE u.id = split_part($1, ':', 1)::bigint AND u.is_active=true AND g.status='active'`

	QueryIssueTokenUser = `
		SELECT u.public_id AS user_public_id, u.name, u.email, COALESCE(r.code, u.role) AS role,
		       g.public_id AS gym_public_id, g.name AS gym_name
		FROM users u
		JOIN gyms g ON g.id=u.gym_id
		LEFT JOIN roles r ON r.id=u.role_id AND r.gym_id=u.gym_id
		WHERE u.id=$1 AND u.gym_id=$2 AND u.is_active=true AND g.status='active'`

	QueryCountUsersByEmail = "SELECT COUNT(*) FROM users WHERE LOWER(email)=$1"

	QueryGoogleUserBySubOrEmail = `
		SELECT id AS user_id, gym_id, google_sub
		FROM users
		WHERE google_sub=$1 OR LOWER(email)=$2
		ORDER BY CASE WHEN google_sub=$1 THEN 0 ELSE 1 END, id ASC
		LIMIT 1`

	QueryLinkGoogleUser = "UPDATE users SET auth_provider='google', google_sub=$2, avatar_url=$3, updated_at=NOW() WHERE id=$1"

	QueryInsertGymReturningID = `
		INSERT INTO gyms(public_id,name,phone,email,address,timezone,currency,status,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,'active',NOW(),NOW())
		RETURNING id`

	QueryInsertOwnerUserReturningID = `
		INSERT INTO users(public_id,gym_id,name,email,password_hash,role,role_id,is_active,auth_provider,google_sub,avatar_url,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,'owner',$6,true,$7,$8,$9,NOW(),NOW())
		RETURNING id`

	QueryInsertDefaultRole = `
		INSERT INTO roles(public_id,gym_id,name,code,description,is_active,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,true,NOW(),NOW())
		ON CONFLICT (gym_id, code) DO NOTHING`

	QueryOwnerRoleID = "SELECT id FROM roles WHERE gym_id=$1 AND code='owner'"

	QueryBasicSaasPlan = "SELECT id, duration_days FROM saas_plans WHERE code='basic' AND is_active=true"

	QueryInsertGymSubscriptionReturningID = `
		INSERT INTO gym_subscriptions(public_id,gym_id,saas_plan_id,start_date,end_date,status,auto_renew,source,created_by,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,'trialing',false,'owner_register',$6,NOW(),NOW())
		RETURNING id`

	QueryInsertGymSubscriptionFreePayment = `
		INSERT INTO gym_subscription_payments(public_id,gym_id,gym_subscription_id,invoice_no,payment_method,amount,currency,status,paid_at,notes,created_by,created_at,updated_at)
		VALUES($1,$2,$3,$4,'free',0,$5,'paid',NOW(),'Basic trial 30 hari',$6,NOW(),NOW())`

	QueryProfile = `
		SELECT u.public_id, u.name, u.email, COALESCE(r.code, u.role) AS role, u.is_active, u.last_login_at,
		       g.public_id AS gym_public_id, g.name AS gym_name
		FROM users u
		JOIN gyms g ON g.id=u.gym_id
		LEFT JOIN roles r ON r.id=u.role_id AND r.gym_id=u.gym_id
		WHERE u.id=$1 AND u.gym_id=$2`

	QueryUpdateGymProfile = `
		UPDATE gyms SET name=COALESCE($2,name), phone=$3, email=$4, address=$5, timezone=COALESCE($6,timezone), currency=COALESCE($7,currency), updated_at=NOW()
		WHERE id=$1
		RETURNING public_id, name, phone, email, address, timezone, currency, status`

	QueryGymProfile = "SELECT public_id, name, phone, email, address, timezone, currency, status FROM gyms WHERE id=$1"

	QueryListSaasPlans = `
		SELECT public_id, code, name, description, duration_days, price, currency, billing_cycle, features, limits, is_active
		FROM saas_plans
		WHERE is_active=true
		ORDER BY price ASC, id ASC`

	QueryCurrentGymSubscription = `
		SELECT gs.public_id, gs.start_date, gs.end_date, gs.status, gs.auto_renew, gs.source,
		       p.public_id AS plan_public_id, p.code AS plan_code, p.name AS plan_name, p.price, p.currency,
		       p.duration_days, p.billing_cycle, p.features, p.limits,
		       pay.public_id AS payment_public_id, pay.invoice_no, pay.payment_method, pay.amount AS payment_amount, pay.status AS payment_status, pay.paid_at
		FROM gym_subscriptions gs
		JOIN saas_plans p ON p.id=gs.saas_plan_id
		LEFT JOIN LATERAL (
			SELECT public_id, invoice_no, payment_method, amount, status, paid_at
			FROM gym_subscription_payments
			WHERE gym_subscription_id=gs.id
			ORDER BY created_at DESC
			LIMIT 1
		) pay ON true
		WHERE gs.gym_id=$1
		ORDER BY gs.created_at DESC
		LIMIT 1`

	QuerySaasPlanByPublicID = `
		SELECT id, public_id, code, name, duration_days, price, currency, billing_cycle, features, limits
		FROM saas_plans
		WHERE public_id=$1 AND is_active=true
		LIMIT 1`

	QueryCancelCurrentGymSubscriptions = `
		UPDATE gym_subscriptions
		SET status='cancelled', updated_at=NOW()
		WHERE gym_id=$1 AND status IN ('trialing', 'active', 'pending', 'past_due')`

	QueryInsertOwnerGymSubscription = `
		INSERT INTO gym_subscriptions(public_id,gym_id,saas_plan_id,start_date,end_date,status,auto_renew,source,created_by,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,'owner',$8,NOW(),NOW())
		RETURNING id, public_id, start_date, end_date, status, auto_renew, source`

	QueryGymSubscriptionPaymentSequence = "SELECT COUNT(*)+1 FROM gym_subscription_payments WHERE gym_id=$1 AND DATE(created_at)=$2"

	QueryInsertGymSubscriptionPayment = `
		INSERT INTO gym_subscription_payments(public_id,gym_id,gym_subscription_id,invoice_no,payment_method,amount,currency,status,paid_at,notes,created_by,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW(),NOW())
		RETURNING public_id, invoice_no, payment_method, amount, currency, status, paid_at, notes`

	QueryActivePlanFeatures = `
		SELECT p.code, p.features
		FROM gym_subscriptions gs
		JOIN saas_plans p ON p.id=gs.saas_plan_id
		WHERE gs.gym_id=$1
		  AND gs.status IN ('trialing', 'active')
		  AND gs.start_date <= CURRENT_DATE
		  AND gs.end_date >= CURRENT_DATE
		ORDER BY gs.created_at DESC
		LIMIT 1`

	QueryRoleIDByCode = "SELECT id FROM roles WHERE gym_id=$1 AND code=$2"

	QueryInsertUser = `
		INSERT INTO users(public_id,gym_id,name,email,password_hash,role,role_id,is_active,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,true,NOW(),NOW())
		RETURNING public_id, name, email, role, is_active, created_at, updated_at`

	QueryInsertMember = `
		INSERT INTO members(public_id,gym_id,member_code,full_name,phone,email,gender,birth_date,address,emergency_contact_name,emergency_contact_phone,joined_at,status,notes,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,'active',$13,NOW(),NOW())
		RETURNING public_id, member_code, full_name, phone, email, gender, birth_date, address, emergency_contact_name, emergency_contact_phone, joined_at, status, notes, created_at, updated_at`

	QueryInsertMembershipPackage = `
		INSERT INTO membership_packages(public_id,gym_id,name,duration_days,price,description,is_active,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,true,NOW(),NOW())
		RETURNING public_id, name, duration_days, price, description, is_active, created_at, updated_at`

	QueryInsertExpenseCategory = `
		INSERT INTO expense_categories(public_id,gym_id,name,is_active,created_at,updated_at)
		VALUES($1,$2,$3,true,NOW(),NOW())
		RETURNING public_id, name, is_active, created_at, updated_at`

	QueryInsertExpense = `
		INSERT INTO expenses(public_id,gym_id,expense_category_id,title,description,amount,expense_date,status,created_by,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW(),NOW())
		RETURNING public_id, expense_category_id, title, description, amount, expense_date, status, created_at, updated_at`

	QueryInsertReminderRule = `
		INSERT INTO reminder_rules(public_id,gym_id,name,days_before_expiry,channel,message_template,is_active,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,true,NOW(),NOW())
		RETURNING public_id, name, days_before_expiry, channel, message_template, is_active, created_at, updated_at`

	QueryMemberForQR = "SELECT id, public_id, full_name FROM members WHERE gym_id=$1 AND public_id=$2"

	QueryActiveQR = `
		SELECT q.public_id, $3::uuid AS member_public_id, $4::text AS member_name, q.qr_token,
		       $5::text || '/' || q.qr_token AS qr_content, q.status, q.generated_at
		FROM member_qrcodes q WHERE q.gym_id=$1 AND q.member_id=$2 AND q.status='active' LIMIT 1`

	QueryRevokeActiveQRByMemberID = "UPDATE member_qrcodes SET status='revoked', revoked_at=NOW(), updated_at=NOW() WHERE gym_id=$1 AND member_id=$2 AND status='active'"

	QueryInsertMemberQR = `
		INSERT INTO member_qrcodes(public_id,gym_id,member_id,qr_token,status,generated_at,created_at,updated_at)
		VALUES($1,$2,$3,$4,'active',NOW(),NOW(),NOW())
		RETURNING public_id, $5::uuid AS member_public_id, $6::text AS member_name, qr_token, $7::text || '/' || qr_token AS qr_content, status, generated_at`

	QueryRevokeQR = `
		UPDATE member_qrcodes q SET status='revoked', revoked_at=NOW(), updated_at=NOW()
		FROM members m
		WHERE q.gym_id=$1 AND q.member_id=m.id AND m.public_id=$2 AND q.status='active'
		RETURNING q.public_id, q.qr_token, q.status, q.revoked_at`

	QueryCheckinQRBase = `
		SELECT q.gym_id, q.member_id, m.public_id AS member_public_id, m.member_code, m.full_name AS member_name, m.status AS member_status,
		       g.name AS gym_name, g.timezone
		FROM member_qrcodes q
		JOIN members m ON m.id=q.member_id AND m.gym_id=q.gym_id
		JOIN gyms g ON g.id=q.gym_id
		WHERE q.qr_token=$1 AND q.status='active'%s LIMIT 1`

	QueryActiveMemberSubscription = `
		SELECT id, public_id, start_date, end_date FROM subscriptions
		WHERE gym_id=$1 AND member_id=$2 AND status='active' AND start_date <= $3 AND end_date >= $3
		ORDER BY end_date DESC LIMIT 1`

	QueryValidCheckinToday = "SELECT checkin_at FROM member_checkins WHERE gym_id=$1 AND member_id=$2 AND checkin_date=$3 AND status='valid' LIMIT 1"

	QueryManualCheckinMember = "SELECT id, public_id AS member_public_id, member_code, full_name AS member_name, status AS member_status FROM members WHERE gym_id=$1 AND public_id=$2"

	QueryManualActiveSubscription = "SELECT id, end_date FROM subscriptions WHERE gym_id=$1 AND member_id=$2 AND status='active' AND start_date <= $3 AND end_date >= $3 ORDER BY end_date DESC LIMIT 1"

	QueryInsertMemberCheckin = `
		INSERT INTO member_checkins(public_id,gym_id,member_id,subscription_id,checkin_date,checkin_at,source,status,scanned_by,notes,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW(),NOW())`

	QueryPublicQRIdentity = `
		SELECT m.id AS member_id, m.full_name AS member_name, m.member_code, g.id AS gym_id, g.name AS gym_name,
		       COALESCE((SELECT status FROM subscriptions s WHERE s.gym_id=q.gym_id AND s.member_id=q.member_id ORDER BY end_date DESC LIMIT 1),'expired') AS membership_status,
		       (SELECT end_date FROM subscriptions s WHERE s.gym_id=q.gym_id AND s.member_id=q.member_id ORDER BY end_date DESC LIMIT 1) AS subscription_end_date
		FROM member_qrcodes q JOIN members m ON m.id=q.member_id JOIN gyms g ON g.id=q.gym_id
		WHERE q.qr_token=$1 AND q.status='active' LIMIT 1`

	QueryPublicMemberStatus = `
		SELECT m.full_name AS member_name, m.member_code, m.status AS member_status, g.name AS gym_name, g.timezone,
		       sub.status AS membership_status, sub.start_date AS subscription_start_date, sub.end_date AS subscription_end_date
		FROM member_qrcodes q
		JOIN members m ON m.id=q.member_id
		JOIN gyms g ON g.id=q.gym_id
		LEFT JOIN LATERAL (SELECT status,start_date,end_date FROM subscriptions s WHERE s.gym_id=q.gym_id AND s.member_id=q.member_id ORDER BY end_date DESC LIMIT 1) sub ON true
		WHERE q.qr_token=$1 AND q.status='active' LIMIT 1`

	QueryQRCodeMemberIDs = "SELECT gym_id, member_id FROM member_qrcodes WHERE qr_token=$1 AND status='active' LIMIT 1"

	QueryCountMemberCheckins = "SELECT COUNT(*) FROM member_checkins WHERE gym_id=$1 AND member_id=$2"

	QueryPublicCheckinHistory = `
		SELECT checkin_date, checkin_at, status, source
		FROM member_checkins
		WHERE gym_id=$1 AND member_id=$2
		ORDER BY checkin_at DESC LIMIT $3 OFFSET $4`

	QueryDashboard = `
		SELECT
		  (SELECT COUNT(*) FROM members WHERE gym_id=$1 AND status='active') AS total_active_members,
		  (SELECT COUNT(DISTINCT member_id) FROM subscriptions WHERE gym_id=$1 AND status='expired') AS total_expired_members,
		  (SELECT COALESCE(SUM(final_amount),0) FROM payments WHERE gym_id=$1 AND status='paid' AND date_trunc('month', paid_at)=date_trunc('month', NOW())) AS income_this_month,
		  (SELECT COALESCE(SUM(amount),0) FROM expenses WHERE gym_id=$1 AND status='approved' AND date_trunc('month', expense_date)=date_trunc('month', NOW())) AS expense_this_month,
		  (SELECT COUNT(*) FROM member_checkins WHERE gym_id=$1 AND status='valid' AND checkin_date=CURRENT_DATE) AS checkins_today,
		  (SELECT COUNT(*) FROM payments WHERE gym_id=$1 AND status='pending') AS payment_pending,
		  (SELECT COUNT(*) FROM reminder_logs WHERE gym_id=$1 AND status='failed') AS reminder_failed`

	QueryMemberPackageForSubscription = `
		SELECT m.id AS member_id, m.full_name AS member_name, p.id AS package_id, p.name AS package_name, p.duration_days
		FROM members m JOIN membership_packages p ON p.gym_id=m.gym_id
		WHERE m.gym_id=$1 AND m.public_id=$2 AND p.public_id=$3 AND m.status='active' AND p.is_active=true`

	QueryInsertSubscription = `
		INSERT INTO subscriptions(public_id,gym_id,member_id,membership_package_id,start_date,end_date,status,source,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,'active',$7,NOW(),NOW())
		RETURNING public_id, start_date, end_date, status, source, created_at, updated_at`

	QueryMemberPackageForPayment = `
		SELECT m.id AS member_id, m.full_name AS member_name, p.id AS package_id, p.name AS package_name, p.duration_days, p.price
		FROM members m JOIN membership_packages p ON p.gym_id=m.gym_id
		WHERE m.gym_id=$1 AND m.public_id=$2 AND p.public_id=$3 AND m.status='active' AND p.is_active=true`

	QueryInsertPaymentSubscription = `
		INSERT INTO subscriptions(public_id,gym_id,member_id,membership_package_id,start_date,end_date,status,source,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6,'active','manual',NOW(),NOW())
		RETURNING id, public_id, start_date, end_date, status`

	QueryPaymentSequence = "SELECT COUNT(*)+1 FROM payments WHERE gym_id=$1 AND DATE(paid_at)=$2"

	QueryInsertPayment = `
		INSERT INTO payments(public_id,gym_id,member_id,subscription_id,invoice_no,payment_type,payment_method,package_price,discount_amount,final_amount,status,paid_at,notes,created_by,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,'membership',$6,$7,$8,$9,$10,$11,$12,$13,NOW(),NOW())
		RETURNING public_id, invoice_no, package_price, discount_amount, final_amount, payment_method, status, paid_at, notes`

	QueryPaymentSubscriptionID = "SELECT subscription_id FROM payments WHERE gym_id=$1 AND public_id=$2"

	QueryUpdatePaymentStatus = "UPDATE payments SET status=$3, updated_at=NOW() WHERE gym_id=$1 AND public_id=$2 RETURNING public_id, invoice_no, status"

	QueryCancelSubscriptionByID = "UPDATE subscriptions SET status='cancelled', updated_at=NOW() WHERE gym_id=$1 AND id=$2"

	QueryPublicPackages = `
		SELECT p.name, p.duration_days, p.price, p.description
		FROM member_qrcodes q
		JOIN membership_packages p ON p.gym_id=q.gym_id
		WHERE q.qr_token=$1 AND q.status='active' AND p.is_active=true
		ORDER BY p.price ASC`

	QueryPublicGymContact = `
		SELECT g.name AS gym_name, g.phone, g.email, g.address
		FROM member_qrcodes q
		JOIN gyms g ON g.id=q.gym_id
		WHERE q.qr_token=$1 AND q.status='active'
		LIMIT 1`

	QueryInsertMuscleGroup = `
		INSERT INTO muscle_groups(public_id,gym_id,name,code,description,is_active,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,true,NOW(),NOW())
		RETURNING public_id, name, code, description, is_active, created_at, updated_at`

	QueryInsertWorkoutTemplate = `
		INSERT INTO workout_templates(public_id,gym_id,name,description,is_active,created_by,created_at,updated_at)
		VALUES($1,$2,$3,$4,true,$5,NOW(),NOW())
		RETURNING public_id, name, description, is_active, created_at, updated_at`
)
