package gym

type UpdateGymRequest struct {
	Name     string `json:"name,omitempty" example:"Gym Sehat"`
	Phone    string `json:"phone,omitempty" example:"08123456789"`
	Email    string `json:"email,omitempty" example:"admin@gym.com"`
	Address  string `json:"address,omitempty" example:"Jl. Contoh No. 1"`
	Timezone string `json:"timezone,omitempty" example:"Asia/Jakarta"`
	Currency string `json:"currency,omitempty" example:"IDR"`
}
