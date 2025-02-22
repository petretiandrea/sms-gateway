package main

import (
	"context"
	_ "embed"
	"sms-gateway/internal/api"
	"sms-gateway/internal/api/middleware"
	"sms-gateway/internal/application"
	"sms-gateway/internal/config"
	"sms-gateway/internal/events"
	"sms-gateway/internal/health"
	"sms-gateway/internal/infra"
	"sms-gateway/internal/infra/changes"
	"sms-gateway/internal/infra/repos/mongo"
	"strings"
	"time"

	firebase "firebase.google.com/go"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

//go:embed version.txt
var version string

func main() {
	log := zap.Must(zap.NewProduction()).Sugar()
	zap.ReplaceGlobals(log.Desugar())
	defer log.Sync()

	log.Infof("Running SMS Gateway, version: %s\n", version)

	k := koanf.New(".")
	if err := k.Load(file.Provider("../config/app.yaml"), yaml.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	// Load environment variables
	if err := k.Load(env.Provider("ENV", ".", func(s string) string {
		return config.StripUnderscore(strings.ToLower(strings.TrimLeft(s, "ENV_")), ".")
	}), nil); err != nil {
		log.Fatalf("error loading env variables %v", err)
	}

	cleanupTracer, errTracer := initTracer(OpenTelemetryConfig{
		serviceName:    k.String("app.name"),
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

	ginLogger := log.Desugar().Named("gin")
	server.Use(ginzap.GinzapWithConfig(ginLogger, &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		SkipPaths:  []string{"/health"},
	}))
	server.Use(ginzap.RecoveryWithZap(ginLogger, true))
	server.Use(otelgin.Middleware(
		k.String("app.name"),
		otelgin.WithFilter(health.FilterHealthCheck),
	))

	// create async firebase ctx
	ctx := context.Background()
	credentials := option.WithCredentialsFile(k.String("firebase.credentials_file"))
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
	pushService := infra.NewFirebasePushNotification(firebaseMessaging)
	if k.Bool("dry_run") {
		pushService.EnableDryRun()
	}

	// initialize mongo
	mongoClient, err := connectMongo(k.String("mongo_connection_string"))
	if err != nil {
		log.Error("Failed to initialize mongodb")
		return
	}
	mongoDatabase := mongoClient.Database(k.String("mongo_database_name"))

	// user account example
	accountRepository := mongo.NewMongoUserAccountRepository(mongoDatabase.Collection("accounts"))
	messageRepository := mongo.NewMongoMessageRepository(mongoDatabase.Collection("messages"))
	phoneRepository := mongo.NewMongoPhoneRepository(mongoDatabase.Collection("phones"))
	deliveryNotificationRepo := mongo.NewMongoDeliveryNotificationRepository(
		mongoDatabase.Collection("deliveryconfigs"),
	)

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

