package database

import (
	"context"
	"forum/internal/models"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx"
	v4 "github.com/jackc/pgx/v4"
	"log"
	"strconv"
	"time"
)

// /thread/{slug_or_id}/create
func (r *Repo) CreateThreadPost(check string, posts []models.Post) ([]models.Post, int) {

	thread := models.Thread{}

	query := `SELECT id, forum
					FROM threads
					WHERE slug = $1`

	row := r.db.QueryRow(context.Background(), query, check)
	err := row.Scan(&thread.Id, &thread.Forum)

	if err != nil {
		return posts, models.NotFound
	}

	times := time.Now()

	if len(posts) == 0 {
		return posts, models.OK
	}

	tx, err := r.db.Begin(context.Background())
	query = `INSERT INTO posts (author, post, created_at, forum,  isEdited, parent, thread, path) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
			RETURNING  id;`
	ins, _ := tx.Prepare(context.Background(), "insert", query)

	result := []models.Post{}
	for _, p := range posts {
		p.Forum = thread.Forum
		p.Thread = thread.Id

		if p.Parent != 0 {
			old := 0
			query2 := `SELECT thread FROM posts WHERE id = $1`
			row = tx.QueryRow(context.Background(), query2, p.Parent)
			err := row.Scan(&old)
			if err != nil || old != p.Thread {
				return []models.Post{}, models.ForumConflict
			}
		}

		err = tx.QueryRow(context.Background(), ins.Name, p.Author, p.Message, times, thread.Forum, false, p.Parent, thread.Id, []int{}).Scan(&p.Id)

		p.CreatedAt = times.Format(time.RFC3339)

		if err != nil {
			tx.Rollback(context.Background())
			if pqError, ok := err.(pgx.PgError); ok {
				switch pqError.Code {
				case "23503":
					return []models.Post{}, models.NotFound
				case "23505":
					return []models.Post{}, models.ForumConflict
				case "22000":
					return []models.Post{}, models.NotFound
				default:
					return []models.Post{}, models.ForumConflict
				}
			}
		}

		query2 := `INSERT INTO forum_users(nickname,
    	forum) VALUES ($1, $2) ON CONFLICT DO NOTHING;`
		_, err = r.db.Exec(context.Background(), query2, p.Author, p.Forum)
		if err != nil {
			return []models.Post{}, models.NotFound
		}
		result = append(result, p)
		r.info.Post++
	}
	tx.Commit(context.Background())
	query3 := `UPDATE forums SET posts = posts + $2 WHERE slug =$1`
	_, err = r.db.Exec(context.Background(), query3, thread.Forum, len(result))
	if err != nil {
		log.Fatal(err)
	}
	return result, models.OK
}

func (r *Repo) GetThreadBySlugOrId(check string, thread models.Thread) (models.Thread, int) {
	var row v4.Row

	if value, err := strconv.Atoi(check); err != nil {
		thread.Slug = check
		query := `SELECT id, author, message, title, created_at, forum, slug, votes
					FROM threads
					WHERE slug = $1`
		row = r.db.QueryRow(context.Background(), query, thread.Slug)

	} else {
		query := `SELECT id, author, message, title, created_at, forum, slug, votes
					FROM threads
					WHERE id = $1`
		row = r.db.QueryRow(context.Background(), query, value)
	}

	err := row.Scan(&thread.Id, &thread.Author, &thread.Message, &thread.Title,
		&thread.CreatedAt, &thread.Forum, &thread.Slug, &thread.Votes)

	if err != nil {
		return thread, models.NotFound
	}

	return thread, models.OK
}

// thread/{slug_or_id}/posts
func (r *Repo) GetThreadsPosts(limit, since, desc, sort, check string) (models.Posts, int) {

	var row v4.Rows
	ps := models.Posts{}
	//TODO: получить только id
	thread, status := r.GetThreadBySlug(check, models.Thread{})
	if status == models.NotFound {
		return ps, models.NotFound
	}

	switch sort {
	case "flat":
		row = r.getFlat(thread.Id, since, limit, desc)

	case "tree":
		row = r.getTree(thread.Id, since, limit, desc)

	case "parent_tree":
		row = r.getParentTree(thread.Id, since, limit, desc)

	default:
		row = r.getFlat(thread.Id, since, limit, desc)
	}

	defer row.Close()
	for row.Next() {

		pr := models.Post{}
		times := time.Time{}
		err := row.Scan(&pr.Id, &pr.Author, &pr.Message, &times, &pr.Forum, &pr.IsEdited, &pr.Parent)
		pr.Thread = thread.Id
		pr.CreatedAt = times.Format(time.RFC3339)
		if err != nil {
		}
		ps = append(ps, pr)
	}

	return ps, models.OK
}

