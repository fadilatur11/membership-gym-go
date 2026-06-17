package auth

type LoginRequest struct {
	GymPublicID string `json:"gym_public_id,omitempty" example:"00000000-0000-4000-8000-000000000001"`
	Email       string `json:"email" binding:"required,email" example:"owner@gym.com"`
	Password    string `json:"password" binding:"required" example:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"refresh-token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"refresh-token"`
}

type RegisterOwnerRequest struct {
	GymName       string `json:"gym_name" binding:"required" example:"Strong Gym"`
	GymPhone      string `json:"gym_phone,omitempty" example:"08123456789"`
	GymEmail      string `json:"gym_email,omitempty" example:"admin@stronggym.com"`
	GymAddress    string `json:"gym_address,omitempty" example:"Jl. Contoh No. 1"`
	OwnerName     string `json:"owner_name" binding:"required" example:"Budi Owner"`
	OwnerEmail    string `json:"owner_email" binding:"required,email" example:"owner@stronggym.com"`
	OwnerPassword string `json:"owner_password" binding:"required,min=6" example:"secret123"`
}
