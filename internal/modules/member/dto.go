package member

type CreateMemberRequest struct {
	MemberCode            string `json:"member_code" binding:"required" example:"MBR-001"`
	FullName              string `json:"full_name" binding:"required" example:"Budi Santoso"`
	Phone                 string `json:"phone,omitempty" example:"08123456789"`
	Email                 string `json:"email,omitempty" example:"budi@email.com"`
	Gender                string `json:"gender,omitempty" example:"male"`
	BirthDate             string `json:"birth_date,omitempty" example:"1998-01-10"`
	Address               string `json:"address,omitempty" example:"Probolinggo"`
	EmergencyContactName  string `json:"emergency_contact_name,omitempty" example:"Siti"`
	EmergencyContactPhone string `json:"emergency_contact_phone,omitempty" example:"08999999999"`
	JoinedAt              string `json:"joined_at,omitempty" example:"2026-06-16"`
	Notes                 string `json:"notes,omitempty" example:"Member baru"`
}

type UpdateMemberRequest struct {
	FullName              string `json:"full_name,omitempty" example:"Budi Santoso Updated"`
	Phone                 string `json:"phone,omitempty" example:"08123456789"`
	Email                 string `json:"email,omitempty" example:"budi.updated@email.com"`
	Gender                string `json:"gender,omitempty" example:"male"`
	Address               string `json:"address,omitempty" example:"Probolinggo"`
	Status                string `json:"status,omitempty" example:"active"`
	Notes                 string `json:"notes,omitempty" example:"Updated notes"`
	EmergencyContactName  string `json:"emergency_contact_name,omitempty" example:"Siti"`
	EmergencyContactPhone string `json:"emergency_contact_phone,omitempty" example:"08999999999"`
}
