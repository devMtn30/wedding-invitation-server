package sqldb

import (
	"cloud.google.com/go/firestore"
)

var (
	firestoreClient *firestore.Client
)

func SetClient(client *firestore.Client) {
	firestoreClient = client
	if err := initializeGuestbookTable(); err != nil {
		panic(err)
	}
	if err := initializeAttendanceTable(); err != nil {
		panic(err)
	}
}

func getClient() *firestore.Client {
	return firestoreClient
}
