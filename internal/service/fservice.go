package service

import (
	"encoding/json"
	"forum/internal/database"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

type FastHandler struct {
	r *database.Repo
}

func NewFastHandler(r *database.Repo) *FastHandler {
	return &FastHandler{r: r}
}

func (fh FastHandler) Clear(ctx *routing.Context) error {

	fh.r.Clear()
	ctx.SetStatusCode(fasthttp.StatusOK)
	return nil
}

// /service/status
func (fh FastHandler) Status(ctx *routing.Context) error {
	info := fh.r.Info()
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetContentType("application/json")
	data, _ := json.Marshal(info)
	ctx.SetBody(data)
	return nil
}

func (fh FastHandler) Index(ctx *routing.Context) error {
	ctx.SetStatusCode(fasthttp.StatusOK)
	return nil
}
