package database

import (
	"context"
	"fmt"
	"forum/internal/models"
	"github.com/jackc/pgx/v4"
	"time"
)

func (r *Repo) GetPost(ps models.PostFull, related []string) (models.PostFull, int) {
	pr := models.PostFull{
		Author: nil,
		Forum:  nil,
		Post:   models.Post{},
		Thread: nil,
	}

	p := models.Post{}
	p.Id = ps.Post.Id
	query := `SELECT author, post, created_at, forum, isedited, parent, thread
	FROM posts 
	WHERE id = $1`

	times := time.Time{}
	row := r.db.QueryRow(context.Background(), query, ps.Post.Id)
	err := row.Scan(&p.Author, &p.Message, &times,
		&p.Forum, &p.IsEdited, &p.Parent, &p.Thread)
	p.CreatedAt = times.Format(time.RFC3339)

	if err != nil {
		return pr, models.NotFound
	}

	pr.Post = p

	for j := 0; j < len(related); j++ {
		if related[j] == "user" {
			u, _ := r.GetUser(models.User{NickName: p.Author})
			pr.Author = &u
		}
		if related[j] == "forum" {

			f, _ := r.GetForumBySlag(models.Forum{Slug: p.Forum})
			pr.Forum = &f

		}
		if related[j] == "thread" {
			t, _ := r.GetThreadByID(p.Thread, models.Thread{})
			pr.Thread = &t

		}
	}
	return pr, models.OK
}

///post/{id}/details
func (r *Repo) UpdatePost(update models.PostUpdate) (models.Post, int) {

	res := models.Post{}
	//проверить наличие поста
	query := `SELECT id, author, post, created_at,
                       forum, isEdited, parent, thread
				FROM posts 
				WHERE id = $1 `

	row := r.db.QueryRow(context.Background(), query, update.Id)

	times := time.Time{}
	err := row.Scan(&res.Id, &res.Author, &res.Message, &times,
		&res.Forum, &res.IsEdited, &res.Parent, &res.Thread)
	res.CreatedAt = times.Format(time.RFC3339)
	//поста нет
	if err != nil {
		return models.Post{}, models.NotFound
	}

	if update.Message == "" || update.Message == res.Message {
		return res, models.OK
	}

	queryupdate := `UPDATE posts SET post = $1, isEdited = $2 WHERE id = $3`
	rowup, err := r.db.Exec(context.Background(), queryupdate, update.Message, true, update.Id)

	if err != nil || rowup.RowsAffected() == 0 {
		return models.Post{}, models.NotFound
	}

	res.Message = update.Message
	res.IsEdited = true

	return res, models.OK
}

func (r *Repo) getFlat(id int, since, limit, desc string) pgx.Rows {
	var rows pgx.Rows

	query := `SELECT id, author, post, created_at, forum, isedited, parent, thread
				FROM posts
				WHERE thread = $1`

	if limit == "" && since == "" {
		if desc == "true" {
			query += ` ORDER BY id DESC`
		} else {
			query += ` ORDER BY id ASC`
		}
		rows, _ = r.db.Query(context.Background(), query, id)
	} else {
		if limit != "" && since == "" {
			if desc == "true" {
				query += ` ORDER BY id DESC LIMIT $2`
			} else {
				query += `ORDER BY id ASC LIMIT $2`
			}
			rows, _ = r.db.Query(context.Background(), query, id, limit)
		}

		if limit != "" && since != "" {
			if desc == "true" {
				query += `AND id < $2 ORDER BY id DESC LIMIT $3`
			} else {
				query += `AND id > $2 ORDER BY id ASC LIMIT $3`
			}
			rows, _ = r.db.Query(context.Background(), query, id, since, limit)
		}

		if limit == "" && since != "" {
			if desc == "true" {
				query += `AND id < $2 ORDER BY id DESC`
			} else {
				query += `AND id > $2 ORDER BY id ASC`
			}
			rows, _ = r.db.Query(context.Background(), query, id, since)
		}
	}

	return rows
}

