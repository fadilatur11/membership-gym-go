package payment

type CreatePaymentRequest struct {
	MemberPublicID            string `json:"member_public_id" binding:"required" example:"member-public-id"`
	MembershipPackagePublicID string `json:"membership_package_public_id" binding:"required" example:"package-public-id"`
	StartDate                 string `json:"start_date,omitempty" example:"2026-06-16"`
	PaymentMethod             string `json:"payment_method" binding:"required,oneof=cash transfer qris" example:"cash"`
	DiscountAmount            int64  `json:"discount_amount,omitempty" example:"50000"`
	Status                    string `json:"status,omitempty" example:"paid"`
	Notes                     string `json:"notes,omitempty" example:"Promo member lama"`
}

type PaymentActionRequest struct {
	Notes string `json:"notes,omitempty" example:"Salah input pembayaran"`
}
