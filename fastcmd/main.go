package main

import (
	"context"
	"forum/internal/database"
	"forum/internal/forum"
	"forum/internal/posts"
	"forum/internal/service"
	"forum/internal/threads"
	"forum/internal/user"
	"github.com/jackc/pgx/v4/pgxpool"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"log"
)

type fastService struct {
	Port   string
	Router *routing.Router
}

func main() {

	//conn := "postgres://postgres:password@127.0.0.1:5432/db4?pool_max_conns=1000"
	conn := "postgres://docker:docker@127.0.0.1:5432/docker?pool_max_conns=1000"

	pool, err := pgxpool.Connect(context.Background(), conn)
	if err != nil {
		log.Fatal(pool)
	}

	r := database.NewRepo(pool)
	h := forum.NewFastHandler(r)
	h2 := posts.NewFastHandler(r)
	h3 := threads.NewFastHandler(r)
	h4 := user.NewFastHandler(r)
	h5 := service.NewFastHandler(r)

	router := routing.New()
	router.To("POST", "/api/forum/create", h.CreateForum)
	router.To("POST", "/api/forum/<slug>/create", h.CreateSlug)
	router.To("GET", "/api/forum/<slug>/details", h.SlugDetails)
	router.To("GET", "/api/forum/<slug>/threads", h.SlugThreads)
	router.To("GET", "/api/forum/<slug>/users", h.SlugUsers)

	router.To("POST", "/api/post/<id>/details", h2.PostUpdateDetails)
	router.To("GET", "/api/post/<id>/details", h2.PostDetails)

	router.To("GET", "/api/service/status", h5.Status)
	router.To("POST", "/api/service/clear", h5.Clear)
	router.To("GET", "/api", h5.Index)

	router.To("POST", "/api/thread/<slug_or_id>/create", h3.Create)
	router.To("POST", "/api/thread/<slug_or_id>/details", h3.Update)
	router.To("POST", "/api/thread/<slug_or_id>/vote", h3.Vote)

	router.To("GET", "/api/thread/<slug_or_id>/details", h3.Details)
	router.To("GET", "/api/thread/<slug_or_id>/posts", h3.Posts)

	router.To("POST", "/api/user/<nickname>/create", h4.Create)
	router.To("POST", "/api/user/<nickname>/profile", h4.Update)
	router.To("GET", "/api/user/<nickname>/profile", h4.Details)

	s := fastService{
		Port:   ":5000",
		Router: router,
	}
	log.Printf("Server running at %v\n", s.Port)
	fasthttp.ListenAndServe(s.Port, s.Router.HandleRequest)

}