func (r *Repo) getTree(id int, since, limit, desc string) pgx.Rows {

	var rows pgx.Rows

	query := ``

	if limit == "" && since == "" {
		if desc == "true" {
			query = `SELECT id, author, post, created_at, forum, isedited, parent, thread
				FROM posts
				WHERE thread = $1 ORDER BY path, id DESC`
		} else {
			query = ` SELECT id, author, post, created_at, forum, isedited, parent, thread
				FROM posts
				WHERE thread = $1 ORDER BY path, id ASC`
		}
		rows, _ = r.db.Query(context.Background(), query, id)
	} else {
		if limit != "" && since == "" {
			if desc == "true" {
				query += `SELECT id, author, post, created_at, forum, isedited, parent, thread
				FROM posts
				WHERE thread = $1 ORDER BY path DESC, id DESC LIMIT $2`
			} else {
				query += `SELECT id, author, post, created_at, forum, isedited, parent, thread
				FROM posts
				WHERE thread = $1 ORDER BY path, id ASC LIMIT $2`
			}
			rows, _ = r.db.Query(context.Background(), query, id, limit)
		}

		if limit != "" && since != "" {
			if desc == "true" {
				query = `SELECT posts.id, posts.author, posts.post, 
				posts.created_at, posts.forum, posts.isedited, posts.parent, posts.thread
				FROM posts JOIN posts parent ON parent.id = $2 WHERE posts.path < parent.path AND  posts.thread = $1
				ORDER BY posts.path DESC, posts.id DESC LIMIT $3`
			} else {
				query = `SELECT posts.id, posts.author, posts.post, 
				posts.created_at, posts.forum, posts.isedited, posts.parent, posts.thread
				FROM posts JOIN posts parent ON parent.id = $2 WHERE posts.path > parent.path AND  posts.thread = $1
				ORDER BY posts.path ASC, posts.id ASC LIMIT $3`
			}
			rows, _ = r.db.Query(context.Background(), query, id, since, limit)
		}

		if limit == "" && since != "" {
			if desc == "true" {
				query = `SELECT posts.id, posts.author, posts.post, 
				posts.created_at, posts.forum, posts.isedited, posts.parent, posts.thread
				FROM posts JOIN posts parent ON parent.id = $2 WHERE posts.path < parent.path AND  posts.thread = $1
				ORDER BY posts.path DESC, posts.id DESC`
			} else {
				query = `SELECT posts.id, posts.author, posts.post, 
				posts.created_at, posts.forum, posts.isedited, posts.parent, posts.thread
				FROM posts JOIN posts parent ON parent.id = $2 WHERE posts.path > parent.path AND  posts.thread = $1
				ORDER BY posts.path ASC, posts.id ASC`
			}
			rows, _ = r.db.Query(context.Background(), query, id, since)
		}
	}

	return rows
}

func (r *Repo) getParentTree(id int, since, limit, desc string) pgx.Rows {
	var rows pgx.Rows

	//все корневые посты
	parents := fmt.Sprintf(`SELECT id FROM posts WHERE thread = %d AND parent = 0 `, id)

	if since != "" {
		if desc == "true" {
			parents += ` AND path[1] < ` + fmt.Sprintf(`(SELECT path[1] FROM posts WHERE id = %s) `, since)
		} else {
			parents += ` AND path[1] > ` + fmt.Sprintf(`(SELECT path[1] FROM posts WHERE id = %s) `, since)
		}
	}

	if desc == "true" {
		parents += ` ORDER BY id DESC `
	} else {
		parents += ` ORDER BY id ASC `
	}

	if limit != "" {
		parents += " LIMIT " + limit
	}

	query := fmt.Sprintf(
		`SELECT id, author, post, created_at, forum, isedited, parent, thread FROM posts WHERE path[1] = ANY (%s) `, parents)

	if desc == "true" {
		query += ` ORDER BY path[1] DESC, path,  id `
	} else {
		query += ` ORDER BY path[1] ASC, path,  id `
	}

	rows, _ = r.db.Query(context.Background(), query)
	return rows

}
