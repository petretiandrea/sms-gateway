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
	"net/http"
	userApi "sms-gateway/internal/api"
	"sms-gateway/internal/application"
	"sms-gateway/internal/config"
	"sms-gateway/internal/events"
	"sms-gateway/internal/generated/openapi"
	"sms-gateway/internal/health"
	"sms-gateway/internal/infra"
	"sms-gateway/internal/infra/changes"
	"sms-gateway/internal/infra/repos/mongo"
	"strconv"
	"strings"
	"time"
)

const VERSION = "2.0.0"

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
	firebaseMessaging, err := app.Messaging(ctx)
	if err != nil {
		log.Error("Failed to initialize firebase messaging")
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
	mongoDatabase := mongoClient.Database(appConfig.MongoDatabaseName)

	// user account example
	accountRepository := mongo.NewMongoUserAccountRepository(ctx, mongoDatabase.Collection("accounts"))
	messageRepository := mongo.NewMongoMessageRepository(ctx, mongoDatabase.Collection("messages"))
	phoneRepository := mongo.NewMongoPhoneRepository(ctx, mongoDatabase.Collection("phones"))
	deliveryNotificationRepo := mongo.NewMongoDeliveryNotificationRepository(mongoContext, mongoDatabase.Collection("deliveryconfigs"))

	changeFeedProducer := changes.NewMessageChangeFeedProducer()
	webHookNotifier := userApi.HttpWebhookNotifier{}
	deliveryNotificationService := application.NewDeliveryNotificationService(deliveryNotificationRepo, messageRepository, webHookNotifier)
	userAccountService := application.NewUserAccountService(accountRepository)
	smsService := application.NewSmsService(&messageRepository, application.NewPhoneService(&phoneRepository), pushService, changeFeedProducer)

	deliveryConsumer := events.NewDeliveryNotificationConsumer(changeFeedProducer, deliveryNotificationService)
	go deliveryConsumer.Start()
	defer deliveryConsumer.Stop()

	server.Use(userApi.NewApiKeyMiddleware(userAccountService, func(request *http.Request) bool {
		return strings.Contains(request.URL.Path, "/phones") ||
			strings.Contains(request.URL.Path, "/messages") ||
			strings.Contains(request.URL.Path, "/webhook") ||
			strings.Contains(request.URL.Path, "/attempts")
	}))

	health.RegisterGinHealthCheck(server, mongoClient)

	openapi.NewRouterWithGinEngine(server, openapi.ApiHandleFunctions{
		AccountAPI: userApi.UserAccountController{
			CreateUserAccountUseCase: application.NewUserAccountService(accountRepository),
		},
		PhoneAPI: userApi.PhoneApiController{
			Phone:   application.NewPhoneService(&phoneRepository),
			Account: userAccountService,
		},
		SmsAPI: userApi.SmsApiController{
			Account: userAccountService,
			Sms:     smsService,
		},
		WebhooksAPI: userApi.DeliveryNotificationController{
			Account:              userAccountService,
			DeliveryNotification: deliveryNotificationService,
		},
		ReportsAPI: userApi.AttemptController{
			SmsService: smsService,
		},
	})

	err = server.Run("0.0.0.0:8080")
	if err != nil {
		return
	}
}

func PrintInfo() {
	fmt.Printf("\n   _____                  _____       _                           \n  / ____|                / ____|     | |                          \n | (___  _ __ ___  ___  | |  __  __ _| |_ _____      ____ _ _   _ \n  \\___ \\| '_ ` _ \\/ __| | | |_ |/ _` | __/ _ \\ \\ /\\ / / _` | | | |\n  ____) | | | | | \\__ \\ | |__| | (_| | ||  __/\\ V  V / (_| | |_| |\n |_____/|_| |_| |_|___/  \\_____|\\__,_|\\__\\___| \\_/\\_/ \\__,_|\\__, |\n                                                             __/ |\n                                                            |___/ \n")
	fmt.Printf("Version %s\n", VERSION)
}
