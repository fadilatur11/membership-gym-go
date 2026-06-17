package routes

import "github.com/gin-gonic/gin"

func RegisterAdminRoutes(r *gin.RouterGroup, deps Dependencies) {
	auth := r.Group("/auth")
	auth.POST("/login", deps.Auth.Login)
	auth.POST("/refresh", deps.Auth.Refresh)

	protected := r.Group("")
	protected.Use(deps.AuthMiddleware.RequireAuth())

	protected.POST("/auth/logout", deps.Auth.Logout)
	protected.GET("/auth/profile", deps.Auth.Profile)

	protected.GET("/gym/profile", deps.Gym.GymProfile)
	protected.PATCH("/gym/profile", deps.RoleMiddleware.Require("owner", "admin"), deps.Gym.UpdateGym)

	protected.GET("/saas-plans", deps.Subscription.ListSaasPlans)
	protected.GET("/gym/subscription", deps.Subscription.CurrentGymSubscription)
	protected.POST("/gym/subscription", deps.RoleMiddleware.Require("owner"), deps.Subscription.CreateGymSubscription)

	protected.GET("/users", deps.User.RequirePlanFeature("staff.users"), deps.User.ListUsers)
	protected.POST("/users", deps.User.RequirePlanFeature("staff.users"), deps.RoleMiddleware.Require("owner", "admin"), deps.User.CreateUser)
	protected.GET("/users/:user_public_id", deps.User.RequirePlanFeature("staff.users"), deps.User.DetailUser)
	protected.PATCH("/users/:user_public_id", deps.User.RequirePlanFeature("staff.users"), deps.RoleMiddleware.Require("owner", "admin"), deps.User.UpdateUserDTO)
	protected.DELETE("/users/:user_public_id", deps.User.RequirePlanFeature("staff.users"), deps.RoleMiddleware.Require("owner", "admin"), deps.User.DeleteUser)

	protected.GET("/members", deps.Member.RequirePlanFeature("membership.members"), deps.Member.ListMembers)
	protected.POST("/members", deps.Member.RequirePlanFeature("membership.members"), deps.RoleMiddleware.Require("owner", "admin", "cashier"), deps.Member.CreateMember)
	protected.GET("/members/:member_public_id", deps.Member.RequirePlanFeature("membership.members"), deps.Member.DetailMember)
	protected.PATCH("/members/:member_public_id", deps.Member.RequirePlanFeature("membership.members"), deps.RoleMiddleware.Require("owner", "admin", "cashier"), deps.Member.UpdateMemberDTO)
	protected.DELETE("/members/:member_public_id", deps.Member.RequirePlanFeature("membership.members"), deps.RoleMiddleware.Require("owner", "admin"), deps.Member.DeleteMember)

	protected.GET("/members/:member_public_id/qrcode", deps.QRCode.RequirePlanFeature("member.qrcode"), deps.QRCode.GetQR)
	protected.POST("/members/:member_public_id/qrcode/generate", deps.QRCode.RequirePlanFeature("member.qrcode"), deps.QRCode.GenerateQR)
	protected.POST("/members/:member_public_id/qrcode/regenerate", deps.QRCode.RequirePlanFeature("member.qrcode"), deps.QRCode.RegenerateQR)
	protected.POST("/members/:member_public_id/qrcode/revoke", deps.QRCode.RequirePlanFeature("member.qrcode"), deps.QRCode.RevokeQR)
	protected.GET("/members/:member_public_id/checkins", deps.Checkin.RequirePlanFeature("member.checkin"), deps.Checkin.MemberCheckins)

	protected.GET("/membership-packages", deps.MembershipPackage.RequirePlanFeature("membership.packages"), deps.MembershipPackage.ListPackages)
	protected.POST("/membership-packages", deps.MembershipPackage.RequirePlanFeature("membership.packages"), deps.RoleMiddleware.Require("owner", "admin"), deps.MembershipPackage.CreatePackage)
	protected.GET("/membership-packages/:package_public_id", deps.MembershipPackage.RequirePlanFeature("membership.packages"), deps.MembershipPackage.DetailPackage)
	protected.PATCH("/membership-packages/:package_public_id", deps.MembershipPackage.RequirePlanFeature("membership.packages"), deps.RoleMiddleware.Require("owner", "admin"), deps.MembershipPackage.UpdatePackageDTO)
	protected.DELETE("/membership-packages/:package_public_id", deps.MembershipPackage.RequirePlanFeature("membership.packages"), deps.RoleMiddleware.Require("owner", "admin"), deps.MembershipPackage.DeletePackage)

	protected.GET("/subscriptions", deps.Subscription.RequirePlanFeature("membership.subscriptions"), deps.Subscription.ListSubscriptions)
	protected.POST("/subscriptions", deps.Subscription.RequirePlanFeature("membership.subscriptions"), deps.RoleMiddleware.Require("owner", "admin", "cashier"), deps.Subscription.CreateSubscription)
	protected.GET("/subscriptions/:subscription_public_id", deps.Subscription.RequirePlanFeature("membership.subscriptions"), deps.Subscription.DetailSubscription)
	protected.PATCH("/subscriptions/:subscription_public_id/cancel", deps.Subscription.RequirePlanFeature("membership.subscriptions"), deps.RoleMiddleware.Require("owner", "admin"), deps.Subscription.CancelSubscription)
	protected.PATCH("/subscriptions/:subscription_public_id/expire", deps.Subscription.RequirePlanFeature("membership.subscriptions"), deps.RoleMiddleware.Require("owner", "admin"), deps.Subscription.ExpireSubscription)

	protected.GET("/payments", deps.Payment.RequirePlanFeature("membership.payments"), deps.Payment.ListPayments)
	protected.POST("/payments", deps.Payment.RequirePlanFeature("membership.payments"), deps.RoleMiddleware.Require("owner", "admin", "cashier"), deps.Payment.CreatePayment)
	protected.GET("/payments/:payment_public_id", deps.Payment.RequirePlanFeature("membership.payments"), deps.Payment.DetailPayment)
	protected.PATCH("/payments/:payment_public_id/cancel", deps.Payment.RequirePlanFeature("membership.payments"), deps.RoleMiddleware.Require("owner", "admin"), deps.Payment.CancelPayment)
	protected.PATCH("/payments/:payment_public_id/refund", deps.Payment.RequirePlanFeature("membership.payments"), deps.RoleMiddleware.Require("owner", "admin"), deps.Payment.RefundPayment)

	protected.POST("/checkins/scan", deps.Checkin.RequirePlanFeature("member.checkin"), deps.Checkin.ScanCheckin)
	protected.POST("/checkins/manual", deps.Checkin.RequirePlanFeature("member.checkin"), deps.Checkin.ManualCheckin)
	protected.GET("/checkins", deps.Checkin.RequirePlanFeature("member.checkin"), deps.Checkin.ListCheckins)
	protected.GET("/checkins/today", deps.Checkin.RequirePlanFeature("member.checkin"), deps.Checkin.TodayCheckins)

	protected.GET("/expense-categories", deps.ExpenseCategory.RequirePlanFeature("expenses"), deps.ExpenseCategory.ListExpenseCategories)
	protected.POST("/expense-categories", deps.ExpenseCategory.RequirePlanFeature("expenses"), deps.RoleMiddleware.Require("owner", "admin", "cashier"), deps.ExpenseCategory.CreateExpenseCategory)
	protected.PATCH("/expense-categories/:category_public_id", deps.ExpenseCategory.RequirePlanFeature("expenses"), deps.RoleMiddleware.Require("owner", "admin"), deps.ExpenseCategory.UpdateExpenseCategoryDTO)
	protected.DELETE("/expense-categories/:category_public_id", deps.ExpenseCategory.RequirePlanFeature("expenses"), deps.RoleMiddleware.Require("owner", "admin"), deps.ExpenseCategory.DeleteExpenseCategory)

	protected.GET("/expenses", deps.Expense.RequirePlanFeature("expenses"), deps.Expense.ListExpenses)
	protected.POST("/expenses", deps.Expense.RequirePlanFeature("expenses"), deps.RoleMiddleware.Require("owner", "admin", "cashier"), deps.Expense.CreateExpense)
	protected.GET("/expenses/:expense_public_id", deps.Expense.RequirePlanFeature("expenses"), deps.Expense.DetailExpense)
	protected.PATCH("/expenses/:expense_public_id", deps.Expense.RequirePlanFeature("expenses"), deps.RoleMiddleware.Require("owner", "admin"), deps.Expense.UpdateExpenseDTO)
	protected.PATCH("/expenses/:expense_public_id/approve", deps.Expense.RequirePlanFeature("expenses"), deps.RoleMiddleware.Require("owner", "admin"), deps.Expense.ApproveExpense)
	protected.PATCH("/expenses/:expense_public_id/reject", deps.Expense.RequirePlanFeature("expenses"), deps.RoleMiddleware.Require("owner", "admin"), deps.Expense.RejectExpense)

	protected.GET("/reminder-rules", deps.ReminderRule.RequirePlanFeature("reminders"), deps.ReminderRule.ListReminderRules)
	protected.POST("/reminder-rules", deps.ReminderRule.RequirePlanFeature("reminders"), deps.RoleMiddleware.Require("owner", "admin"), deps.ReminderRule.CreateReminderRule)
	protected.GET("/reminder-rules/:rule_public_id", deps.ReminderRule.RequirePlanFeature("reminders"), deps.ReminderRule.DetailReminderRule)
	protected.PATCH("/reminder-rules/:rule_public_id", deps.ReminderRule.RequirePlanFeature("reminders"), deps.RoleMiddleware.Require("owner", "admin"), deps.ReminderRule.UpdateReminderRuleDTO)
	protected.DELETE("/reminder-rules/:rule_public_id", deps.ReminderRule.RequirePlanFeature("reminders"), deps.RoleMiddleware.Require("owner", "admin"), deps.ReminderRule.DeleteReminderRule)
	protected.GET("/reminder-logs", deps.ReminderLog.RequirePlanFeature("reminders"), deps.ReminderLog.ListReminderLogs)
	protected.POST("/reminder-logs/:log_public_id/retry", deps.ReminderLog.RequirePlanFeature("reminders"), deps.ReminderLog.RetryReminderLog)

	protected.GET("/reports/dashboard", deps.Report.RequirePlanFeature("reports.dashboard"), deps.Report.Dashboard)
	protected.GET("/reports/payments", deps.Report.RequirePlanFeature("reports.financial"), deps.Report.ReportPayments)
	protected.GET("/reports/expenses", deps.Report.RequirePlanFeature("reports.financial"), deps.Report.ReportExpenses)
	protected.GET("/reports/profit-loss", deps.Report.RequirePlanFeature("reports.financial"), deps.Report.ReportProfitLoss)
	protected.GET("/reports/checkins", deps.Report.RequirePlanFeature("reports.dashboard"), deps.Report.ReportCheckins)
	protected.GET("/reports/expired-members", deps.Report.RequirePlanFeature("reports.dashboard"), deps.Report.ReportExpiredMembers)
	protected.GET("/audit-logs", deps.Audit.RequirePlanFeature("audit_logs"), deps.Audit.AuditLogs)
}
