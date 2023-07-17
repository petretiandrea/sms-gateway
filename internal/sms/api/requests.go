package api

type SendSmsRequest struct {
	Content string `json:"content" binding:"required"`
	From    string `json:"from" binding:"required"`
	To      string `json:"to" binding:"required"`
}
