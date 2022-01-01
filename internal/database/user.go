package database

import (
	"context"
	"forum/internal/models"
	"github.com/jackc/pgconn"
)

func (r *Repo) CreateUser(user models.User) ([]models.User, int) {
	results := []models.User{}
	result := user

	query := `INSERT INTO users (email, fullname, nickname, about) 
			VALUES ($1, $2, $3, $4) RETURNING nickname`

	_, err := r.db.Exec(context.Background(), query, user.Email, user.FullName, user.NickName, user.About)
	if err != nil {
		if pqError, ok := err.(*pgconn.PgError); ok {
			switch pqError.Code {
			case "23505":
				us, _ := r.GetUserOnConflict(user)
				return us, models.ForumConflict
			default:
				us, _ := r.GetUserOnConflict(user)
				return us, models.ForumConflict
			}
		}
	}

	r.info.User++
	results = append(results, result)
	return results, models.OK
}

func (r *Repo) GetUserOnConflict(user models.User) ([]models.User, int) {
	results := []models.User{}
	query := `SELECT email, fullname, nickname, about
	FROM users
	WHERE email = $1 or nickname =  $2`

	rows, _ := r.db.Query(context.Background(), query, user.Email, user.NickName)
	defer rows.Close()

	for rows.Next() {
		result := models.User{}
		rows.Scan(&result.Email, &result.FullName, &result.NickName, &result.About)
		results = append(results, result)
	}
	return results, models.OK
}

//GET /user/{nickname}/profile
func (r *Repo) GetUser(user models.User) (models.User, int) {
	result := models.User{}
	query := `SELECT email, fullname, nickname, about
	FROM users
	WHERE  nickname =  $1`

	rows := r.db.QueryRow(context.Background(), query, user.NickName)

	err := rows.Scan(&result.Email, &result.FullName, &result.NickName, &result.About)
	if err != nil {
		return result, models.NotFound
	}
	return result, models.OK
}

func (r *Repo) CheckUser(user models.User) (models.User, int) {
	result := models.User{}
	query := `SELECT nickname
	FROM users
	WHERE nickname =  $1`

	rows := r.db.QueryRow(context.Background(), query, user.NickName)

	err := rows.Scan(&result.NickName)
	if err != nil {
		return result, models.NotFound
	}
	return result, models.OK
}

//POST /user/{nickname}/profile
func (r *Repo) UpdateUser(user models.User) (models.User, int) {

	us, status := r.GetUser(user)
	if status == models.NotFound {
		return us, models.NotFound
	}

	if user.FullName != "" {
		us.FullName = user.FullName
	}
	if user.Email != "" {
		us.Email = user.Email
	}
	if user.About != "" {
		us.About = user.About
	}

	query := `UPDATE users 
	SET fullname=$1, email=$2, about=$3 
	WHERE nickname = $4 
	RETURNING nickname, fullname, about, email;`

	rows := r.db.QueryRow(context.Background(), query, us.FullName, us.Email, us.About, us.NickName)
	err := rows.Scan(&us.NickName, &us.FullName, &us.About, &us.Email)

	if err != nil {
		if pqError, ok := err.(*pgconn.PgError); ok {
			switch pqError.Code {
			case "23505":
				return us, models.ForumConflict
			case "23503":
				return us, models.UserNotFound
			}
		}
	}

	return us, models.OK
}
