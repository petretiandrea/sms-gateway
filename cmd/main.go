package main

import (
	"context"
	firebase "firebase.google.com/go"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	userApi "sms-gateway/internal/api"
	"sms-gateway/internal/application"
	"sms-gateway/internal/config"
	"sms-gateway/internal/health"
	"sms-gateway/internal/infra"
	"sms-gateway/internal/infra/repos"
	"strconv"
	"time"
)

func main() {
	appConfig := config.LoadConfig("app.yaml")
	log, _ := zap.NewProduction()
	server := gin.New()
	server.Use(ginzap.Ginzap(log, time.RFC3339, true))
	server.Use(ginzap.RecoveryWithZap(log, true))

	// create async firebase ctx
	ctx := context.Background()
	credentials := option.WithCredentialsFile(appConfig.FirebaseConfig.CredentialsFile)
	app, err := firebase.NewApp(ctx, nil, credentials)
	if err != nil {
		log.Error("Failed to initialize firebase app")
		return
	}
	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		log.Error("Failed to initialize firestore")
		return
	}
	defer firestoreClient.Close()
	firebaseMessaging, err := app.Messaging(ctx)
	if err != nil {
		log.Error("Failed to initialize firestore")
		return
	}
	pushService := infra.NewFirebasePushNotification(ctx, firebaseMessaging)
	if active, _ := strconv.ParseBool(appConfig.DryRun); active {
		pushService.EnableDryRun()
	}

	// user account example
	accountRepository := repos.NewFirestoreUserAccountRepository(ctx, firestoreClient, appConfig.FirebaseConfig.UserAccount)
	smsRepository := repos.NewMessageFirestoreRepository(ctx, firestoreClient, appConfig.FirebaseConfig.Sms)
	phoneRepository := repos.NewFirestorePhoneRepository(ctx, firestoreClient, appConfig.FirebaseConfig.Phone)

	userAccountController := userApi.UserAccountController{
		CreateUserAccountUseCase: application.NewUserAccountService(accountRepository),
	}

	smsController := userApi.SmsApiController{
		Account: application.NewUserAccountService(accountRepository),
		Sms:     application.NewSmsService(&smsRepository, application.NewPhoneService(&phoneRepository), pushService),
	}

	phoneApiController := userApi.PhoneApiController{
		Phone:   application.NewPhoneService(&phoneRepository),
		Account: application.NewUserAccountService(accountRepository),
	}

	listener := infra.NewFirestoreEventListener(ctx, firestoreClient, appConfig.FirebaseConfig.Sms)
	go listener.ListenChanges()
	defer listener.StopListenChanges()

	health.RegisterGinHealthCheck(server, firestoreClient)
	userAccountController.RegisterRoutes(server)
	smsController.RegisterRoutes(server)
	phoneApiController.RegisterRoutes(server)
	err = server.Run("0.0.0.0:8080")
	if err != nil {
		return
	}
}
