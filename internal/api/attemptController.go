package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sms-gateway/internal/application"
	"sms-gateway/internal/domain"
	"sms-gateway/internal/generated/openapi"
)

type AttemptController struct {
	SmsService application.SmsService
}

func (a AttemptController) ReportMessageStatus(c *gin.Context) {
	var attemptRequest openapi.SendAttempt
	user := c.MustGet("user").(domain.UserAccount)
	if err := c.BindJSON(&attemptRequest); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	// TODO: replace by sending it to a queue\
	if attempt := reportRequestToAttempt(attemptRequest); attempt == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"reason": "invalid attempt type",
		})
	} else {
		if sms, err := a.SmsService.RegisterAttempt(
			domain.SmsId(attemptRequest.MessageId),
			user.Id,
			attempt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, err)
		} else if sms == nil {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusAccepted)
		}
	}
}

func reportRequestToAttempt(attemptRequest openapi.SendAttempt) domain.Attempt {
	switch attemptRequest.Result.Type {
	case "success":
		return domain.SuccessAttempt{
			AttemptCount: attemptRequest.Attempt,
			PhoneId:      domain.PhoneId(attemptRequest.PhoneId),
		}
	case "failure":
		return domain.FailedAttempt{
			AttemptCount: attemptRequest.Attempt,
			Reason:       attemptRequest.Result.Reason,
			PhoneId:      domain.PhoneId(attemptRequest.PhoneId),
		}
	}
	return nil
}

var (
	_ openapi.ReportsAPI = (*AttemptController)(nil)
)
