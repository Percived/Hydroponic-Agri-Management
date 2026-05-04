package alert

type UpdateAlertStatusRequest struct {
	Status  string `json:"status" binding:"required,oneof=ACK CLOSED"`
	Comment string `json:"comment" binding:"omitempty,max=255"`
}
