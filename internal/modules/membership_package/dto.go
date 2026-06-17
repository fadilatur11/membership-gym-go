package membership_package

type CreateMembershipPackageRequest struct {
	Name         string `json:"name" binding:"required" example:"Bulanan"`
	DurationDays int    `json:"duration_days" binding:"required,min=1" example:"30"`
	Price        int64  `json:"price" binding:"min=0" example:"300000"`
	Description  string `json:"description,omitempty" example:"Membership 30 hari"`
}

type UpdateMembershipPackageRequest struct {
	Name         string `json:"name,omitempty" example:"Bulanan Promo"`
	DurationDays int    `json:"duration_days,omitempty" example:"30"`
	Price        int64  `json:"price,omitempty" example:"250000"`
	Description  string `json:"description,omitempty" example:"Promo bulan ini"`
	IsActive     *bool  `json:"is_active,omitempty" example:"true"`
}
