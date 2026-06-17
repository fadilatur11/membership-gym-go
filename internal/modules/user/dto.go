package user

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required" example:"Andi Trainer"`
	Email    string `json:"email" binding:"required,email" example:"andi@gym.com"`
	Password string `json:"password" binding:"required,min=6" example:"secret123"`
	Role     string `json:"role" binding:"required,oneof=owner admin cashier trainer" example:"trainer"`
}

type UpdateUserRequest struct {
	Name     string `json:"name,omitempty" example:"Andi Updated"`
	Email    string `json:"email,omitempty" example:"andi.updated@gym.com"`
	Role     string `json:"role,omitempty" example:"cashier"`
	IsActive *bool  `json:"is_active,omitempty" example:"true"`
}
