package expense_category

type CreateExpenseCategoryRequest struct {
	Name string `json:"name" binding:"required" example:"Listrik"`
}

type UpdateExpenseCategoryRequest struct {
	Name     string `json:"name,omitempty" example:"Operasional"`
	IsActive *bool  `json:"is_active,omitempty" example:"true"`
}
