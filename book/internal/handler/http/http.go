package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/tj330/bookapp/book/internal/controller/book"
)

type Handler struct {
	ctrl *book.Controller
}

func New(ctrl *book.Controller) *Handler {
	return &Handler{ctrl}
}

func (h *Handler) GetBookDetails(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	details, err := h.ctrl.Get(r.Context(), id)
	if err != nil && errors.Is(book.ErrNotFound, err) {
		w.WriteHeader(http.StatusNotFound)
	} else if err != nil {
		log.Printf("Repository get error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	if err := json.NewEncoder(w).Encode(details); err != nil {
		log.Printf("Response encode error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
