package reminder_rule

type CreateReminderRuleRequest struct {
	Name             string `json:"name" binding:"required" example:"Reminder H-3"`
	DaysBeforeExpiry int    `json:"days_before_expiry" binding:"min=0" example:"3"`
	Channel          string `json:"channel,omitempty" example:"whatsapp"`
	MessageTemplate  string `json:"message_template" binding:"required" example:"Halo {member_name}, membership kamu akan habis pada {end_date}."`
}

type UpdateReminderRuleRequest struct {
	Name             string `json:"name,omitempty" example:"Reminder H-7"`
	DaysBeforeExpiry int    `json:"days_before_expiry,omitempty" example:"7"`
	Channel          string `json:"channel,omitempty" example:"whatsapp"`
	MessageTemplate  string `json:"message_template,omitempty" example:"Halo {member_name}, membership kamu akan habis pada {end_date}."`
	IsActive         *bool  `json:"is_active,omitempty" example:"true"`
}
