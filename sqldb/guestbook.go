package sqldb

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/juhonamnam/wedding-invitation-server/env"
	"github.com/juhonamnam/wedding-invitation-server/types"
	"github.com/juhonamnam/wedding-invitation-server/util"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func initializeGuestbookTable() error {
	return nil
}

type guestbookRecord struct {
	ID        int    `firestore:"id"`
	Name      string `firestore:"name"`
	Content   string `firestore:"content"`
	Password  string `firestore:"password"`
	Timestamp int64  `firestore:"timestamp"`
	Valid     bool   `firestore:"valid"`
}

func GetGuestbook(ctx context.Context, offset, limit int) (*types.GuestbookGetResponse, error) {
	client := getClient()
	if client == nil {
		return nil, fmt.Errorf("firestore client not initialized")
	}

	query := client.Collection("guestbook").
		Where("valid", "==", true).
		OrderBy("timestamp", firestore.Desc).
		Offset(offset).
		Limit(limit)

	iter := query.Documents(ctx)
	defer iter.Stop()

	response := &types.GuestbookGetResponse{
		Posts: []types.GuestbookPostForGet{},
	}

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var record guestbookRecord
		if err := doc.DataTo(&record); err != nil {
			return nil, err
		}

		response.Posts = append(response.Posts, types.GuestbookPostForGet{
			Id:        record.ID,
			Name:      record.Name,
			Content:   record.Content,
			Timestamp: uint64(record.Timestamp),
		})
	}

	total, err := countGuestbook(ctx, client)
	if err != nil {
		return nil, err
	}
	response.Total = total

	return response, nil
}

func countGuestbook(ctx context.Context, client *firestore.Client) (int, error) {
	iter := client.Collection("guestbook").
		Where("valid", "==", true).
		Documents(ctx)
	defer iter.Stop()

	total := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, err
		}
		total++
	}

	return total, nil
}

func CreateGuestbookPost(ctx context.Context, name, content, password string) error {
	client := getClient()
	if client == nil {
		return fmt.Errorf("firestore client not initialized")
	}

	phash, err := util.HashPassword(password)
	if err != nil {
		return err
	}

	id, err := nextID(ctx, "guestbook")
	if err != nil {
		return err
	}

	record := guestbookRecord{
		ID:        id,
		Name:      name,
		Content:   content,
		Password:  phash,
		Timestamp: time.Now().Unix(),
		Valid:     true,
	}

	docRef := client.Collection("guestbook").Doc(strconv.Itoa(id))
	_, err = docRef.Set(ctx, record)
	return err
}

func DeleteGuestbookPost(ctx context.Context, id int, password string) error {
	client := getClient()
	if client == nil {
		return fmt.Errorf("firestore client not initialized")
	}

	docRef := client.Collection("guestbook").Doc(strconv.Itoa(id))
	snap, err := docRef.Get(ctx)
	if status.Code(err) == codes.NotFound {
		return fmt.Errorf("NO_GUESTBOOK_POST_FOUND")
	}
	if err != nil {
		return err
	}

	var record guestbookRecord
	if err := snap.DataTo(&record); err != nil {
		return err
	}

	if !record.Valid {
		return fmt.Errorf("NO_GUESTBOOK_POST_FOUND")
	}

	passwordMatch := false
	if env.AdminPassword != "" && env.AdminPassword == password {
		passwordMatch = true
	} else if util.CheckPasswordHash(password, record.Password) {
		passwordMatch = true
	}

	if !passwordMatch {
		return fmt.Errorf("INCORRECT_PASSWORD")
	}

	_, err = docRef.Set(ctx, map[string]interface{}{
		"valid": false,
	}, firestore.MergeAll)
	return err
}

func UpdateGuestbookPost(ctx context.Context, id int, name, content string, password *string) error {
	client := getClient()
	if client == nil {
		return fmt.Errorf("firestore client not initialized")
	}

	docRef := client.Collection("guestbook").Doc(strconv.Itoa(id))
	_, err := docRef.Get(ctx)
	if status.Code(err) == codes.NotFound {
		return fmt.Errorf("NO_GUESTBOOK_POST_FOUND")
	}
	if err != nil {
		return err
	}

	var updates = map[string]interface{}{
		"name":    name,
		"content": content,
	}

	if password != nil {
		if *password == "" {
			updates["password"] = ""
		} else {
			hash, err := util.HashPassword(*password)
			if err != nil {
				return err
			}
			updates["password"] = hash
		}
	}

	_, err = docRef.Set(ctx, updates, firestore.MergeAll)
	return err
}

func ImportGuestbook(ctx context.Context, data *types.GuestbookImport) (int, error) {
	if data == nil || len(data.Posts) == 0 {
		return 0, nil
	}

	client := getClient()
	if client == nil {
		return 0, fmt.Errorf("firestore client not initialized")
	}

	batch := client.Batch()
	maxID := 0
	for _, post := range data.Posts {
		record := guestbookRecord{
			ID:        post.ID,
			Name:      post.Name,
			Content:   post.Content,
			Password:  "",
			Timestamp: int64(post.Timestamp),
			Valid:     true,
		}

		if post.Password != "" {
			hash, err := util.HashPassword(post.Password)
			if err != nil {
				return 0, err
			}
			record.Password = hash
		}

		docRef := client.Collection("guestbook").Doc(strconv.Itoa(post.ID))
		batch.Set(docRef, record)

		if post.ID > maxID {
			maxID = post.ID
		}
	}

	if _, err := batch.Commit(ctx); err != nil {
		return 0, err
	}

	if maxID > 0 {
		if err := ensureCounterAtLeast(ctx, "guestbook", maxID); err != nil {
			return len(data.Posts), err
		}
	}

	return len(data.Posts), nil
}
