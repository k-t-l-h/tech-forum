package service

import (
	"forum/internal/database"
	"forum/internal/response"
	"net/http"
)

type Handler struct {
	r *database.Repo
}

func NewHandler(r *database.Repo) *Handler {
	return &Handler{r: r}
}

// /service/clear
func (h *Handler) Clear(writer http.ResponseWriter, request *http.Request) {
	h.r.Clear()
	response.Respond(writer, http.StatusOK, nil)
}

// /service/status
func (h *Handler) Status(writer http.ResponseWriter, request *http.Request) {
	response.Respond(writer, http.StatusOK, h.r.Info())
}

func (h *Handler) Index(writer http.ResponseWriter, request *http.Request) {
	response.Respond(writer, http.StatusOK, nil)
}
