package forum

import (
	"forum/internal/database"
	"forum/internal/models"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"strings"
)

type FastHandler struct {
	r *database.Repo
}

func NewFastHandler(r *database.Repo) *FastHandler {
	return &FastHandler{r: r}
}

func (fh FastHandler) CreateForum(ctx *routing.Context) error {

	var f models.Forum
	json.Unmarshal(ctx.PostBody(), &f)
	forums, err := fh.r.CreateForum(f)

	switch err {
	case models.OK:
		ctx.SetStatusCode(fasthttp.StatusCreated)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(forums)
		ctx.SetBody(data)

	case models.UserNotFound:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(models.Error{Message: "User not found"})
		ctx.SetBody(data)

	case models.ForumConflict:
		ctx.SetStatusCode(fasthttp.StatusConflict)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(forums)
		ctx.SetBody(data)

	}
	return nil
}

func (fh FastHandler) CreateSlug(ctx *routing.Context) error {

	slug := ctx.Param("slug")

	var t models.Thread
	t.Forum = slug
	json.Unmarshal(ctx.PostBody(), &t)

	th, err := fh.r.CreateSlug(t)

	switch err {
	case models.OK:
		ctx.SetStatusCode(fasthttp.StatusCreated)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(th)
		ctx.SetBody(data)

	case models.UserNotFound:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(models.Error{Message: "User not found"})
		ctx.SetBody(data)

	case models.ForumConflict:
		ctx.SetStatusCode(fasthttp.StatusConflict)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(th)
		ctx.SetBody(data)
	}
	return nil
}

func (fh FastHandler) SlugDetails(ctx *routing.Context) error {
	slug := ctx.Param("slug")
	f := models.Forum{}
	f.Slug = slug

	f, err := fh.r.GetForumBySlag(f)

	switch err {
	case models.OK:
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(f)
		ctx.SetBody(data)

	case models.NotFound:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(models.Error{Message: "User not found"})
		ctx.SetBody(data)
	}
	return nil
}

func (fh FastHandler) SlugThreads(ctx *routing.Context) error {

	slug := ctx.Param("slug")

	limits := strings.Split(string(ctx.QueryArgs().Peek("limit")), ",)")
	sinces := strings.Split(string(ctx.QueryArgs().Peek("since")), ",)")
	descs := strings.Split(string(ctx.QueryArgs().Peek("desc")), ",)")

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

	t, status := fh.r.GetForumThreads(f, limit, since, desc)

	switch status {
	case models.OK:
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(t)
		ctx.SetBody(data)
	case models.NotFound:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(models.Error{Message: "User not found"})
		ctx.SetBody(data)
	}
	return nil
}

func (fh FastHandler) SlugUsers(ctx *routing.Context) error {
	slug := ctx.Param("slug")

	limits := strings.Split(string(ctx.QueryArgs().Peek("limit")), ",)")
	sinces := strings.Split(string(ctx.QueryArgs().Peek("since")), ",)")
	descs := strings.Split(string(ctx.QueryArgs().Peek("desc")), ",)")

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

	u, status := fh.r.GetForumUsers(f, limit, since, desc)

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
	}
	return nil

}
