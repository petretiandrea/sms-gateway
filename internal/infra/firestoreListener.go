package infra

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"time"
)

type FirestoreEventListener struct {
	context    context.Context
	collection string
	firestore  *firestore.Client
	outbox     chan firestore.DocumentChange
}

func NewFirestoreEventListener(ctx context.Context, client *firestore.Client, collection string) FirestoreEventListener {
	return FirestoreEventListener{
		context:    ctx,
		firestore:  client,
		collection: collection,
		outbox:     make(chan firestore.DocumentChange, 10),
	}
}

func (listener *FirestoreEventListener) Changes() <-chan firestore.DocumentChange {
	return listener.outbox
}

func (listener *FirestoreEventListener) ListenChanges(resumeAt time.Time) {
	docReference := listener.firestore.Collection(listener.collection)
	docIterator := docReference.Where("updatedAt", ">=", resumeAt).Snapshots(listener.context)
	defer docIterator.Stop()
	for {
		snapshot, err := docIterator.Next()
		if err != nil {
			fmt.Println(err)
		}

		for _, change := range snapshot.Changes {
			listener.outbox <- change
		}
	}
}

func (listener *FirestoreEventListener) StopListenChanges() {
	fmt.Println("Stop listen changes")
	close(listener.outbox)
}
