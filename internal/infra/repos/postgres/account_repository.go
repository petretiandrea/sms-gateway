package postgres

import (
	"context"
	"sms-gateway/internal/domain"
	"time"
)

type UserAccountRepository struct {
	db DBTX
}

func NewUserAccountRepository(db DBTX) UserAccountRepository {
	return UserAccountRepository{db: db}
}

func (r UserAccountRepository) Save(ctx context.Context, account domain.UserAccount) (bool, error) {
	createdAt := account.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	_, err := r.db.Exec(ctx, `
INSERT INTO accounts (id, phone, api_key, is_suspended, created_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET
    phone = EXCLUDED.phone,
    api_key = EXCLUDED.api_key,
    is_suspended = EXCLUDED.is_suspended`,
		string(account.Id),
		account.Phone,
		string(account.ApiKey),
		account.IsSuspended,
		createdAt,
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r UserAccountRepository) FindById(ctx context.Context, accountID domain.AccountID) *domain.UserAccount {
	return r.findOne(ctx, "id = $1", string(accountID))
}

func (r UserAccountRepository) FindByApiKey(ctx context.Context, apiKey domain.ApiKey) *domain.UserAccount {
	return r.findOne(ctx, "api_key = $1", string(apiKey))
}

func (r UserAccountRepository) findOne(ctx context.Context, predicate string, arg any) *domain.UserAccount {
	row := r.db.QueryRow(ctx, `
SELECT id, phone, api_key, is_suspended, created_at
FROM accounts
WHERE `+predicate, arg)

	var account domain.UserAccount
	var id string
	var apiKey string
	if err := row.Scan(&id, &account.Phone, &apiKey, &account.IsSuspended, &account.CreatedAt); err != nil {
		return nil
	}

	account.Id = domain.AccountID(id)
	account.ApiKey = domain.ApiKey(apiKey)
	return &account
}

var _ domain.UserAccountRepository = (*UserAccountRepository)(nil)
