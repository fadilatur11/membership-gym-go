package modules

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"membership-gym/config"
	auditmodule "membership-gym/internal/modules/audit"
	authmodule "membership-gym/internal/modules/auth"
	checkinmodule "membership-gym/internal/modules/checkin"
	expensemodule "membership-gym/internal/modules/expense"
	expensecategorymodule "membership-gym/internal/modules/expense_category"
	gymmodule "membership-gym/internal/modules/gym"
	membermodule "membership-gym/internal/modules/member"
	membershippackagemodule "membership-gym/internal/modules/membership_package"
	paymentmodule "membership-gym/internal/modules/payment"
	qrcodemodule "membership-gym/internal/modules/qrcode"
	reminderlogmodule "membership-gym/internal/modules/reminder_log"
	reminderrulemodule "membership-gym/internal/modules/reminder_rule"
	reportmodule "membership-gym/internal/modules/report"
	subscriptionmodule "membership-gym/internal/modules/subscription"
	usermodule "membership-gym/internal/modules/user"
	workoutmodule "membership-gym/internal/modules/workout"
)

type Controllers struct {
	Auth              *authmodule.Controller
	Gym               *gymmodule.Controller
	Subscription      *subscriptionmodule.Controller
	User              *usermodule.Controller
	Member            *membermodule.Controller
	QRCode            *qrcodemodule.Controller
	MembershipPackage *membershippackagemodule.Controller
	Payment           *paymentmodule.Controller
	Checkin           *checkinmodule.Controller
	ExpenseCategory   *expensecategorymodule.Controller
	Expense           *expensemodule.Controller
	ReminderRule      *reminderrulemodule.Controller
	ReminderLog       *reminderlogmodule.Controller
	Report            *reportmodule.Controller
	Audit             *auditmodule.Controller
	Workout           *workoutmodule.Controller
}

func NewControllers(cfg config.Config, db *pgxpool.Pool, redisClient *redis.Client) Controllers {
	authRepo := authmodule.NewRepository(db)
	authService := authmodule.NewService(cfg, authRepo, redisClient)
	subscriptionRepo := subscriptionmodule.NewRepository(db)
	subscriptionService := subscriptionmodule.NewService(cfg, subscriptionRepo)
	userRepo := usermodule.NewRepository(db)
	userService := usermodule.NewService(userRepo)
	checkinRepo := checkinmodule.NewRepository(db)
	checkinService := checkinmodule.NewService(cfg, checkinRepo)
	memberRepo := membermodule.NewRepository(db)
	memberService := membermodule.NewService(cfg, memberRepo)
	membershipPackageRepo := membershippackagemodule.NewRepository(db)
	membershipPackageService := membershippackagemodule.NewService(membershipPackageRepo)
	gymRepo := gymmodule.NewRepository(db)
	gymService := gymmodule.NewService(gymRepo)
	qrCodeRepo := qrcodemodule.NewRepository(db)
	qrCodeService := qrcodemodule.NewService(cfg, qrCodeRepo)
	expenseCategoryRepo := expensecategorymodule.NewRepository(db)
	expenseCategoryService := expensecategorymodule.NewService(expenseCategoryRepo)
	expenseRepo := expensemodule.NewRepository(db)
	expenseService := expensemodule.NewService(cfg, expenseRepo)
	paymentRepo := paymentmodule.NewRepository(db)
	paymentService := paymentmodule.NewService(cfg, paymentRepo)
	workoutRepo := workoutmodule.NewRepository(db)
	workoutService := workoutmodule.NewService(cfg, workoutRepo)
	reminderRuleRepo := reminderrulemodule.NewRepository(db)
	reminderRuleService := reminderrulemodule.NewService(reminderRuleRepo)
	reminderLogRepo := reminderlogmodule.NewRepository(db)
	reminderLogService := reminderlogmodule.NewService(reminderLogRepo)
	auditRepo := auditmodule.NewRepository(db)
	auditService := auditmodule.NewService(auditRepo)
	reportRepo := reportmodule.NewRepository(db)
	reportService := reportmodule.NewService(reportRepo)

	return Controllers{
		Auth:              authmodule.NewController(authService),
		Gym:               gymmodule.NewController(gymService),
		Subscription:      subscriptionmodule.NewController(subscriptionService),
		User:              usermodule.NewController(userService),
		Member:            membermodule.NewController(memberService),
		QRCode:            qrcodemodule.NewController(qrCodeService),
		MembershipPackage: membershippackagemodule.NewController(membershipPackageService),
		Payment:           paymentmodule.NewController(paymentService),
		Checkin:           checkinmodule.NewController(checkinService),
		ExpenseCategory:   expensecategorymodule.NewController(expenseCategoryService),
		Expense:           expensemodule.NewController(expenseService),
		ReminderRule:      reminderrulemodule.NewController(reminderRuleService),
		ReminderLog:       reminderlogmodule.NewController(reminderLogService),
		Report:            reportmodule.NewController(reportService),
		Audit:             auditmodule.NewController(auditService),
		Workout:           workoutmodule.NewController(workoutService),
	}
}
