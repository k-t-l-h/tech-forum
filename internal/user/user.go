package user

import (
	"encoding/json"
	"forum/internal/database"
	"forum/internal/models"
	"forum/internal/response"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

type Handler struct {
	r *database.Repo
}

func NewHandler(r *database.Repo) *Handler {
	return &Handler{r: r}
}

func (h Handler) Create(writer http.ResponseWriter, request *http.Request) {

	vars := mux.Vars(request)
	name := vars["nickname"]

	var u models.User
	err := json.NewDecoder(request.Body).Decode(&u)
	u.NickName = name

	if err != nil {
		response.Respond(writer, http.StatusBadRequest, nil)
	}
	user, status := h.r.CreateUser(u)

	switch status {
	case models.OK:
		response.Respond(writer, http.StatusCreated, user[0])
	case models.ForumConflict:
		response.Respond(writer, http.StatusConflict, user)

	}

}

func (h Handler) Update(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	name := vars["nickname"]

	var p models.User

	bt, _ := ioutil.ReadAll(request.Body)
	json.Unmarshal(bt, &p)

	p.NickName = name

	u, status := h.r.UpdateUser(p)

	switch status {
	case models.OK:
		response.Respond(writer, http.StatusOK, u)
	case models.NotFound:
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "User not found"})
	case models.ForumConflict:
		response.Respond(writer, http.StatusConflict, models.Error{Message: "User not found"})
	}
}

//GET /user/{nickname}/profile
func (h Handler) Details(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	name := vars["nickname"]

	us := models.User{}
	us.NickName = name

	u, status := h.r.GetUser(us)

	switch status {
	case models.OK:
		response.Respond(writer, http.StatusOK, u)
	case models.NotFound:
		response.Respond(writer, http.StatusNotFound, models.Error{Message: "User not found"})
	}
}
