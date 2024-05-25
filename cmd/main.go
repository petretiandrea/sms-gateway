package main

import (
	"context"
	firebase "firebase.google.com/go"
	"fmt"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	userApi "sms-gateway/internal/api"
	"sms-gateway/internal/application"
	"sms-gateway/internal/config"
	"sms-gateway/internal/events"
	"sms-gateway/internal/health"
	"sms-gateway/internal/infra"
	"sms-gateway/internal/infra/repos"
	"strconv"
	"time"
)

const VERSION = "1.0.1"

func main() {
	PrintInfo()
	log, _ := zap.NewProduction()
	zap.ReplaceGlobals(log)
	appConfig := config.LoadConfig("app.yaml")
	cleanupTracer, errTracer := initTracer(OpenTelemetryConfig{
		serviceName:    appConfig.ServiceName,
		serviceVersion: VERSION,
		ctx:            context.Background(),
	})
	if errTracer != nil {
		log.Error("Failed to initialize OpenTelemetry!")
	}
	defer cleanupTracer()

	server := gin.New()
	server.Use(ginzap.GinzapWithConfig(log, &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		SkipPaths:  []string{"/health"},
	}))
	server.Use(ginzap.RecoveryWithZap(log, true))
	server.Use(otelgin.Middleware(
		appConfig.ServiceName,
		otelgin.WithFilter(health.FilterHealthCheck),
	))

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

	// initialize mongo
	mongoContext := context.Background()
	mongoClient, err := connectMongo(appConfig.MongoConnectionString)
	if err != nil {
		log.Error("Failed to initialize mongodb")
		return
	}

	// user account example
	accountRepository := repos.NewFirestoreUserAccountRepository(ctx, firestoreClient, appConfig.FirebaseConfig.UserAccount)
	smsRepository := repos.NewMessageFirestoreRepository(ctx, firestoreClient, appConfig.FirebaseConfig.Sms)
	phoneRepository := repos.NewFirestorePhoneRepository(ctx, firestoreClient, appConfig.FirebaseConfig.Phone)
	deliveryNotificationRepo := repos.NewMongoDeliveryNotificationRepository(mongoContext, mongoClient, appConfig.MongoDatabaseName)

	webHookNotifier := userApi.HttpWebhookNotifier{}
	deliveryNotificationService := application.NewDeliveryNotificationService(deliveryNotificationRepo, smsRepository, webHookNotifier)
	userAccountService := application.NewUserAccountService(accountRepository)

	userAccountController := userApi.UserAccountController{
		CreateUserAccountUseCase: application.NewUserAccountService(accountRepository),
	}

	smsController := userApi.SmsApiController{
		Account: userAccountService,
		Sms:     application.NewSmsService(&smsRepository, application.NewPhoneService(&phoneRepository), pushService),
	}

	phoneApiController := userApi.PhoneApiController{
		Phone:   application.NewPhoneService(&phoneRepository),
		Account: userAccountService,
	}

	deliveryNotificationController := userApi.DeliveryNotificationController{
		Account:              userAccountService,
		DeliveryNotification: deliveryNotificationService,
	}

	listener := infra.NewFirestoreEventListener(ctx, firestoreClient, appConfig.FirebaseConfig.Sms)

	deliveryConsumer := events.NewDeliveryNotificationConsumer(listener, deliveryNotificationService)
	go deliveryConsumer.Start()
	defer deliveryConsumer.Stop()

	health.RegisterGinHealthCheck(server, firestoreClient)
	userAccountController.RegisterRoutes(server)
	smsController.RegisterRoutes(server)
	phoneApiController.RegisterRoutes(server)
	deliveryNotificationController.RegisterRoutes(server)

	err = server.Run("0.0.0.0:8080")
	if err != nil {
		return
	}
}

func PrintInfo() {
	fmt.Printf("\n   _____                  _____       _                           \n  / ____|                / ____|     | |                          \n | (___  _ __ ___  ___  | |  __  __ _| |_ _____      ____ _ _   _ \n  \\___ \\| '_ ` _ \\/ __| | | |_ |/ _` | __/ _ \\ \\ /\\ / / _` | | | |\n  ____) | | | | | \\__ \\ | |__| | (_| | ||  __/\\ V  V / (_| | |_| |\n |_____/|_| |_| |_|___/  \\_____|\\__,_|\\__\\___| \\_/\\_/ \\__,_|\\__, |\n                                                             __/ |\n                                                            |___/ \n")
	fmt.Printf("Version %s\n", VERSION)
}
