package forum

import (
	"forum/internal/database"
	"forum/internal/models"
	"forum/internal/response"
	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"log"
	"net/http"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Handler struct {
	r *database.Repo
}

func NewHandler(r *database.Repo) *Handler {
	return &Handler{r: r}
}

func (h *Handler) CreateForum(writer http.ResponseWriter, request *http.Request) {
	var f models.Forum

	bt, _ := ioutil.ReadAll(request.Body)
	jsonerr := json.Unmarshal(bt, &f)

	if jsonerr != nil {
		log.Fatalln(jsonerr)
	}

	forums, err := h.r.CreateForum(f)

	switch err {
	case models.OK:
		response.Respond(writer, http.StatusCreated, forums)
	case models.UserNotFound:
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "User not found"})
	case models.ForumConflict:
		response.Respond(writer, http.StatusConflict, forums)
	}
}

func (h *Handler) CreateSlug(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["slug"]
	var t models.Thread
	t.Forum = slug

	bt, _ := ioutil.ReadAll(request.Body)
	jsonerr := json.Unmarshal(bt, &t)
	if jsonerr != nil {
		panic(jsonerr)
	}

	th, err := h.r.CreateSlug(t)

	switch err {
	case models.OK:

		response.Respond(writer, http.StatusCreated, th)

	case models.UserNotFound:
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "User not found"})

	case models.ForumConflict:
		response.Respond(writer, http.StatusConflict, th)

	default:
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusTeapot)
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
		response.Respond(writer, http.StatusOK, f)
	case models.NotFound:
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "User not found"})
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
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "User not found"})
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
