package user

import (
	"encoding/json"
	"forum/internal/database"
	"forum/internal/models"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"log"
)

type FastHandler struct {
	r *database.Repo
}

func NewFastHandler(r *database.Repo) *FastHandler {
	return &FastHandler{r: r}
}

func (fh FastHandler) Create(ctx *routing.Context) error{

	name := ctx.Param("nickname")

	var u models.User
	json.Unmarshal(ctx.PostBody(), &u)
	u.NickName = name

	user, status := fh.r.CreateUser(u)



	switch status {
	case models.OK:
		ctx.SetStatusCode(fasthttp.StatusCreated)
		ctx.SetContentType("application/json")
		m := models.User{user[0].About, user[0].Email, user[0].FullName, user[0].NickName}
		data, err := json.Marshal(m)

		ctx.SetBody(data)
		log.Print(err)

	case models.ForumConflict:
		ctx.SetStatusCode(fasthttp.StatusConflict)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(user)
		ctx.SetBody(data)

	}

	return nil
}

func (fh FastHandler) Update(ctx *routing.Context) error{
	name := ctx.Param("nickname")

	var p models.User
	json.Unmarshal(ctx.PostBody(), &p)
	p.NickName = name

	u, status := fh.r.UpdateUser(p)

	switch status {
	case models.OK:
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(u)
		ctx.SetBody(data)
	case models.NotFound:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(models.Error{Message: "User not found"})
		ctx.SetBody(data)
	case models.ForumConflict:
		ctx.SetStatusCode(fasthttp.StatusConflict)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(models.Error{Message: "StatusConflict"})
		ctx.SetBody(data)
	}
	return nil
}

//GET /user/{nickname}/profile
func (fh FastHandler) Details(ctx *routing.Context)error {
	name := ctx.Param("nickname")

	us := models.User{}
	us.NickName = name

	u, status := fh.r.GetUser(us)

	switch status {
	case models.OK:
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(u)
		ctx.SetBody(data)
	case models.NotFound:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(models.Error{Message: "User not found"})
		ctx.SetBody(data)
	case models.ForumConflict:
		ctx.SetStatusCode(fasthttp.StatusConflict)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(models.Error{Message: "StatusConflict"})
		ctx.SetBody(data)
	}
	return nil
}

