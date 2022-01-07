package posts

import (
	"encoding/json"
	"forum/internal/database"
	"forum/internal/models"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"strconv"
	"strings"
)

type FastHandler struct {
	r *database.Repo
}

func NewFastHandler(r *database.Repo) *FastHandler {
	return &FastHandler{r: r}
}

//GET /post/{id}/details
func (fh FastHandler) PostDetails(ctx *routing.Context) error {
	ids := ctx.Param("id")
	id, _ := strconv.Atoi(ids)

	relateds := strings.Split(string(ctx.QueryArgs().Peek("related")), ",)")

	related := []string{}

	if len(relateds) > 0 {
		related = strings.Split(relateds[0], ",")
	}

	pu := models.PostFull{}

	json.Unmarshal(ctx.PostBody(), &pu)

	pu.Post.Id = id
	res, status := fh.r.GetPost(pu, related)

	switch status {
	case models.OK:
		ctx.SetStatusCode(fasthttp.StatusOK)

		ctx.SetContentType("application/json")
		data, _ := json.Marshal(res)
		ctx.SetBody(data)

	case models.NotFound:
		ctx.SetStatusCode(fasthttp.StatusNotFound)

		ctx.SetContentType("application/json")
		data, _ := json.Marshal(models.Error{Message: "Thread not in forum"})
		ctx.SetBody(data)
	}
	return nil
}

//POST /post/{id}/details
func (fh FastHandler) PostUpdateDetails(ctx *routing.Context) error {

	ids := ctx.Param("id")

	pu := models.PostUpdate{}
	json.Unmarshal(ctx.PostBody(), &pu)
	id, err := strconv.Atoi(ids)

	if err == nil {
		pu.Id = id
	}

	up, status := fh.r.UpdatePost(pu)
	switch status {
	case models.OK:
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(up)
		ctx.SetBody(data)
	default:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(models.Error{Message: "Something went wrong"})
		ctx.SetBody(data)

	}

	return nil
}
