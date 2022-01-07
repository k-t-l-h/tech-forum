package threads

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



//POST // /thread/{slug_or_id}/create
func (fh FastHandler) Create(ctx *routing.Context) error{

	slug := ctx.Param("slug_or_id")
	id, err := strconv.Atoi(slug)

	posts := []models.Post{}
	status := 0
	json.Unmarshal(ctx.PostBody(), &posts)

	if err == nil {
		posts, status = fh.r.CreateThreadPostID(id, posts)
	} else {
		posts, status = fh.r.CreateThreadPost(slug, posts)
	}


	switch status {
	case models.OK:
		ctx.SetStatusCode(fasthttp.StatusCreated)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(posts)
		ctx.SetBody(data)

	case models.NotFound:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(models.Error{Message: "User not found"})
		ctx.SetBody(data)
	case models.ForumConflict:
		ctx.SetContentType("application/json")
		ctx.SetStatusCode(fasthttp.StatusConflict)
		data, _ := json.Marshal(models.Error{Message: "StatusConflict"})
		ctx.SetBody(data)
	}
	return nil
}

func (fh FastHandler) Update(ctx *routing.Context) error{
	slug := ctx.Param("slug_or_id")
	id, err := strconv.Atoi(slug)
	status := 0


	var t, tr models.Thread
	json.Unmarshal(ctx.PostBody(), &t)

	if err == nil {
		tr, status = fh.r.ThreadUpdateID(id, t)
	}else {
		tr, status = fh.r.ThreadUpdate(slug, t)
	}

	switch status {
	case models.OK:
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(tr)
		ctx.SetBody(data)
	case models.NotFound:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(models.Error{Message: "User not found"})
		ctx.SetBody(data)
	}
	return nil

}

///thread/{slug_or_id}/vote
func (fh FastHandler) Vote(ctx *routing.Context) error{
	slug := ctx.Param("slug_or_id")
	id, err := strconv.Atoi(slug)
	status := 0



	vote := models.Vote{}
	thread := models.Thread{}
	json.Unmarshal(ctx.PostBody(), &vote)
	if err == nil {
		thread, status = fh.r.ThreadVoteID(id, vote)
	} else {
		thread, status = fh.r.ThreadVote(slug, vote)
	}


	switch status {
	case models.OK:
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(thread)
		ctx.SetBody(data)
	case models.NotFound:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(models.Error{Message: "User not found"})
		ctx.SetBody(data)

	}
	return nil
}

///thread/{slug_or_id}/vote

func (fh FastHandler) Posts(ctx *routing.Context) error{

	Ps := []models.Post{}
	slug := ctx.Param("slug_or_id")
	id, err := strconv.Atoi(slug)
	status := 0

	limits := strings.Split(string(ctx.QueryArgs().Peek("limit")), ",)")
	sinces := strings.Split(string(ctx.QueryArgs().Peek("since")), ",)")
	descs := strings.Split(string(ctx.QueryArgs().Peek("desc")), ",)")
	sorts := strings.Split(string(ctx.QueryArgs().Peek("sort")), ",)")

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


	if err == nil {
		Ps, status = fh.r.GetThreadsPostsID(limit, since, desc, sort, id)
	} else {
		Ps, status = fh.r.GetThreadsPosts(limit, since, desc, sort, slug)
	}

	switch status {
	case models.OK:
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(Ps)
		ctx.SetBody(data)
	case models.NotFound:
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetContentType("application/json")
		data, _ := json.Marshal(models.Error{Message: "User not found"})
		ctx.SetBody(data)
	}

	return nil
}

///thread/{slug_or_id}/details
func (fh FastHandler) Details(ctx *routing.Context) error{
	slug := ctx.Param("slug_or_id")
	id, err := strconv.Atoi(slug)
	status := 0

	var t models.Thread
	json.Unmarshal(ctx.PostBody(), &t)
	t.Slug = slug

	if err == nil {
		t, status = fh.r.GetThreadByID(id, t)
	} else {
		t, status = fh.r.GetThreadBySlug(slug, t)
	}

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

