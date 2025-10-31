package types

type AttendanceCreate struct {
	Side  string `json:"side"`
	Name  string `json:"name"`
	Meal  string `json:"meal"`
	Count int    `json:"count"`
}

type Attendance struct {
	ID        int    `json:"id" firestore:"id"`
	Side      string `json:"side" firestore:"side"`
	Name      string `json:"name" firestore:"name"`
	Meal      string `json:"meal" firestore:"meal"`
	Count     int    `json:"count" firestore:"count"`
	Timestamp int64  `json:"timestamp" firestore:"timestamp"`
}
