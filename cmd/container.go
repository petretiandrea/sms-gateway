package main

import (
	"context"
	"sms-gateway/internal/api"
	"sms-gateway/internal/api/middleware"
	"sms-gateway/internal/application"
	"sms-gateway/internal/config"
	"sms-gateway/internal/health"
	"sms-gateway/internal/infra"
	rabbitmqmessaging "sms-gateway/internal/infra/messaging/rabbitmq"
	"sms-gateway/internal/infra/repos/postgres"
	"strings"
	"time"

	firebase "firebase.google.com/go"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petretiandrea/outbox-go/pkg/outbox"
	outboxpostgres "github.com/petretiandrea/outbox-go/pkg/outbox/postgres"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

type Container struct {
	ctx     context.Context
	version string
	log     *zap.SugaredLogger

	config        *config.AppConfig
	tracerCleanup func()

	server *gin.Engine

	postgresPool *pgxpool.Pool
	postgresDB   *postgres.ContextDB

	pushService     *infra.FirebasePushNotification
	outboxPublisher outbox.Publisher

	accountRepository              *postgres.UserAccountRepository
	messageRepository              *postgres.MessageRepository
	phoneRepository                *postgres.PhoneRepository
	deliveryNotificationRepository *postgres.DeliveryNotificationRepository
	dlqMessageRepository           *postgres.DLQMessageRepository

	webHookNotifier               *api.HttpWebhookNotifier
	userAccountService            *application.UserAccountService
	phoneService                  *application.PhoneService
	smsService                    *application.SmsService
	smsSendProcessor              *application.SMSSendProcessor
	smsAttemptRegisteredProcessor *application.SMSAttemptRegisteredProcessor
	smsOutboxConsumer             *application.SMSOutboxConsumer
	dlqMetrics                    application.DLQMetrics
	dlqConsumer                   *application.DLQConsumer
	deliveryNotificationService   *application.DeliveryNotificationService
	smsSendConsumer               *rabbitmqmessaging.Consumer
	smsDLQConsumer                *rabbitmqmessaging.Consumer
}

func NewContainer(ctx context.Context, version string, log *zap.SugaredLogger) *Container {
	return &Container{
		ctx:     ctx,
		version: version,
		log:     log,
	}
}

func (c *Container) Config() (*config.AppConfig, error) {
	if c.config != nil {
		return c.config, nil
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	c.config = &cfg
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
		serviceName:    k.AppName,
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
		k.AppName,
		otelgin.WithFilter(health.FilterHealthCheck),
	))

	db, err := c.PostgresPool()
	if err != nil {
		return nil, err
	}
	health.RegisterGinHealthCheck(server, db)

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

func (c *Container) PostgresPool() (*pgxpool.Pool, error) {
	if c.postgresPool != nil {
		return c.postgresPool, nil
	}

	dsn, err := c.PostgresDSN()
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.New(c.ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(c.ctx); err != nil {
		pool.Close()
		return nil, err
	}

	c.postgresPool = pool
	return c.postgresPool, nil
}

func (c *Container) PostgresDB() (*postgres.ContextDB, error) {
	if c.postgresDB != nil {
		return c.postgresDB, nil
	}

	pool, err := c.PostgresPool()
	if err != nil {
		return nil, err
	}

	c.postgresDB = postgres.NewContextDB(pool)
	return c.postgresDB, nil
}

func (c *Container) PostgresDSN() (string, error) {
	cfg, err := c.Config()
	if err != nil {
		return "", err
	}

	return cfg.Postgres.DSN, nil
}

func (c *Container) RabbitMQDSN() (string, error) {
	cfg, err := c.Config()
	if err != nil {
		return "", err
	}

	return cfg.RabbitMQ.DSN, nil
}

func (c *Container) PushService() (*infra.FirebasePushNotification, error) {
	if c.pushService != nil {
		return c.pushService, nil
	}

	cfg, err := c.Config()
	if err != nil {
		return nil, err
	}

	credentials := option.WithCredentialsFile(cfg.Firebase.CredentialsFile)
	app, err := firebase.NewApp(c.ctx, nil, credentials)
	if err != nil {
		return nil, err
	}
	firebaseMessaging, err := app.Messaging(c.ctx)
	if err != nil {
		return nil, err
	}

	pushService := infra.NewFirebasePushNotification(firebaseMessaging)
	if cfg.DryRun {
		pushService.EnableDryRun()
	}

	c.pushService = &pushService
	return c.pushService, nil
}

func (c *Container) AccountRepository() (*postgres.UserAccountRepository, error) {
	if c.accountRepository != nil {
		return c.accountRepository, nil
	}

	db, err := c.PostgresDB()
	if err != nil {
		return nil, err
	}

	repo := postgres.NewUserAccountRepository(db)
	c.accountRepository = &repo
	return c.accountRepository, nil
}

func (c *Container) MessageRepository() (*postgres.MessageRepository, error) {
	if c.messageRepository != nil {
		return c.messageRepository, nil
	}

	db, err := c.PostgresDB()
	if err != nil {
		return nil, err
	}

	repo := postgres.NewMessageRepository(db)
	c.messageRepository = &repo
	return c.messageRepository, nil
}

func (c *Container) PhoneRepository() (*postgres.PhoneRepository, error) {
	if c.phoneRepository != nil {
		return c.phoneRepository, nil
	}

	db, err := c.PostgresDB()
	if err != nil {
		return nil, err
	}

	repo := postgres.NewPhoneRepository(db)
	c.phoneRepository = &repo
	return c.phoneRepository, nil
}

func (c *Container) DeliveryNotificationRepository() (*postgres.DeliveryNotificationRepository, error) {
	if c.deliveryNotificationRepository != nil {
		return c.deliveryNotificationRepository, nil
	}

	db, err := c.PostgresDB()
	if err != nil {
		return nil, err
	}

	repo := postgres.NewDeliveryNotificationRepository(db)
	c.deliveryNotificationRepository = &repo
	return c.deliveryNotificationRepository, nil
}

func (c *Container) DLQMessageRepository() (*postgres.DLQMessageRepository, error) {
	if c.dlqMessageRepository != nil {
		return c.dlqMessageRepository, nil
	}

	db, err := c.PostgresDB()
	if err != nil {
		return nil, err
	}

	repo := postgres.NewDLQMessageRepository(db)
	c.dlqMessageRepository = &repo
	return c.dlqMessageRepository, nil
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

func (c *Container) OutboxPublisher() (outbox.Publisher, error) {
	if c.outboxPublisher != nil {
		return c.outboxPublisher, nil
	}

	db, err := c.PostgresDB()
	if err != nil {
		return nil, err
	}

	c.outboxPublisher, err = outboxpostgres.NewPublisher(db, outboxpostgres.PublisherConfig{
		TableName: "outbox_messages",
	})
	if err != nil {
		return nil, err
	}
	return c.outboxPublisher, nil

}

func (c *Container) SmsService() (*application.SmsService, error) {
	if c.smsService != nil {
		return c.smsService, nil
	}

	uow, err := c.PostgresDB()
	if err != nil {
		return nil, err
	}

	messageRepository, err := c.MessageRepository()
	if err != nil {
		return nil, err
	}
	phoneService, err := c.PhoneService()
	if err != nil {
		return nil, err
	}

	outboxPublisher, err := c.OutboxPublisher()
	if err != nil {
		return nil, err
	}

	service := application.NewSmsService(messageRepository, *phoneService, uow, outboxPublisher)
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

func (c *Container) SMSSendProcessor() (*application.SMSSendProcessor, error) {
	if c.smsSendProcessor != nil {
		return c.smsSendProcessor, nil
	}

	messageRepository, err := c.MessageRepository()
	if err != nil {
		return nil, err
	}
	phoneRepository, err := c.PhoneRepository()
	if err != nil {
		return nil, err
	}
	pushService, err := c.PushService()
	if err != nil {
		return nil, err
	}

	c.smsSendProcessor = application.NewSMSSendProcessor(messageRepository, phoneRepository, pushService)
	return c.smsSendProcessor, nil
}

func (c *Container) SMSAttemptRegisteredProcessor() (*application.SMSAttemptRegisteredProcessor, error) {
	if c.smsAttemptRegisteredProcessor != nil {
		return c.smsAttemptRegisteredProcessor, nil
	}

	messageRepository, err := c.MessageRepository()
	if err != nil {
		return nil, err
	}
	deliveryNotificationService, err := c.DeliveryNotificationService()
	if err != nil {
		return nil, err
	}

	c.smsAttemptRegisteredProcessor = application.NewSMSAttemptRegisteredProcessor(messageRepository, deliveryNotificationService)
	return c.smsAttemptRegisteredProcessor, nil
}

func (c *Container) SMSOutboxConsumer() (*application.SMSOutboxConsumer, error) {
	if c.smsOutboxConsumer != nil {
		return c.smsOutboxConsumer, nil
	}

	smsSendProcessor, err := c.SMSSendProcessor()
	if err != nil {
		return nil, err
	}
	smsAttemptRegisteredProcessor, err := c.SMSAttemptRegisteredProcessor()
	if err != nil {
		return nil, err
	}

	c.smsOutboxConsumer = application.NewSMSOutboxConsumer(smsSendProcessor, smsAttemptRegisteredProcessor)
	return c.smsOutboxConsumer, nil
}

func (c *Container) DLQMetrics() (application.DLQMetrics, error) {
	if c.dlqMetrics != nil {
		return c.dlqMetrics, nil
	}

	metrics, err := application.NewDLQMetrics()
	if err != nil {
		return nil, err
	}

	c.dlqMetrics = metrics
	return c.dlqMetrics, nil
}

func (c *Container) DLQConsumer() (*application.DLQConsumer, error) {
	if c.dlqConsumer != nil {
		return c.dlqConsumer, nil
	}

	repo, err := c.DLQMessageRepository()
	if err != nil {
		return nil, err
	}
	metrics, err := c.DLQMetrics()
	if err != nil {
		return nil, err
	}

	c.dlqConsumer = application.NewDLQConsumer(repo, metrics)
	return c.dlqConsumer, nil
}

func (c *Container) SMSSendConsumer() (*rabbitmqmessaging.Consumer, error) {
	if c.smsSendConsumer != nil {
		return c.smsSendConsumer, nil
	}

	dsn, err := c.RabbitMQDSN()
	if err != nil {
		return nil, err
	}
	consumer, err := c.SMSOutboxConsumer()
	if err != nil {
		return nil, err
	}

	c.smsSendConsumer = rabbitmqmessaging.NewConsumer(
		dsn,
		rabbitmqmessaging.QueueSMSSendInternal,
		consumer,
		c.log.Desugar(),
	)
	return c.smsSendConsumer, nil
}

func (c *Container) SMSDLQConsumer() (*rabbitmqmessaging.Consumer, error) {
	if c.smsDLQConsumer != nil {
		return c.smsDLQConsumer, nil
	}

	dsn, err := c.RabbitMQDSN()
	if err != nil {
		return nil, err
	}
	consumer, err := c.DLQConsumer()
	if err != nil {
		return nil, err
	}

	c.smsDLQConsumer = rabbitmqmessaging.NewConsumer(
		dsn,
		rabbitmqmessaging.QueueSMSSendDeadLetter,
		consumer,
		c.log.Desugar(),
	)
	return c.smsDLQConsumer, nil
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

	smsSendConsumer, err := c.SMSSendConsumer()
	if err != nil {
		return err
	}
	go smsSendConsumer.Start(c.ctx)

	smsDLQConsumer, err := c.SMSDLQConsumer()
	if err != nil {
		return err
	}
	go smsDLQConsumer.Start(c.ctx)

	server, err := c.Server()
	if err != nil {
		return err
	}

	return server.Run("0.0.0.0:8080")
}

func (c *Container) Close() {
	if c.smsSendConsumer != nil {
		c.smsSendConsumer.Stop()
	}
	if c.smsDLQConsumer != nil {
		c.smsDLQConsumer.Stop()
	}
	if c.tracerCleanup != nil {
		c.tracerCleanup()
	}
	if c.postgresPool != nil {
		c.postgresPool.Close()
	}
}
