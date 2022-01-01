package forum

import (
	"forum/internal/database"
	"forum/internal/models"
	"forum/internal/response"
	"github.com/gorilla/mux"
	"github.com/mailru/easyjson"
	"log"
	"net/http"
)

type Handler struct {
	r *database.Repo
}

func NewHandler(r *database.Repo) *Handler {
	return &Handler{r: r}
}

func (h *Handler) CreateForum(writer http.ResponseWriter, request *http.Request) {
	var f models.Forum

	jsonerr := easyjson.UnmarshalFromReader(request.Body, &f)

	if jsonerr != nil {
		log.Fatalln(jsonerr)
	}

	forums, err := h.r.CreateForum(f)

	switch err {
	case models.OK:
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		easyjson.MarshalToHTTPResponseWriter(forums, writer)
	case models.UserNotFound:
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		easyjson.MarshalToHTTPResponseWriter(models.Error{Message: "User not found"}, writer)

	case models.ForumConflict:
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusConflict)
		easyjson.MarshalToHTTPResponseWriter(forums, writer)
	}
}

func (h *Handler) CreateSlug(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["slug"]
	var t models.Thread
	t.Forum = slug

	jsonerr := easyjson.UnmarshalFromReader(request.Body, &t)

	if jsonerr != nil {
		panic(jsonerr)
	}

	th, err := h.r.CreateSlug(t)

	switch err {
	case models.OK:
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		easyjson.MarshalToHTTPResponseWriter(th, writer)

	case models.UserNotFound:
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		easyjson.MarshalToHTTPResponseWriter(models.Error{Message: "User not found"}, writer)

	case models.ForumConflict:
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusConflict)
		easyjson.MarshalToHTTPResponseWriter(th, writer)

	default:
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusTeapot)
		easyjson.MarshalToHTTPResponseWriter(th, nil)
	}

}

func (h *Handler) SlugDetails(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["slug"]

	f := models.Forum{}
	f.Slug = slug

	f, err := h.r.GetForumBySlag(f)

	switch err {
	case models.OK:
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		easyjson.MarshalToHTTPResponseWriter(f, writer)

	case models.NotFound:
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		easyjson.MarshalToHTTPResponseWriter(models.Error{Message: "User not found"}, writer)
	}

}

func (h *Handler) SlugThreads(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["slug"]

	query := request.URL.Query()
	limits := query["limit"]
	sinces := query["since"]
	descs := query["desc"]

	limit := ""
	since := ""
	desc := ""

	var f models.Thread
	f.Forum = slug

	if len(limits) > 0 {
		limit = limits[0]
	}

	if len(sinces) > 0 {
		since = sinces[0]
	}

	if len(descs) > 0 {
		desc = descs[0]
	}

	t, status := h.r.GetForumThreads(f, limit, since, desc)

	switch status {
	case models.OK:
		response.Respond(writer, http.StatusOK, t)
	case models.NotFound:
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		easyjson.MarshalToHTTPResponseWriter(models.Error{Message: "User not found"}, writer)
	}
}

func (h *Handler) SlugUsers(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["slug"]

	query := request.URL.Query()
	limits := query["limit"]
	sinces := query["since"]
	descs := query["desc"]

	limit := ""
	since := ""
	desc := ""

	if len(limits) > 0 {
		limit = limits[0]
	}

	if len(sinces) > 0 {
		since = sinces[0]
	}

	if len(descs) > 0 {
		desc = descs[0]
	}

	var f models.Forum
	f.Slug = slug

	u, status := h.r.GetForumUsers(f, limit, since, desc)

	switch status {
	case models.OK:
		response.Respond(writer, http.StatusOK, u)
	case models.NotFound:
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "Forum not found"})
	}

}
