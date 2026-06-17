package subscription

type CreateGymSubscriptionRequest struct {
	SaasPlanPublicID string `json:"saas_plan_public_id" binding:"required" example:"00000000-0000-4000-8000-000000000102"`
	StartDate        string `json:"start_date,omitempty" example:"2026-06-16"`
	PaymentMethod    string `json:"payment_method,omitempty" example:"transfer"`
	Status           string `json:"status,omitempty" example:"paid"`
	AutoRenew        *bool  `json:"auto_renew,omitempty" example:"false"`
	Notes            string `json:"notes,omitempty" example:"Owner subscribe paket Premium"`
}

type CreateSubscriptionRequest struct {
	MemberPublicID            string `json:"member_public_id" binding:"required" example:"member-public-id"`
	MembershipPackagePublicID string `json:"membership_package_public_id" binding:"required" example:"package-public-id"`
	StartDate                 string `json:"start_date,omitempty" example:"2026-06-16"`
}
