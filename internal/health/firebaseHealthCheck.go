package health

import (
	"cloud.google.com/go/firestore"
	"context"
)

type FirebaseHealth struct {
	store *firestore.Client
}

func NewFirebaseHealth(store *firestore.Client) FirebaseHealth {
	return FirebaseHealth{store: store}
}

func (f *FirebaseHealth) Check(ctx context.Context) error {
	_, err := f.store.Collections(ctx).GetAll()
	return err
}
