package postgres

import (
	"context"
	"sms-gateway/internal/domain"
	"time"
)

type PhoneRepository struct {
	db DBTX
}

func NewPhoneRepository(db DBTX) PhoneRepository {
	return PhoneRepository{db: db}
}

func (r *PhoneRepository) Save(ctx context.Context, phone domain.Phone) (*domain.Phone, error) {
	updatedAt := time.Now()
	if !phone.UpdatedAt.IsZero() {
		updatedAt = phone.UpdatedAt
	}
	createdAt := phone.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	_, err := r.db.Exec(ctx, `
INSERT INTO phones (id, phone, account_id, fcm_token, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE SET
    phone = EXCLUDED.phone,
    account_id = EXCLUDED.account_id,
    fcm_token = EXCLUDED.fcm_token,
    updated_at = EXCLUDED.updated_at`,
		string(phone.Id),
		phone.Phone.Number,
		string(phone.UserId),
		string(phone.Token),
		createdAt,
		updatedAt,
	)
	if err != nil {
		return nil, err
	}

	phone.CreatedAt = createdAt
	phone.UpdatedAt = updatedAt
	return &phone, nil
}

func (r *PhoneRepository) FindById(ctx context.Context, id domain.PhoneId) *domain.Phone {
	return r.findOne(ctx, "id = $1", string(id))
}

func (r *PhoneRepository) FindByPhoneNumber(ctx context.Context, number domain.PhoneNumber) *domain.Phone {
	return r.findOne(ctx, "phone = $1", number.Number)
}

func (r *PhoneRepository) Delete(ctx context.Context, id domain.PhoneId) bool {
	tag, err := r.db.Exec(ctx, `DELETE FROM phones WHERE id = $1`, string(id))
	return err == nil && tag.RowsAffected() > 0
}

func (r *PhoneRepository) findOne(ctx context.Context, predicate string, arg any) *domain.Phone {
	row := r.db.QueryRow(ctx, `
SELECT id, phone, account_id, fcm_token, created_at, updated_at
FROM phones
WHERE `+predicate, arg)

	var phone domain.Phone
	var id string
	var phoneNumber string
	var accountID string
	var token string
	if err := row.Scan(&id, &phoneNumber, &accountID, &token, &phone.CreatedAt, &phone.UpdatedAt); err != nil {
		return nil
	}

	phone.Id = domain.PhoneId(id)
	phone.Phone = domain.PhoneNumber{Number: phoneNumber}
	phone.UserId = domain.AccountID(accountID)
	phone.Token = domain.FCMToken(token)
	return &phone
}

var _ domain.PhoneRepository = (*PhoneRepository)(nil)
