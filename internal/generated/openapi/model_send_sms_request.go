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

type SendSmsRequest struct {
	Content string `json:"content"`

	From string `json:"from"`

	To string `json:"to"`
}
