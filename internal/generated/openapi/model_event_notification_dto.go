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

type EventNotificationDto struct {
	EventType EventNotificationType `json:"eventType,omitempty"`

	Data SmsEntityResponse `json:"data,omitempty"`

	Metadata map[string]string `json:"metadata,omitempty"`
}
