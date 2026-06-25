package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sms-gateway/internal/domain"
	mongorepos "sms-gateway/internal/infra/repos/mongo"
	"sms-gateway/internal/infra/repos/postgres"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	mongooptions "go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type mongoCopyOptions struct {
	mongoURI                  string
	mongoDatabase             string
	postgresDSN               string
	accountsCollection        string
	phonesCollection          string
	messagesCollection        string
	deliveryConfigsCollection string
}

type mongoDeliveryNotificationConfigEntity struct {
	ID            string `bson:"_id"`
	AccountID     string `bson:"accountId"`
	WebhookURL    string `bson:"webhookURL"`
	WebhookURLAlt string `bson:"webhookUrl"`
	Enabled       bool   `bson:"enabled"`
}

func newMongoDBCommand(log *zap.SugaredLogger) *cobra.Command {
	command := &cobra.Command{
		Use:   "mongodb",
		Short: "MongoDB maintenance commands",
	}

	command.AddCommand(newMongoDBCopyToPostgresCommand(log))
	return command
}

func newMongoDBCopyToPostgresCommand(log *zap.SugaredLogger) *cobra.Command {
	options := mongoCopyOptions{
		accountsCollection:        "accounts",
		phonesCollection:          "phones",
		messagesCollection:        "messages",
		deliveryConfigsCollection: "delivery_notification_configs",
	}

	command := &cobra.Command{
		Use:   "copy-to-postgres",
		Short: "Copy legacy MongoDB data into PostgreSQL",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMongoDBCopyToPostgres(cmd.Context(), options, log)
		},
	}

	command.Flags().StringVar(&options.mongoURI, "mongo-uri", "", "MongoDB URI. Defaults to ENV_MONGODB__URI, MONGODB_URI, MONGOCONNECTIONSTRING.")
	command.Flags().StringVar(&options.mongoDatabase, "mongo-database", "", "MongoDB database. Defaults to ENV_MONGODB__DATABASE, MONGODB_DATABASE, MONGODATABASENAME.")
	command.Flags().StringVar(&options.postgresDSN, "postgres-dsn", "", "Postgres DSN. Defaults to ENV_POSTGRES__DSN or POSTGRES_DSN.")
	command.Flags().StringVar(&options.accountsCollection, "accounts-collection", options.accountsCollection, "MongoDB accounts collection name.")
	command.Flags().StringVar(&options.phonesCollection, "phones-collection", options.phonesCollection, "MongoDB phones collection name.")
	command.Flags().StringVar(&options.messagesCollection, "messages-collection", options.messagesCollection, "MongoDB messages collection name.")
	command.Flags().StringVar(&options.deliveryConfigsCollection, "delivery-configs-collection", options.deliveryConfigsCollection, "MongoDB delivery notification configs collection name.")

	return command
}

