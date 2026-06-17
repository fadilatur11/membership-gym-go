package checkin

type ScanCheckinRequest struct {
	QRToken string `json:"qr_token" binding:"required" example:"qr-token-random"`
}

type ManualCheckinRequest struct {
	MemberPublicID string `json:"member_public_id" binding:"required" example:"member-public-id"`
	Notes          string `json:"notes,omitempty" example:"Manual check-in karena kamera bermasalah"`
}
