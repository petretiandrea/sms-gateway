package api

type RegisterPhoneRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type UpdateFCMRequest struct {
	Token string `json:"phone" binding:"required"`
}
