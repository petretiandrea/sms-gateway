package main

import (
	"context"
	_ "embed"
	"fmt"
	"sms-gateway/internal/api"
	"sms-gateway/internal/api/middleware"
	"sms-gateway/internal/application"
	"sms-gateway/internal/config"
	"sms-gateway/internal/events"
	"sms-gateway/internal/health"
	"sms-gateway/internal/infra"
	"sms-gateway/internal/infra/changes"
	"sms-gateway/internal/infra/repos/mongo"
	"strconv"
	"strings"
	"time"

	firebase "firebase.google.com/go"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

//go:embed version.txt
var version string

func main() {
	PrintInfo()
	log, _ := zap.NewProduction()
	zap.ReplaceGlobals(log)
	appConfig := config.LoadConfig("../app-dev.yaml")
	cleanupTracer, errTracer := initTracer(OpenTelemetryConfig{
		serviceName:    appConfig.ServiceName,
		serviceVersion: version,
		ctx:            context.Background(),
	})
	if errTracer != nil {
		log.Error("Failed to initialize OpenTelemetry!")
	}
	defer cleanupTracer()

	server := gin.New()
	server.Use(func(c *gin.Context) {
		if strings.HasSuffix(c.Request.URL.Path, "/") {
			c.Request.URL.Path = strings.TrimRight(c.Request.URL.Path, "/")
			server.HandleContext(c)
		} else {
			c.Next()
		}
	})
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
	webHookNotifier := api.HttpWebhookNotifier{}
	deliveryNotificationService := application.NewDeliveryNotificationService(deliveryNotificationRepo, messageRepository, webHookNotifier)
	userAccountService := application.NewUserAccountService(accountRepository)
	smsService := application.NewSmsService(&messageRepository, application.NewPhoneService(&phoneRepository), pushService, changeFeedProducer)

	deliveryConsumer := events.NewDeliveryNotificationConsumer(changeFeedProducer, deliveryNotificationService)
	go deliveryConsumer.Start()
	defer deliveryConsumer.Stop()

	apiKeyMiddleware := middleware.NewApiKeyMiddleware(userAccountService)

	health.RegisterGinHealthCheck(server, mongoClient)

	serverInterface := api.Controller {
		AttemptController: api.AttemptController{
			SmsService: smsService,
		},
		PhoneApiController: api.PhoneApiController {
			Phone: application.NewPhoneService(&phoneRepository),
			Account: userAccountService,
		},
		SmsApiController: api.SmsApiController {
			Account: userAccountService,
			Sms: smsService,
			MessageRepository: messageRepository,
		},
		UserAccountController: api.UserAccountController{
			CreateUserAccountUseCase: application.NewUserAccountService(accountRepository),
		},
		DeliveryNotificationController: api.DeliveryNotificationController{
			Account: userAccountService,
			DeliveryNotification: deliveryNotificationService,
		},
	}

	api.RegisterHandlers(
		server, 
		api.NewStrictHandler(&serverInterface, []api.StrictMiddlewareFunc{
				apiKeyMiddleware,
			},
		),
	)

	err = server.Run("0.0.0.0:8080")
	if err != nil {
		return
	}
}

func PrintInfo() {
	fmt.Printf("Running SMS Gateway, version: %s\n", version)
}
