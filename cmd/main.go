package main

import (
	"context"
	"forum/internal/database"
	"forum/internal/forum"
	"forum/internal/posts"
	"forum/internal/service"
	"forum/internal/threads"
	"forum/internal/user"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"
)

func main() {
	//conn := "postgres://postgres:password@127.0.0.1:5432/db5"
	conn := "postgres://docker:docker@127.0.0.1:5432/docker?pool_max_conns=1000"

	pool, err := pgxpool.Connect(context.Background(), conn)
	if err != nil {
		log.Fatal(pool)
	}

	muxRouter := mux.NewRouter()

	r := database.NewRepo(pool)
	h := forum.NewHandler(r)
	h2 := posts.NewHandler(r)
	h3 := user.NewHandler(r)
	h4 := threads.NewHandler(r)
	h5 := service.NewHandler(r)

	muxRouter.HandleFunc("/api/forum/create", h.CreateForum).Methods("POST")
	muxRouter.HandleFunc("/api/forum/{slug}/create", h.CreateSlug).Methods("POST")
	muxRouter.HandleFunc("/api/forum/{slug}/details", h.SlugDetails).Methods("GET")
	muxRouter.HandleFunc("/api/forum/{slug}/threads", h.SlugThreads).Methods("GET")
	muxRouter.HandleFunc("/api/forum/{slug}/users", h.SlugUsers).Methods("GET")

	muxRouter.HandleFunc("/api/post/{id}/details", h2.PostDetails).Methods("GET")
	muxRouter.HandleFunc("/api/post/{id}/details", h2.PostUpdateDetails).Methods("POST")

	muxRouter.HandleFunc("/api/user/{nickname}/create", h3.Create).Methods("POST")
	muxRouter.HandleFunc("/api/user/{nickname}/profile", h3.Update).Methods("POST")
	muxRouter.HandleFunc("/api/user/{nickname}/profile", h3.Details).Methods("GET")

	muxRouter.HandleFunc("/api/service/status", h5.Status).Methods("GET")
	muxRouter.HandleFunc("/api/service/clear", h5.Clear).Methods("POST")

	muxRouter.HandleFunc("/api", h5.Index).Methods("GET")
	muxRouter.HandleFunc("/api/thread/{id:[0-9]+}/create", h4.CreateID).Methods("POST")
	muxRouter.HandleFunc("/api/thread/{slug_or_id}/create", h4.Create).Methods("POST")

	muxRouter.HandleFunc("/api/thread/{id:[0-9]+}/details", h4.UpdateID).Methods("POST")
	muxRouter.HandleFunc("/api/thread/{slug_or_id}/details", h4.Update).Methods("POST")

	muxRouter.HandleFunc("/api/thread/{id:[0-9]+}/vote", h4.VoteID).Methods("POST")
	muxRouter.HandleFunc("/api/thread/{slug_or_id}/vote", h4.Vote).Methods("POST")

	muxRouter.HandleFunc("/api/thread/{id:[0-9]+}/details", h4.DetailsID).Methods("GET")
	muxRouter.HandleFunc("/api/thread/{slug_or_id}/details", h4.Details).Methods("GET")

	muxRouter.HandleFunc("/api/thread/{id:[0-9]+}/posts", h4.PostsID).Methods("GET")
	muxRouter.HandleFunc("/api/thread/{slug_or_id}/posts", h4.Posts).Methods("GET")

	http.Handle("/", muxRouter)
	log.Print(http.ListenAndServe(":5000", muxRouter))

}
