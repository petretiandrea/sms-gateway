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

type CreateWebhookRequest struct {
	DefaultWebhookUrl string `json:"defaultWebhookUrl,omitempty"`

	Enabled bool `json:"enabled"`
}
