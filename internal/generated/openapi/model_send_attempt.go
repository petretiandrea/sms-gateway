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

type SendAttempt struct {
	MessageId string `json:"messageId,omitempty"`

	PhoneId string `json:"phoneId,omitempty"`

	// Number of attempt
	Attempt int32 `json:"attempt,omitempty"`

	Result SendAttemptResult `json:"result,omitempty"`
}
