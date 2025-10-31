package main

import (
	"context"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/juhonamnam/wedding-invitation-server/env"
	"github.com/juhonamnam/wedding-invitation-server/httphandler"
	"github.com/juhonamnam/wedding-invitation-server/sqldb"
	"github.com/rs/cors"
)

func main() {
	ctx := context.Background()

	if env.GCPProjectID == "" {
		log.Fatal("GCP_PROJECT_ID must be set")
	}

	client, err := firestore.NewClient(ctx, env.GCPProjectID)
	if err != nil {
		log.Fatalf("failed to initialize Firestore client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("failed to close Firestore client: %v", err)
		}
	}()

	sqldb.SetClient(client)

	mux := http.NewServeMux()
	mux.Handle("/guestbook", new(httphandler.GuestbookHandler))
	mux.Handle("/attendance", new(httphandler.AttendanceHandler))
	mux.Handle("/admin/api/", new(httphandler.AdminAPIHandler))
	mux.Handle("/admin/import", new(httphandler.AdminImportHandler))
	adminPageHandler := httphandler.NewAdminPageHandler()
	mux.Handle("/admin", adminPageHandler)
	mux.Handle("/admin/", adminPageHandler)

	corHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{env.AllowOrigin},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowCredentials: true,
	})

	handler := corHandler.Handler(mux)

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
