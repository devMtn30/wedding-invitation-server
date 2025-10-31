package httphandler

import (
	"encoding/json"
	"net/http"

	"github.com/juhonamnam/wedding-invitation-server/sqldb"
	"github.com/juhonamnam/wedding-invitation-server/types"
)

type AttendanceHandler struct {
	http.Handler
}

func (h *AttendanceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var attendance types.AttendanceCreate
		err := json.NewDecoder(r.Body).Decode(&attendance)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("BadRequest"))
			return
		}

		if _, err := sqldb.CreateAttendance(r.Context(), attendance.Side, attendance.Name, attendance.Meal, attendance.Count); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("InternalServerError"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
	case http.MethodGet:
		attendances, err := sqldb.ListAttendances(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("InternalServerError"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(attendances)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method Not Allowed"))
	}
}