func runMongoDBCopyToPostgres(ctx context.Context, options mongoCopyOptions, log *zap.SugaredLogger) error {
	options.applyEnvDefaults()
	if err := options.validate(); err != nil {
		return err
	}

	mongoClient, err := mongo.Connect(ctx, mongooptions.Client().ApplyURI(options.mongoURI))
	if err != nil {
		return fmt.Errorf("connect mongodb: %w", err)
	}
	defer mongoClient.Disconnect(ctx)

	postgresPool, err := pgxpool.New(ctx, options.postgresDSN)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer postgresPool.Close()

	if err := postgresPool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	db := mongoClient.Database(options.mongoDatabase)
	postgresDB := postgres.NewContextDB(postgresPool)

	accountRepo := postgres.NewUserAccountRepository(postgresDB)
	phoneRepo := postgres.NewPhoneRepository(postgresDB)
	deliveryRepo := postgres.NewDeliveryNotificationRepository(postgresDB)
	messageRepo := postgres.NewMessageRepository(postgresDB)

	var total int

	copied, err := copyMongoCollection(ctx, db.Collection(options.accountsCollection), func(entity mongorepos.UserAccountJsonEntity) error {
		_, err := accountRepo.Save(ctx, domain.UserAccount{
			Id:          domain.AccountID(entity.Id),
			Phone:       entity.Phone,
			ApiKey:      domain.ApiKey(entity.ApiKey),
			IsSuspended: entity.IsSuspended,
			CreatedAt:   entity.CreatedAt,
		})
		return err
	})
	if err != nil {
		return fmt.Errorf("copy accounts: %w", err)
	}
	total += copied
	log.Infow("mongodb accounts copied", "count", copied)

	copied, err = copyMongoCollection(ctx, db.Collection(options.phonesCollection), func(entity mongorepos.PhoneJsonEntity) error {
		_, err := phoneRepo.Save(ctx, domain.Phone{
			Id:        domain.PhoneId(entity.Id),
			Phone:     domain.PhoneNumber{Number: entity.Phone},
			UserId:    domain.AccountID(entity.Account),
			Token:     domain.FCMToken(entity.FCMToken),
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
		})
		return err
	})
	if err != nil {
		return fmt.Errorf("copy phones: %w", err)
	}
	total += copied
	log.Infow("mongodb phones copied", "count", copied)

	copied, err = copyMongoCollection(ctx, db.Collection(options.deliveryConfigsCollection), func(entity mongoDeliveryNotificationConfigEntity) error {
		accountID := entity.accountID()
		if accountID == "" {
			return errors.New("delivery notification config missing account id")
		}

		_, err := deliveryRepo.Save(ctx, domain.DeliveryNotificationConfig{
			AccountId:  domain.AccountID(accountID),
			WebhookURL: entity.webhookURL(),
			Enabled:    entity.Enabled,
		})
		return err
	})
	if err != nil {
		return fmt.Errorf("copy delivery notification configs: %w", err)
	}
	total += copied
	log.Infow("mongodb delivery notification configs copied", "count", copied)

	copied, err = copyMongoCollection(ctx, db.Collection(options.messagesCollection), func(entity mongorepos.MongoMessageEntity) error {
		message := entity.ToMessage(entity.Id)
		_, err := messageRepo.Save(ctx, message)
		return err
	})
	if err != nil {
		return fmt.Errorf("copy sms messages: %w", err)
	}
	total += copied
	log.Infow("mongodb sms messages copied", "count", copied)
	log.Infow("mongodb copy to postgres completed", "total", total)

	return nil
}

func (entity mongoDeliveryNotificationConfigEntity) accountID() string {
	if entity.AccountID != "" {
		return entity.AccountID
	}
	return entity.ID
}

func (entity mongoDeliveryNotificationConfigEntity) webhookURL() string {
	if entity.WebhookURL != "" {
		return entity.WebhookURL
	}
	return entity.WebhookURLAlt
}

func (options *mongoCopyOptions) applyEnvDefaults() {
	if options.mongoURI == "" {
		options.mongoURI = firstNonEmptyEnv("ENV_MONGODB__URI", "MONGODB_URI", "MONGOCONNECTIONSTRING")
	}
	if options.mongoDatabase == "" {
		options.mongoDatabase = firstNonEmptyEnv("ENV_MONGODB__DATABASE", "MONGODB_DATABASE", "MONGODATABASENAME")
	}
	if options.postgresDSN == "" {
		options.postgresDSN = firstNonEmptyEnv("ENV_POSTGRES__DSN", "POSTGRES_DSN")
	}
}

func (options mongoCopyOptions) validate() error {
	switch {
	case options.mongoURI == "":
		return errors.New("mongodb uri is required")
	case options.mongoDatabase == "":
		return errors.New("mongodb database is required")
	case options.postgresDSN == "":
		return errors.New("postgres dsn is required")
	default:
		return nil
	}
}

func firstNonEmptyEnv(names ...string) string {
	for _, name := range names {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return ""
}

func copyMongoCollection[T any](
	ctx context.Context,
	collection *mongo.Collection,
	save func(T) error,
) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var copied int
	for cursor.Next(ctx) {
		var entity T
		if err := cursor.Decode(&entity); err != nil {
			return copied, err
		}
		if err := save(entity); err != nil {
			return copied, err
		}
		copied++
	}
	if err := cursor.Err(); err != nil {
		return copied, err
	}
	return copied, nil
}
