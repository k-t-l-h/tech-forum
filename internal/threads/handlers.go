package threads

import (
	"encoding/json"
	"forum/internal/database"
	"forum/internal/models"
	"forum/internal/response"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Handler struct {
	r *database.Repo
}

func NewHandler(r *database.Repo) *Handler {
	return &Handler{r: r}
}

//POST // /thread/{slug_or_id}/create
func (h *Handler) Create(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["slug_or_id"]

	posts := []models.Post{}
	decoder := json.NewDecoder(request.Body)
	decoder.Decode(&posts)

	posts, status := h.r.CreateThreadPost(slug, posts)

	switch status {
	case models.OK:
		response.Respond(writer, http.StatusCreated, posts)
	case models.NotFound:
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "User not found"})
	case models.ForumConflict:
		response.Respond(writer, http.StatusConflict, models.Error{Message: "User not found"})
	}
}

func (h *Handler) Update(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["slug_or_id"]

	var t models.Thread

	bt, _ := ioutil.ReadAll(request.Body)
	json.Unmarshal(bt, &t)

	tr, status := h.r.ThreadUpdate(slug, t)

	switch status {
	case models.OK:
		response.Respond(writer, http.StatusOK, tr)
	case models.NotFound:
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "User not found"})
	}

}

func (h *Handler) UpdateID(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["id"]
	id, _ := strconv.Atoi(slug)

	var t models.Thread
	bt, _ := ioutil.ReadAll(request.Body)
	json.Unmarshal(bt, &t)

	tr, status := h.r.ThreadUpdateID(id, t)

	switch status {
	case models.OK:
		response.Respond(writer, http.StatusOK, tr)
	case models.NotFound:
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "User not found"})
	}

}

///thread/{slug_or_id}/vote
func (h *Handler) Vote(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["slug_or_id"]
	vote := models.Vote{}

	bt, _ := ioutil.ReadAll(request.Body)
	json.Unmarshal(bt, &vote)
	thread, status := h.r.ThreadVote(slug, vote)

	switch status {
	case models.OK:
		response.Respond(writer, http.StatusOK, thread)
	case models.NotFound:
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "User not found"})

	}
}

///thread/{slug_or_id}/vote
func (h *Handler) VoteID(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["id"]
	vote := models.Vote{}

	id, _ := strconv.Atoi(slug)

	bt, _ := ioutil.ReadAll(request.Body)
	json.Unmarshal(bt, &vote)
	thread, status := h.r.ThreadVoteID(id, vote)

	switch status {
	case models.OK:
		response.Respond(writer, http.StatusOK, thread)

	case models.NotFound:
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "User not found"})

	}
}

//GET

func (h *Handler) Posts(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["slug_or_id"]

	query := request.URL.Query()

	limits := query["limit"]
	sinces := query["since"]
	descs := query["desc"]
	sorts := query["sort"]

	limit := ""
	since := ""
	desc := ""
	sort := ""

	if len(limits) > 0 {
		limit = limits[0]
	}

	if len(sinces) > 0 {
		since = sinces[0]
	}

	if len(descs) > 0 {
		desc = descs[0]
	}
	if len(sorts) > 0 {
		sort = sorts[0]
	}

	Ps, status := h.r.GetThreadsPosts(limit, since, desc, sort, slug)
	switch status {
	case models.OK:
		//успешно
		response.Respond(writer, http.StatusOK, Ps)
	case models.NotFound:
		//нет ветки
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "Thread not found"})
	}

}

///thread/{slug_or_id}/details
func (h *Handler) Details(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["slug_or_id"]

	var t models.Thread
	bt, _ := ioutil.ReadAll(request.Body)
	json.Unmarshal(bt, &t)
	t.Slug = slug

	thread, status := h.r.GetThreadBySlug(slug, t)

	switch status {
	case models.OK:
		//успешно
		response.Respond(writer, http.StatusOK, thread)
	case models.NotFound:
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "Thread not found"})
	}

}

///thread/{slug_or_id}/details
func (h *Handler) DetailsID(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["id"]
	id, _ := strconv.Atoi(slug)

	var t models.Thread
	bt, _ := ioutil.ReadAll(request.Body)
	json.Unmarshal(bt, &t)
	t.Id = id

	thread, status := h.r.GetThreadByID(id, t)

	switch status {
	case models.OK:
		//успешно
		response.Respond(writer, http.StatusOK, thread)
	case models.NotFound:
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "Thread not found"})
	}

}

func (h *Handler) CreateID(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["id"]
	id, _ := strconv.Atoi(slug)

	posts := []models.Post{}
	decoder := json.NewDecoder(request.Body)
	decoder.Decode(&posts)

	posts, status := h.r.CreateThreadPostID(id, posts)

	switch status {
	case models.OK:
		response.Respond(writer, http.StatusCreated, posts)
	case models.NotFound:
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "Forum not found"})
	case models.ForumConflict:
		response.Respond(writer, http.StatusConflict, models.Error{Message: "Thread not found"})
	}
}

func (h *Handler) PostsID(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["id"]
	id, _ := strconv.Atoi(slug)

	query := request.URL.Query()

	limits := query["limit"]
	sinces := query["since"]
	descs := query["desc"]
	sorts := query["sort"]

	limit := ""
	since := ""
	desc := ""
	sort := ""

	if len(limits) > 0 {
		limit = limits[0]
	}

	if len(sinces) > 0 {
		since = sinces[0]
	}

	if len(descs) > 0 {
		desc = descs[0]
	}
	if len(sorts) > 0 {
		sort = sorts[0]
	}

	Ps, status := h.r.GetThreadsPostsID(limit, since, desc, sort, id)
	switch status {
	case models.OK:
		//успешно
		response.Respond(writer, http.StatusOK, Ps)
	case models.NotFound:
		//нет ветки
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "Thread not found"})
	}

}
