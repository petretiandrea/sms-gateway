/*
 * SMS Gateway
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 1.0.0
 * Contact: petretiandrea@gmail.com
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

import (
	"github.com/gin-gonic/gin"
)

type SmsAPI interface {

	// GetSmsById Get /sms/:smsId
	// Get an sms
	GetSmsById(c *gin.Context)

	// SendSms Post /sms/
	// Send a new sms
	SendSms(c *gin.Context)
}
