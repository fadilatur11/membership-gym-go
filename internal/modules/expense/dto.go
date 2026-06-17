package expense

type CreateExpenseRequest struct {
	ExpenseCategoryPublicID string `json:"expense_category_public_id" binding:"required" example:"category-public-id"`
	Title                   string `json:"title" binding:"required" example:"Bayar listrik Juni"`
	Description             string `json:"description,omitempty" example:"Token listrik"`
	Amount                  int64  `json:"amount" binding:"required,min=0" example:"500000"`
	ExpenseDate             string `json:"expense_date,omitempty" example:"2026-06-16"`
	Status                  string `json:"status,omitempty" example:"approved"`
}

type UpdateExpenseRequest struct {
	Title       string `json:"title,omitempty" example:"Bayar listrik Juni Updated"`
	Description string `json:"description,omitempty" example:"Token listrik"`
	Amount      int64  `json:"amount,omitempty" example:"550000"`
	ExpenseDate string `json:"expense_date,omitempty" example:"2026-06-16"`
	Status      string `json:"status,omitempty" example:"approved"`
}
