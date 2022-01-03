package database

import (
	"context"
	"forum/internal/models"
	"github.com/jackc/pgconn"
	v4 "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

type Repo struct {
	db   *pgxpool.Pool
	info models.Status
}

func NewRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db, info: models.Status{}}
}

func (r *Repo) CreateForum(forum models.Forum) (models.Forum, int) {

	user, code := r.CheckUser(models.User{NickName: forum.User})
	if code != models.OK {
		return models.Forum{}, models.UserNotFound
	}

	forum.User = user.NickName
	//уменьшаем количество записей
	query := `INSERT INTO forums (title, author, slug) 
	VALUES ($1, $2, $3);`

	//уменьшаем количество выделений памяти
	var results models.Forum
	_, err := r.db.Exec(context.Background(),
		query, forum.Title, forum.User, forum.Slug)

	if err != nil {
		if pqError, ok := err.(*pgconn.PgError); ok {
			switch pqError.Code {
			case "23505":
				result, _ := r.GetForumBySlag(forum)
				return result, models.ForumConflict
			case "23503":
				return results, models.UserNotFound
			default:
				result, _ := r.GetForumBySlag(forum)
				return result, models.ForumConflict
			}
		}
	}

	r.info.Forum++
	return forum, models.OK
}

///forum/{slug}/create
func (r *Repo) CreateSlug(thread models.Thread) (models.Thread, int) {

	user, code := r.CheckUser(models.User{NickName: thread.Author})
	if code != models.OK {
		return models.Thread{}, models.UserNotFound
	}
	thread.Author = user.NickName

	f, status := r.ForumCheck(models.Forum{Slug: thread.Forum})
	if status == models.NotFound {
		return models.Thread{}, models.UserNotFound
	}

	thread.Forum = f.Slug

	t := thread

	if thread.Slug != "" {
		thread, status := r.CheckSlug(thread)
		if status == models.OK {
			th, _ := r.GetThreadBySlug(thread.Slug, t)
			return th, models.ForumConflict
		}
	}

	query := `INSERT INTO threads (author, message, title, created_at, forum, slug, votes)
				VALUES ($1, $2, $3, $4, $5, $6, $7)	RETURNING id`

	row := r.db.QueryRow(context.Background(), query, thread.Author, thread.Message, thread.Title,
		thread.CreatedAt, thread.Forum, thread.Slug, 0)

	err := row.Scan(&t.Id)

	if err != nil {
		if pqError, ok := err.(*pgconn.PgError); ok {
			switch pqError.Code {
			case "23505":
				return t, models.ForumConflict
			case "23503":
				return models.Thread{}, models.UserNotFound
			default:
				log.Print(pqError.Code)
				return models.Thread{}, models.UserNotFound
			}
		}
	}
	r.info.Thread++
	query2 := `INSERT INTO forum_users(nickname,
    forum) VALUES ($1, $2) ON CONFLICT DO NOTHING;`
	_, err = r.db.Exec(context.Background(), query2, thread.Author, thread.Forum)
	if err != nil {
		log.Fatal(err)
	}

	query3 := `UPDATE forums SET threads = threads + 1 WHERE slug =$1`
	_, err = r.db.Exec(context.Background(), query3, thread.Forum)
	if err != nil {
		log.Fatal(err)
	}
	return t, models.OK
}

///forum/{slug}/details
func (r *Repo) GetForumBySlag(forum models.Forum) (models.Forum, int) {
	query := `SELECT title, author, slug, posts, threads
				FROM forums 
				WHERE slug = $1;`

	row := r.db.QueryRow(context.Background(), query, forum.Slug)

	err := row.Scan(&forum.Title, &forum.User, &forum.Slug, &forum.Posts, &forum.Threads)

	if err != nil {
		return forum, models.NotFound
	}

	return forum, models.OK
}

func (r *Repo) ForumCheck(forum models.Forum) (models.Forum, int) {
	query := `SELECT slug FROM forums 
				WHERE slug = $1;`

	row := r.db.QueryRow(context.Background(), query, forum.Slug)

	err := row.Scan(&forum.Slug)

	if err != nil {
		return forum, models.NotFound
	}

	return forum, models.OK
}

