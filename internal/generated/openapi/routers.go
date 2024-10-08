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
	"net/http"

	"github.com/gin-gonic/gin"
)

// Route is the information for every URI.
type Route struct {
	// Name is the name of this Route.
	Name string
	// Method is the string for the HTTP method. ex) GET, POST etc..
	Method string
	// Pattern is the pattern of the URI.
	Pattern string
	// HandlerFunc is the handler function of this route.
	HandlerFunc gin.HandlerFunc
}

// NewRouter returns a new router.
func NewRouter(handleFunctions ApiHandleFunctions) *gin.Engine {
	return NewRouterWithGinEngine(gin.Default(), handleFunctions)
}

// NewRouter add routes to existing gin engine.
func NewRouterWithGinEngine(router *gin.Engine, handleFunctions ApiHandleFunctions) *gin.Engine {
	for _, route := range getRoutes(handleFunctions) {
		if route.HandlerFunc == nil {
			route.HandlerFunc = DefaultHandleFunc
		}
		switch route.Method {
		case http.MethodGet:
			router.GET(route.Pattern, route.HandlerFunc)
		case http.MethodPost:
			router.POST(route.Pattern, route.HandlerFunc)
		case http.MethodPut:
			router.PUT(route.Pattern, route.HandlerFunc)
		case http.MethodPatch:
			router.PATCH(route.Pattern, route.HandlerFunc)
		case http.MethodDelete:
			router.DELETE(route.Pattern, route.HandlerFunc)
		}
	}

	return router
}

// Default handler for not yet implemented routes
func DefaultHandleFunc(c *gin.Context) {
	c.String(http.StatusNotImplemented, "501 not implemented")
}

type ApiHandleFunctions struct {

	// Routes for the AccountAPI part of the API
	AccountAPI AccountAPI
	// Routes for the PhoneAPI part of the API
	PhoneAPI PhoneAPI
	// Routes for the ReportsAPI part of the API
	ReportsAPI ReportsAPI
	// Routes for the SmsAPI part of the API
	SmsAPI SmsAPI
	// Routes for the WebhooksAPI part of the API
	WebhooksAPI WebhooksAPI
}

func getRoutes(handleFunctions ApiHandleFunctions) []Route {
	return []Route{
		{
			"GetAccountById",
			http.MethodGet,
			"/accounts/:accountId",
			handleFunctions.AccountAPI.GetAccountById,
		},
		{
			"RegisterAccount",
			http.MethodPost,
			"/accounts",
			handleFunctions.AccountAPI.RegisterAccount,
		},
		{
			"GetPhoneById",
			http.MethodGet,
			"/phones/:phoneId",
			handleFunctions.PhoneAPI.GetPhoneById,
		},
		{
			"RegisterPhone",
			http.MethodPost,
			"/phones",
			handleFunctions.PhoneAPI.RegisterPhone,
		},
		{
			"UpdateFcmToken",
			http.MethodPut,
			"/phones/:phoneId",
			handleFunctions.PhoneAPI.UpdateFcmToken,
		},
		{
			"ReportMessageStatus",
			http.MethodPost,
			"/attempts",
			handleFunctions.ReportsAPI.ReportMessageStatus,
		},
		{
			"GetMessages",
			http.MethodGet,
			"/messages",
			handleFunctions.SmsAPI.GetMessages,
		},
		{
			"GetSmsById",
			http.MethodGet,
			"/messages/:smsId",
			handleFunctions.SmsAPI.GetSmsById,
		},
		{
			"SendSms",
			http.MethodPost,
			"/messages",
			handleFunctions.SmsAPI.SendSms,
		},
		{
			"WebhooksPost",
			http.MethodPost,
			"/webhooks",
			handleFunctions.WebhooksAPI.WebhooksPost,
		},
	}
}
