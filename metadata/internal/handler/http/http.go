package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/tj330/bookapp/metadata/internal/controller/metadata"
	"github.com/tj330/bookapp/metadata/internal/repository"
)

type Handler struct {
	ctrl *metadata.Controller
}

func New(ctrl *metadata.Controller) *Handler {
	return &Handler{ctrl}
}

func (h *Handler) GetMetadata(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	m, err := h.ctrl.Get(ctx, id)

	if err != nil && errors.Is(repository.ErrNotFound, err) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("repository got error for book %s: %v\n", id, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(m); err != nil {
		log.Printf("Response encode error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
