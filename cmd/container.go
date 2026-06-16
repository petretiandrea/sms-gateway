package main

import (
	"context"
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
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

type Container struct {
	ctx     context.Context
	version string
	log     *zap.SugaredLogger

	config        *koanf.Koanf
	tracerCleanup func()

	server *gin.Engine

	mongoClient   *mongodriver.Client
	mongoDatabase *mongodriver.Database

	pushService *infra.FirebasePushNotification

	accountRepository              *mongo.MongoUserAccountRepository
	messageRepository              *mongo.MongoMessageRepository
	phoneRepository                *mongo.MongoPhoneRepository
	deliveryNotificationRepository *mongo.MongoDeliveryNotificationRepository

	changeFeedProducer          *changes.MessageChangeFeedProducer
	webHookNotifier             *api.HttpWebhookNotifier
	userAccountService          *application.UserAccountService
	phoneService                *application.PhoneService
	smsService                  *application.SmsService
	deliveryNotificationService *application.DeliveryNotificationService
	deliveryConsumer            *events.MessageChangeFeedProcessor
}

func NewContainer(ctx context.Context, version string, log *zap.SugaredLogger) *Container {
	return &Container{
		ctx:     ctx,
		version: version,
		log:     log,
	}
}

func (c *Container) Config() (*koanf.Koanf, error) {
	if c.config != nil {
		return c.config, nil
	}

	k := koanf.New(".")
	if err := k.Load(file.Provider("../config/app.yaml"), yaml.Parser()); err != nil {
		return nil, err
	}

	if err := k.Load(env.Provider("ENV", ".", func(s string) string {
		return config.StripUnderscore(strings.ToLower(strings.TrimLeft(s, "ENV_")), ".")
	}), nil); err != nil {
		return nil, err
	}

	c.config = k
	return c.config, nil
}

func (c *Container) InitTracer() error {
	if c.tracerCleanup != nil {
		return nil
	}

	k, err := c.Config()
	if err != nil {
		return err
	}

	cleanupTracer, err := initTracer(OpenTelemetryConfig{
		serviceName:    k.String("app.name"),
		serviceVersion: c.version,
		ctx:            c.ctx,
	})
	if err != nil {
		return err
	}

	c.tracerCleanup = cleanupTracer
	return nil
}

func (c *Container) Server() (*gin.Engine, error) {
	if c.server != nil {
		return c.server, nil
	}

	k, err := c.Config()
	if err != nil {
		return nil, err
	}

	server := gin.New()
	server.Use(func(ctx *gin.Context) {
		if strings.HasSuffix(ctx.Request.URL.Path, "/") {
			ctx.Request.URL.Path = strings.TrimRight(ctx.Request.URL.Path, "/")
			server.HandleContext(ctx)
		} else {
			ctx.Next()
		}
	})

	ginLogger := c.log.Desugar().Named("gin")
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

	mongoClient, err := c.MongoClient()
	if err != nil {
		return nil, err
	}
	health.RegisterGinHealthCheck(server, mongoClient)

	controller, err := c.Controller()
	if err != nil {
		return nil, err
	}
	apiKeyMiddleware, err := c.APIKeyMiddleware()
	if err != nil {
		return nil, err
	}
	api.RegisterHandlers(
		server,
		api.NewStrictHandler(controller, []api.StrictMiddlewareFunc{
			apiKeyMiddleware,
		}),
	)

	c.server = server
	return c.server, nil
}

func (c *Container) MongoClient() (*mongodriver.Client, error) {
	if c.mongoClient != nil {
		return c.mongoClient, nil
	}

	k, err := c.Config()
	if err != nil {
		return nil, err
	}

	client, err := connectMongo(k.String("mongo_connection_string"))
	if err != nil {
		return nil, err
	}

	c.mongoClient = client
	return c.mongoClient, nil
}

func (c *Container) MongoDatabase() (*mongodriver.Database, error) {
	if c.mongoDatabase != nil {
		return c.mongoDatabase, nil
	}

	k, err := c.Config()
	if err != nil {
		return nil, err
	}
	client, err := c.MongoClient()
	if err != nil {
		return nil, err
	}

	c.mongoDatabase = client.Database(k.String("mongo_database_name"))
	return c.mongoDatabase, nil
}

func (c *Container) PushService() (*infra.FirebasePushNotification, error) {
	if c.pushService != nil {
		return c.pushService, nil
	}

	k, err := c.Config()
	if err != nil {
		return nil, err
	}

	credentials := option.WithCredentialsFile(k.String("firebase.credentials_file"))
	app, err := firebase.NewApp(c.ctx, nil, credentials)
	if err != nil {
		return nil, err
	}
	firebaseMessaging, err := app.Messaging(c.ctx)
	if err != nil {
		return nil, err
	}

	pushService := infra.NewFirebasePushNotification(firebaseMessaging)
	if k.Bool("dry_run") {
		pushService.EnableDryRun()
	}

	c.pushService = &pushService
	return c.pushService, nil
}

func (c *Container) AccountRepository() (*mongo.MongoUserAccountRepository, error) {
	if c.accountRepository != nil {
		return c.accountRepository, nil
	}

	db, err := c.MongoDatabase()
	if err != nil {
		return nil, err
	}

	repo := mongo.NewMongoUserAccountRepository(db.Collection("accounts"))
	c.accountRepository = &repo
	return c.accountRepository, nil
}

func (c *Container) MessageRepository() (*mongo.MongoMessageRepository, error) {
	if c.messageRepository != nil {
		return c.messageRepository, nil
	}

	db, err := c.MongoDatabase()
	if err != nil {
		return nil, err
	}

	repo := mongo.NewMongoMessageRepository(db.Collection("messages"))
	c.messageRepository = &repo
	return c.messageRepository, nil
}

func (c *Container) PhoneRepository() (*mongo.MongoPhoneRepository, error) {
	if c.phoneRepository != nil {
		return c.phoneRepository, nil
	}

	db, err := c.MongoDatabase()
	if err != nil {
		return nil, err
	}

	repo := mongo.NewMongoPhoneRepository(db.Collection("phones"))
	c.phoneRepository = &repo
	return c.phoneRepository, nil
}

func (c *Container) DeliveryNotificationRepository() (*mongo.MongoDeliveryNotificationRepository, error) {
	if c.deliveryNotificationRepository != nil {
		return c.deliveryNotificationRepository, nil
	}

	db, err := c.MongoDatabase()
	if err != nil {
		return nil, err
	}

	repo := mongo.NewMongoDeliveryNotificationRepository(db.Collection("deliveryconfigs"))
	c.deliveryNotificationRepository = &repo
	return c.deliveryNotificationRepository, nil
}

func (c *Container) ChangeFeedProducer() *changes.MessageChangeFeedProducer {
	if c.changeFeedProducer == nil {
		c.changeFeedProducer = changes.NewMessageChangeFeedProducer()
	}
	return c.changeFeedProducer
}

func (c *Container) WebHookNotifier() *api.HttpWebhookNotifier {
	if c.webHookNotifier == nil {
		notifier := api.HttpWebhookNotifier{}
		c.webHookNotifier = &notifier
	}
	return c.webHookNotifier
}

func (c *Container) UserAccountService() (*application.UserAccountService, error) {
	if c.userAccountService != nil {
		return c.userAccountService, nil
	}

	repo, err := c.AccountRepository()
	if err != nil {
		return nil, err
	}

	service := application.NewUserAccountService(repo)
	c.userAccountService = &service
	return c.userAccountService, nil
}

func (c *Container) PhoneService() (*application.PhoneService, error) {
	if c.phoneService != nil {
		return c.phoneService, nil
	}

	repo, err := c.PhoneRepository()
	if err != nil {
		return nil, err
	}

	service := application.NewPhoneService(repo)
	c.phoneService = &service
	return c.phoneService, nil
}

func (c *Container) SmsService() (*application.SmsService, error) {
	if c.smsService != nil {
		return c.smsService, nil
	}

	messageRepository, err := c.MessageRepository()
	if err != nil {
		return nil, err
	}
	phoneService, err := c.PhoneService()
	if err != nil {
		return nil, err
	}
	pushService, err := c.PushService()
	if err != nil {
		return nil, err
	}

	service := application.NewSmsService(messageRepository, *phoneService, *pushService, c.ChangeFeedProducer())
	c.smsService = &service
	return c.smsService, nil
}

func (c *Container) DeliveryNotificationService() (*application.DeliveryNotificationService, error) {
	if c.deliveryNotificationService != nil {
		return c.deliveryNotificationService, nil
	}

	deliveryRepository, err := c.DeliveryNotificationRepository()
	if err != nil {
		return nil, err
	}
	messageRepository, err := c.MessageRepository()
	if err != nil {
		return nil, err
	}

	service := application.NewDeliveryNotificationService(deliveryRepository, messageRepository, c.WebHookNotifier())
	c.deliveryNotificationService = &service
	return c.deliveryNotificationService, nil
}

func (c *Container) DeliveryConsumer() (*events.MessageChangeFeedProcessor, error) {
	if c.deliveryConsumer != nil {
		return c.deliveryConsumer, nil
	}

	service, err := c.DeliveryNotificationService()
	if err != nil {
		return nil, err
	}

	consumer := events.NewDeliveryNotificationConsumer(c.ChangeFeedProducer(), *service)
	c.deliveryConsumer = &consumer
	return c.deliveryConsumer, nil
}

func (c *Container) APIKeyMiddleware() (api.StrictMiddlewareFunc, error) {
	userAccountService, err := c.UserAccountService()
	if err != nil {
		return nil, err
	}
	return middleware.NewApiKeyMiddleware(*userAccountService), nil
}

func (c *Container) Controller() (*api.Controller, error) {
	userAccountService, err := c.UserAccountService()
	if err != nil {
		return nil, err
	}
	phoneService, err := c.PhoneService()
	if err != nil {
		return nil, err
	}
	smsService, err := c.SmsService()
	if err != nil {
		return nil, err
	}
	messageRepository, err := c.MessageRepository()
	if err != nil {
		return nil, err
	}
	deliveryNotificationService, err := c.DeliveryNotificationService()
	if err != nil {
		return nil, err
	}

	return &api.Controller{
		AttemptController: api.AttemptController{
			SmsService: *smsService,
		},
		PhoneApiController: api.PhoneApiController{
			Phone:   *phoneService,
			Account: *userAccountService,
		},
		SmsApiController: api.SmsApiController{
			Account:           *userAccountService,
			Sms:               *smsService,
			MessageRepository: messageRepository,
		},
		UserAccountController: api.UserAccountController{
			CreateUserAccountUseCase: *userAccountService,
		},
		DeliveryNotificationController: api.DeliveryNotificationController{
			Account:              *userAccountService,
			DeliveryNotification: *deliveryNotificationService,
		},
	}, nil
}

func (c *Container) StartHTTPServer() error {
	if err := c.InitTracer(); err != nil {
		c.log.Error("Failed to initialize OpenTelemetry!")
	}

	deliveryConsumer, err := c.DeliveryConsumer()
	if err != nil {
		return err
	}
	go deliveryConsumer.Start()

	server, err := c.Server()
	if err != nil {
		return err
	}

	return server.Run("0.0.0.0:8080")
}

func (c *Container) Close() {
	if c.deliveryConsumer != nil {
		c.deliveryConsumer.Stop()
	}
	if c.tracerCleanup != nil {
		c.tracerCleanup()
	}
	if c.mongoClient != nil {
		if err := c.mongoClient.Disconnect(c.ctx); err != nil {
			c.log.Errorw("failed to disconnect mongodb", "error", err)
		}
	}
}