///forum/{slug}/users
func (r *Repo) GetForumUsers(forum models.Forum, limit, since, desc string) ([]models.User, int) {
	us := []models.User{}
	var row v4.Rows
	var err error

	forum, state := r.ForumCheck(forum)
	if state != models.OK {
		return us, models.NotFound
	}

	query := ``

	if limit == "" && since == "" {
		if desc == "true" {
			query = `SELECT email,
                      fullname,
                      users.nickname,
                      about
		FROM users JOIN forum_users ON users.nickname = forum_users.nickname WHERE forum = $1
		ORDER BY forum_users.nickname DESC`
		} else {
			query = `SELECT email,
                      fullname,
                      users.nickname,
                      about
		FROM users JOIN forum_users ON users.nickname = forum_users.nickname WHERE forum = $1
		ORDER BY forum_users.nickname ASC`
		}

		row, err = r.db.Query(context.Background(), query, forum.Slug)
	}

	if limit != "" && since == "" {
		if desc == "true" {
			query = `SELECT email,
                      fullname,
                      users.nickname,
                      about
		FROM users JOIN forum_users ON users.nickname = forum_users.nickname WHERE forum = $1
		ORDER BY forum_users.nickname DESC LIMIT $2`
		} else {
			query = `SELECT email,
                      fullname,
                      users.nickname,
                      about
		FROM users JOIN forum_users ON users.nickname = forum_users.nickname WHERE forum = $1
		ORDER BY forum_users.nickname ASC LIMIT $2`
		}

		row, err = r.db.Query(context.Background(), query, forum.Slug, limit)
	}

	if limit == "" && since != "" {
		if desc == "true" {
			query = `SELECT email,
                      fullname,
                      users.nickname,
                      about
		FROM users JOIN forum_users ON users.nickname = forum_users.nickname WHERE forum = $1
		AND forum_users.nickname < $2
		ORDER BY forum_users.nickname  DESC  `
		} else {
			query = `SELECT email,
                      fullname,
                     users.nickname,
                      about
		FROM users JOIN forum_users ON users.nickname = forum_users.nickname WHERE forum = $1
		AND forum_users.nickname > $2
		ORDER BY forum_users.nickname  ASC`
		}

		row, err = r.db.Query(context.Background(), query, forum.Slug, since)
	}

	if limit != "" && since != "" {
		if desc == "true" {
			query = `SELECT email,
                      fullname,
                      users.nickname,
                      about
		FROM users JOIN forum_users ON users.nickname = forum_users.nickname WHERE forum = $1
		AND forum_users.nickname < $2
		ORDER BY forum_users.nickname  DESC  LIMIT $3`
		} else {
			query = `SELECT email,
                      fullname,
                      users.nickname,
                      about
		FROM users JOIN forum_users ON users.nickname = forum_users.nickname WHERE forum = $1
		AND forum_users.nickname > $2
		ORDER BY forum_users.nickname ASC LIMIT $3`
		}

		row, err = r.db.Query(context.Background(), query, forum.Slug, since, limit)
		log.Print(err)
	}

	defer row.Close()

	for row.Next() {
		a := models.User{}
		row.Scan(&a.Email, &a.FullName, &a.NickName, &a.About)
		us = append(us, a)
	}
	return us, models.OK
}

///forum/{slug}/threads
func (r *Repo) GetForumThreads(t models.Thread, limit, since, desc string) ([]models.Thread, int) {

	th := []models.Thread{}
	var row v4.Rows
	var err error

	query := ``

	if limit == "" && since == "" {
		if desc == "" || desc == "false" {
			query = `SELECT id, slug, author, created_at, forum, title, message, votes
						FROM threads
						WHERE forum = $1
						ORDER BY created_at ASC`

		} else {
			query = `SELECT id, slug, author, created_at, forum, title, message, votes
						FROM threads
						WHERE forum = $1
						ORDER BY created_at DESC`
		}
		row, err = r.db.Query(context.Background(), query, t.Forum)
	} else {

		if limit != "" && since == "" {
			if desc == "" || desc == "false" {
				query = `SELECT id, slug, author, created_at, forum, title, message, votes
						FROM threads
						WHERE forum = $1
						ORDER BY created_at ASC  LIMIT $2`

			} else {
				query = `SELECT id, slug, author, created_at, forum, title, message, votes
						FROM threads
						WHERE forum = $1
						ORDER BY created_at DESC  LIMIT $2`
			}

			row, err = r.db.Query(context.Background(), query, t.Forum, limit)
		}

		if since != "" && limit == "" {
			if desc == "" || desc == "false" {
				query = `SELECT id, slug, author, created_at, forum, title, message, votes
						FROM threads
						WHERE forum = $1 AND created_at >= $2
						ORDER BY created_at ASC `
			} else {
				query = `SELECT id, slug, author, created_at, forum, title, message, votes
						FROM threads
						WHERE forum = $1 AND created_at <= $2
						ORDER BY created_at DESC `
			}

			row, err = r.db.Query(context.Background(), query, t.Forum, since)
		}

		if since != "" && limit != "" {

			if desc == "" || desc == "false" {

				query = `SELECT id, slug, author, created_at, forum, title, message, votes
						FROM threads
						WHERE forum = $1 AND created_at >= $2
						ORDER BY created_at ASC LIMIT $3`
			} else {
				query = `SELECT id, slug, author, created_at, forum, title, message, votes
						FROM threads
						WHERE forum = $1 AND created_at <= $2
						ORDER BY created_at DESC LIMIT $3`
			}
			row, err = r.db.Query(context.Background(), query, t.Forum, since, limit)
		}
	}
	defer row.Close()
	for row.Next() {
		t := models.Thread{}
		err = row.Scan(&t.Id, &t.Slug, &t.Author, &t.CreatedAt, &t.Forum, &t.Title, &t.Message, &t.Votes)

		th = append(th, t)
	}
	if err == nil {

	}

	if len(th) == 0 {
		_, status := r.GetForumBySlag(models.Forum{Slug: t.Forum})
		if status != models.OK {
			return th, models.NotFound
		}
		return th, models.OK
	}

	return th, models.OK
}

func (r *Repo) CheckSlug(thread models.Thread) (models.Thread, int) {
	query := `SELECT slug, author
				FROM threads 
				WHERE slug = $1;`

	row := r.db.QueryRow(context.Background(), query, thread.Slug)

	err := row.Scan(&thread.Slug, &thread.Author)

	if err != nil {
		return thread, models.NotFound
	}

	return thread, models.OK
}

func (r *Repo) GetThreadBySlug(check string, thread models.Thread) (models.Thread, int) {
	query := `SELECT id, author, message, title, created_at, forum, slug, votes
					FROM threads
					WHERE slug = $1`
	row := r.db.QueryRow(context.Background(), query, check)

	err := row.Scan(&thread.Id, &thread.Author, &thread.Message, &thread.Title,
		&thread.CreatedAt, &thread.Forum, &thread.Slug, &thread.Votes)

	if err != nil {
		return thread, models.NotFound
	}

	return thread, models.OK
}
