package httphandler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/juhonamnam/wedding-invitation-server/sqldb"
	"github.com/juhonamnam/wedding-invitation-server/types"
)

type GuestbookHandler struct {
	http.Handler
}

func (h *GuestbookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {

		offsetQ := r.URL.Query().Get("offset")
		limitQ := r.URL.Query().Get("limit")

		offset, err := strconv.Atoi(offsetQ)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		limit, err := strconv.Atoi(limitQ)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		guestbook, err := sqldb.GetGuestbook(r.Context(), offset, limit)

		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		pbytes, err := json.Marshal(guestbook)

		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(pbytes)
	} else if r.Method == http.MethodPost {
		decoder := json.NewDecoder(r.Body)
		var post types.GuestbookPostForCreate
		err := decoder.Decode(&post)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		err = sqldb.CreateGuestbookPost(r.Context(), post.Name, post.Content, post.Password)

		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
	} else if r.Method == http.MethodPut {
		decoder := json.NewDecoder(r.Body)
		var post types.GuestbookPostForDelete
		err := decoder.Decode(&post)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		err = sqldb.DeleteGuestbookPost(r.Context(), post.Id, post.Password)

		if err != nil {
			if err.Error() == "INCORRECT_PASSWORD" {
				writeError(w, http.StatusForbidden, err)
			} else {
				writeError(w, http.StatusInternalServerError, err)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
	} else {
		writeError(w, http.StatusMethodNotAllowed, nil)
	}
}