///thread/{slug_or_id}/vote
func (r *Repo) ThreadVote(check string, vote models.Vote) (models.Thread, int) {

	user, code := r.CheckUser(models.User{NickName: vote.NickName})
	if code != models.OK {
		return models.Thread{}, models.NotFound
	}
	vote.NickName = user.NickName

	thread, status := r.GetSlugID(check, models.Thread{})
	if status == models.NotFound {
		return thread, models.NotFound
	}

	query := `INSERT INTO VOTES (author, vote, thread) VALUES ($1, $2, $3) RETURNING vote`
	row := r.db.QueryRow(context.Background(), query, vote.NickName, vote.Voice, thread.Id)

	value := 0

	err := row.Scan(&vote.Voice)
	if err != nil {
		if pqError, ok := err.(*pgconn.PgError); ok {
			switch pqError.Code {
			case "23503":
				return thread, models.NotFound
			case "23505":
				upd := "WITH u AS ( SELECT vote FROM votes WHERE author = $2 AND thread = $3)" +
					"UPDATE votes SET vote =  $1 WHERE author = $2 AND thread = $3 " +
					"RETURNING vote, (SELECT vote FROM u)"
				row := r.db.QueryRow(context.Background(), upd, vote.Voice, vote.NickName, thread.Id)
				err := row.Scan(&vote.Voice, &value)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
	query = `UPDATE threads SET votes=votes+$1 WHERE id = $2;`
	_, err = r.db.Exec(context.Background(), query, vote.Voice-value, thread.Id)

	thread, status = r.GetThreadByID(thread.Id, models.Thread{})
	return thread, models.OK
}

// /thread/{slug_or_id}/details
func (r *Repo) ThreadUpdate(check string, thread models.Thread) (models.Thread, int) {
	t, status := r.GetThreadBySlug(check, thread)
	if status == models.NotFound {
		return thread, models.NotFound
	}

	if thread.Message != "" {
		t.Message = thread.Message
	}

	if thread.Title != "" {
		t.Title = thread.Title
	}

	query := `UPDATE threads
	SET message=$1, title=$2
	WHERE id = $3
	RETURNING author, created_at, forum, slug, votes`

	row := r.db.QueryRow(context.Background(), query, t.Message, t.Title, t.Id)
	res := models.Thread{}

	err := row.Scan(&res.Author,&res.CreatedAt, &res.Forum, &res.Slug, &res.Votes)
	if err == nil {
	}
	res.Message = t.Message
	res.Title = t.Title
	res.Id = t.Id
	return res, models.OK
}

func (r *Repo) ThreadVoteID(check int, vote models.Vote) (models.Thread, int) {

	user, code := r.CheckUser(models.User{NickName: vote.NickName})
	if code != models.OK {
		return models.Thread{}, models.NotFound
	}
	vote.NickName = user.NickName

	thread, _ := r.GetThreadByID(check, models.Thread{})

	query := `INSERT INTO VOTES (author, vote, thread) VALUES ($1, $2, $3) RETURNING vote`
	row := r.db.QueryRow(context.Background(), query, vote.NickName, vote.Voice, thread.Id)

	value := 0

	err := row.Scan(&vote.Voice)
	if err != nil {
		if pqError, ok := err.(*pgconn.PgError); ok {
			switch pqError.Code {
			case "23503":
				return thread, models.NotFound
			case "23505":
				upd := "WITH u AS ( SELECT vote FROM votes WHERE author = $2 AND thread = $3)" +
					"UPDATE votes SET vote =  $1 WHERE author = $2 AND thread = $3 " +
					"RETURNING vote, (SELECT vote FROM u)"
				row := r.db.QueryRow(context.Background(), upd, vote.Voice, vote.NickName, thread.Id)
				err := row.Scan(&vote.Voice, &value)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
	query = `UPDATE threads SET votes=votes+$1 WHERE id = $2;`
	_, err = r.db.Exec(context.Background(), query, vote.Voice-value, thread.Id)

	thread, _ = r.GetThreadByID(thread.Id, models.Thread{})
	return thread, models.OK
}

func (r *Repo) CreateThreadPostID(id int, posts []models.Post) ([]models.Post, int) {
	thread := models.Thread{Id: id}

	query := `SELECT forum
					FROM threads
					WHERE id = $1`

	row := r.db.QueryRow(context.Background(), query, id)
	err := row.Scan(&thread.Forum)

	if err != nil {
		log.Print(err)
		return posts, models.NotFound
	}

	times := time.Now()

	if len(posts) == 0 {
		return posts, models.OK
	}

	tx, err := r.db.Begin(context.Background())
	query = `INSERT INTO posts (author, post, created_at, forum,  isEdited, parent, thread, path) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
			RETURNING  id;`
	ins, _ := tx.Prepare(context.Background(), "insert", query)

	result := []models.Post{}
	for _, p := range posts {
		p.Forum = thread.Forum
		p.Thread = thread.Id

		if p.Parent != 0 {
			old := 0
			query2 := `SELECT thread FROM posts WHERE id = $1`
			row = tx.QueryRow(context.Background(), query2, p.Parent)
			err := row.Scan(&old)
			if err != nil || old != p.Thread {
				return []models.Post{}, models.ForumConflict
			}
		}

		err = tx.QueryRow(context.Background(), ins.Name, p.Author, p.Message, times, thread.Forum, false, p.Parent, thread.Id, []int{}).Scan(&p.Id)

		p.CreatedAt = times.Format(time.RFC3339)

		if err != nil {
			tx.Rollback(context.Background())
			if pqError, ok := err.(pgx.PgError); ok {
				switch pqError.Code {
				case "23503":
					return []models.Post{}, models.NotFound
				case "23505":
					return []models.Post{}, models.ForumConflict
				case "22000":
					return []models.Post{}, models.NotFound
				default:
					return []models.Post{}, models.ForumConflict
				}
			}
		}

		query2 := `INSERT INTO forum_users(nickname,
    	forum) VALUES ($1, $2) ON CONFLICT DO NOTHING;`
		_, err = r.db.Exec(context.Background(), query2, p.Author, p.Forum)
		if err != nil {
			return []models.Post{}, models.NotFound
		}
		result = append(result, p)
		r.info.Post++
	}
	tx.Commit(context.Background())
	query3 := `UPDATE forums SET posts = posts + $2 WHERE slug =$1`
	_, err = r.db.Exec(context.Background(), query3, thread.Forum, len(result))
	if err != nil {
		log.Fatal(err)
	}
	return result, models.OK
}

func (r *Repo) GetThreadByID(id int, thread models.Thread) (models.Thread, int) {
	var row v4.Row

	query := `SELECT author, message, title, created_at, forum, slug, votes
					FROM threads
					WHERE id = $1`
	row = r.db.QueryRow(context.Background(), query, id)

	err := row.Scan(&thread.Author, &thread.Message, &thread.Title,
		&thread.CreatedAt, &thread.Forum, &thread.Slug, &thread.Votes)
	thread.Id = id

	if err != nil {
		return thread, models.NotFound
	}

	return thread, models.OK
}

// /thread/{slug_or_id}/details
func (r *Repo) ThreadUpdateID(id int, thread models.Thread) (models.Thread, int) {
	t := models.Thread{}
	var row v4.Row

	query := `SELECT author, message, title, created_at, forum, slug, votes
					FROM threads
					WHERE id = $1`
	row = r.db.QueryRow(context.Background(), query, id)

	err := row.Scan(&t.Author, &t.Message, &t.Title,
		&t.CreatedAt, &t.Forum, &t.Slug, &t.Votes)
	t.Id = id

	if err != nil {
		return thread, models.NotFound
	}

	if thread.Message != "" {
		t.Message = thread.Message
	}

	if thread.Title != "" {
		t.Title = thread.Title
	}

	query = `UPDATE threads
	SET message=$1, title=$2
	WHERE id = $3`

	_, err = r.db.Exec(context.Background(), query, t.Message, t.Title, t.Id)

	if err == nil {
	}
	return t, models.OK
}

func (r *Repo) GetThreadsPostsID(limit, since, desc, sort string, id int) (models.Posts, int) {

	var row v4.Rows
	ps := models.Posts{}
	thread := models.Thread{}

	query := `SELECT id	FROM threads WHERE id = $1`
	rows := r.db.QueryRow(context.Background(), query, id)

	err := rows.Scan(&thread.Id)

	if err != nil {
		return ps, models.NotFound
	}

	switch sort {
	case "flat":
		row = r.getFlat(thread.Id, since, limit, desc)

	case "tree":
		row = r.getTree(thread.Id, since, limit, desc)

	case "parent_tree":
		row = r.getParentTree(thread.Id, since, limit, desc)

	default:
		row = r.getFlat(thread.Id, since, limit, desc)
	}

	defer row.Close()
	for row.Next() {

		pr := models.Post{}
		times := time.Time{}

		err := row.Scan(&pr.Id, &pr.Author, &pr.Message, &times, &pr.Forum, &pr.IsEdited, &pr.Parent)
		pr.Thread = thread.Id
		pr.CreatedAt = times.Format(time.RFC3339)
		if err != nil {
		}
		ps = append(ps, pr)
	}

	return ps, models.OK
}

func (r *Repo) GetSlugID(check string, thread models.Thread) (models.Thread, int) {
	var row v4.Row

	query := `SELECT id FROM threads WHERE slug = $1`
	row = r.db.QueryRow(context.Background(), query, check)

	err := row.Scan(&thread.Id)

	if err != nil {
		return thread, models.NotFound
	}

	return thread, models.OK
}
