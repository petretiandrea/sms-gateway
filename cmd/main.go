package main

import (
	"context"
	firebase "firebase.google.com/go"
	"fmt"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	"net/http"
	"sms-gateway/internal/account"
	userApi "sms-gateway/internal/account/api"
	"sms-gateway/internal/phone"
	phoneApi "sms-gateway/internal/phone/api"
	"sms-gateway/internal/sms"
	"sms-gateway/internal/sms/api"
)

func main() {

	// create async firebase context
	firebaseContext := context.Background()
	credentials := option.WithCredentialsFile("be-aesthetic-admin-sdk.json")
	app, err := firebase.NewApp(firebaseContext, nil, credentials)
	if err != nil {
		fmt.Println("Failed to initialize firebase app")
		return
	}
	firestoreClient, err := app.Firestore(firebaseContext)
	if err != nil {
		fmt.Println("Failed to initialize firestore")
		return
	}
	defer firestoreClient.Close()
	// user account example
	accountRepository := account.NewFirestoreUserAccountRepository(firebaseContext, firestoreClient, "userAccounts")
	smsRepository := sms.NewMessageFirestoreRepository(firebaseContext, firestoreClient, "sms")
	phoneRepository := phone.NewFirestorePhoneRepository(firebaseContext, firestoreClient, "phones")

	userAccountController := userApi.UserAccountController{
		CreateUserAccountUseCase: account.NewUserAccountService(accountRepository),
	}

	smsController := api.SmsApiController{
		Account: account.NewUserAccountService(accountRepository),
		Sms:     sms.NewSmsService(&smsRepository),
	}

	phoneApiController := phoneApi.PhoneApiController{
		Phone:   phone.NewPhoneService(&phoneRepository),
		Account: account.NewUserAccountService(accountRepository),
	}

	server := gin.Default()
	server.GET("/ping", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	userAccountController.RegisterRoutes(server)
	smsController.RegisterRoutes(server)
	phoneApiController.RegisterRoutes(server)
	err = server.Run("0.0.0.0:8080")
	if err != nil {
		return
	}
}
