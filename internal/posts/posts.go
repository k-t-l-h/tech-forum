package posts

import (
	"encoding/json"
	"forum/internal/database"
	"forum/internal/models"
	"forum/internal/response"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	r *database.Repo
}

func NewHandler(r *database.Repo) *Handler {
	return &Handler{r: r}
}

//GET /post/{id}/details
func (h *Handler) PostDetails(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	ids := vars["id"]
	query := request.URL.Query()
	relateds := query["related"]
	related := []string{}

	if len(relateds) > 0 {
		related = strings.Split(relateds[0], ",")
	}
	id, _ := strconv.Atoi(ids)

	pu := models.PostFull{}
	bt, _ := ioutil.ReadAll(request.Body)
	json.Unmarshal(bt, &pu)

	pu.Post.Id = id
	res, status := h.r.GetPost(pu, related)

	switch status {
	case models.OK:
		response.Respond(writer, http.StatusOK, res)
	case models.NotFound:
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "User not found"})
	}
}

//POST /post/{id}/details
func (h *Handler) PostUpdateDetails(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	ids := vars["id"]

	pu := models.PostUpdate{}
	bt, _ := ioutil.ReadAll(request.Body)
	json.Unmarshal(bt, &pu)

	id, err := strconv.Atoi(ids)

	if err == nil {
		pu.Id = id
	}

	up, status := h.r.UpdatePost(pu)
	switch status {
	case models.OK:
		response.Respond(writer, http.StatusOK, up)
	default:
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "User not found"})

	}

}
