package models

import (
	"time"

	"github.com/google/uuid"
)

type GymStatus string
type UserRole string
type MemberStatus string
type SubscriptionStatus string
type PaymentMethod string
type PaymentStatus string
type ExpenseStatus string
type QRCodeStatus string
type CheckinSource string
type CheckinStatus string
type ReminderLogStatus string

const (
	GymStatusActive    GymStatus = "active"
	GymStatusInactive  GymStatus = "inactive"
	GymStatusSuspended GymStatus = "suspended"

	UserRoleOwner   UserRole = "owner"
	UserRoleAdmin   UserRole = "admin"
	UserRoleCashier UserRole = "cashier"
	UserRoleTrainer UserRole = "trainer"

	MemberStatusActive   MemberStatus = "active"
	MemberStatusInactive MemberStatus = "inactive"
	MemberStatusBanned   MemberStatus = "banned"

	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"

	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodTransfer PaymentMethod = "transfer"
	PaymentMethodQRIS     PaymentMethod = "qris"

	PaymentStatusPaid      PaymentStatus = "paid"
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCancelled PaymentStatus = "cancelled"
	PaymentStatusRefunded  PaymentStatus = "refunded"

	ExpenseStatusDraft    ExpenseStatus = "draft"
	ExpenseStatusApproved ExpenseStatus = "approved"
	ExpenseStatusRejected ExpenseStatus = "rejected"

	QRCodeStatusActive  QRCodeStatus = "active"
	QRCodeStatusRevoked QRCodeStatus = "revoked"

	CheckinSourceQR     CheckinSource = "qr"
	CheckinSourceManual CheckinSource = "manual"

	CheckinStatusValid          CheckinStatus = "valid"
	CheckinStatusInvalidQR      CheckinStatus = "invalid_qr"
	CheckinStatusExpired        CheckinStatus = "expired"
	CheckinStatusInactiveMember CheckinStatus = "inactive_member"
	CheckinStatusDuplicate      CheckinStatus = "duplicate"
	CheckinStatusCancelled      CheckinStatus = "cancelled"

	ReminderLogStatusPending ReminderLogStatus = "pending"
	ReminderLogStatusSent    ReminderLogStatus = "sent"
	ReminderLogStatusFailed  ReminderLogStatus = "failed"
)

type Gym struct {
	BaseModel
	Name, Phone, Email, Address string
	Timezone, Currency, Status  string
}

type User struct {
	BaseTenantModel
	Name, Email, PasswordHash, Role string
	IsActive                        bool
	LastLoginAt                     *time.Time
}

type Member struct {
	BaseTenantModel
	MemberCode, FullName, Phone, Email, Gender, Address string
	BirthDate                                           *time.Time
	EmergencyContactName, EmergencyContactPhone         string
	JoinedAt                                            time.Time
	Status, Notes                                       string
}

type MembershipPackage struct {
	BaseTenantModel
	Name, Description string
	DurationDays      int
	Price             int64
	IsActive          bool
}

type Subscription struct {
	BaseTenantModel
	MemberID, MembershipPackageID int64
	StartDate, EndDate            time.Time
	Status, Source                string
}

type Payment struct {
	BaseTenantModel
	MemberID, SubscriptionID                      int64
	InvoiceNo, PaymentType, PaymentMethod, Status string
	PackagePrice, DiscountAmount, FinalAmount     int64
	PaidAt                                        time.Time
	Notes                                         string
	CreatedBy                                     *int64
}

type ExpenseCategory struct {
	BaseTenantModel
	Name     string
	IsActive bool
}

type Expense struct {
	BaseTenantModel
	ExpenseCategoryID  int64
	Title, Description string
	Amount             int64
	ExpenseDate        time.Time
	Status             string
	CreatedBy          *int64
}

type MemberQRCode struct {
	BaseTenantModel
	MemberID        int64
	QRToken, Status string
	GeneratedAt     time.Time
	RevokedAt       *time.Time
}

type MemberCheckin struct {
	BaseTenantModel
	MemberID, SubscriptionID int64
	CheckinDate, CheckinAt   time.Time
	Source, Status, Notes    string
	ScannedBy                *int64
}

type ReminderRule struct {
	BaseTenantModel
	Name, Channel, MessageTemplate string
	DaysBeforeExpiry               int
	IsActive                       bool
}

type ReminderLog struct {
	BaseTenantModel
	MemberID, SubscriptionID, ReminderRuleID int64
	Channel, Recipient, Status               string
	SentAt                                   *time.Time
	ProviderMessageID, ErrorMessage          string
}

type AuditLog struct {
	ID         int64
	PublicID   uuid.UUID
	GymID      int64
	UserID     *int64
	Action     string
	EntityType string
	EntityID   *int64
	Payload    []byte
	CreatedAt  time.Time
}
