package sqldb

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/juhonamnam/wedding-invitation-server/types"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func initializeAttendanceTable() error {
	return nil
}

func CreateAttendance(ctx context.Context, side, name, meal string, count int) (int, error) {
	client := getClient()
	if client == nil {
		return 0, fmt.Errorf("firestore client not initialized")
	}

	id, err := nextID(ctx, "attendance")
	if err != nil {
		return 0, err
	}

	attendance := types.Attendance{
		ID:        id,
		Side:      side,
		Name:      name,
		Meal:      meal,
		Count:     count,
		Timestamp: time.Now().Unix(),
	}

	docRef := client.Collection("attendance").Doc(strconv.Itoa(attendance.ID))
	_, err = docRef.Set(ctx, attendance)
	return attendance.ID, err
}

func ListAttendances(ctx context.Context) ([]types.Attendance, error) {
	client := getClient()
	if client == nil {
		return nil, fmt.Errorf("firestore client not initialized")
	}

	iter := client.Collection("attendance").
		OrderBy("timestamp", firestore.Desc).
		Documents(ctx)

	defer iter.Stop()

	var attendances []types.Attendance
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var attendance types.Attendance
		if err := doc.DataTo(&attendance); err != nil {
			return nil, err
		}
		attendances = append(attendances, attendance)
	}

	return attendances, nil
}

func ImportAttendances(ctx context.Context, attendances []types.Attendance) (int, error) {
	if len(attendances) == 0 {
		return 0, nil
	}

	client := getClient()
	if client == nil {
		return 0, fmt.Errorf("firestore client not initialized")
	}

	batch := client.Batch()
	maxID := 0
	for _, attendance := range attendances {
		record := types.Attendance{
			ID:        attendance.ID,
			Side:      attendance.Side,
			Name:      attendance.Name,
			Meal:      attendance.Meal,
			Count:     attendance.Count,
			Timestamp: attendance.Timestamp,
		}

		docRef := client.Collection("attendance").Doc(strconv.Itoa(attendance.ID))
		batch.Set(docRef, record)

		if attendance.ID > maxID {
			maxID = attendance.ID
		}
	}

	if _, err := batch.Commit(ctx); err != nil {
		return 0, err
	}

	if maxID > 0 {
		if err := ensureCounterAtLeast(ctx, "attendance", maxID); err != nil {
			return len(attendances), err
		}
	}

	return len(attendances), nil
}

func UpdateAttendance(ctx context.Context, attendance types.Attendance) error {
	client := getClient()
	if client == nil {
		return fmt.Errorf("firestore client not initialized")
	}

	docRef := client.Collection("attendance").Doc(strconv.Itoa(attendance.ID))
	if _, err := docRef.Get(ctx); status.Code(err) == codes.NotFound {
		return fmt.Errorf("ATTENDANCE_NOT_FOUND")
	} else if err != nil {
		return err
	}

	_, err := docRef.Set(ctx, map[string]interface{}{
		"side":      attendance.Side,
		"name":      attendance.Name,
		"meal":      attendance.Meal,
		"count":     attendance.Count,
		"timestamp": attendance.Timestamp,
	}, firestore.MergeAll)
	return err
}

func DeleteAttendance(ctx context.Context, id int) error {
	client := getClient()
	if client == nil {
		return fmt.Errorf("firestore client not initialized")
	}

	docRef := client.Collection("attendance").Doc(strconv.Itoa(id))
	if _, err := docRef.Get(ctx); status.Code(err) == codes.NotFound {
		return fmt.Errorf("ATTENDANCE_NOT_FOUND")
	} else if err != nil {
		return err
	}

	_, err := docRef.Delete(ctx)
	return err
}
