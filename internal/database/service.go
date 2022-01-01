package database

import (
	"context"
	"forum/internal/models"
)

// /service/clear
func (r *Repo) Clear() error {
	query := `TRUNCATE TABLE users, forums, threads, post CASCADE;`
	r.db.Exec(context.Background(), query)

	r.info = models.Status{
		Forum:  0,
		Post:   0,
		Thread: 0,
		User:   0,
	}
	return nil
}

// /service/status
func (r *Repo) Info() models.Status {
	return r.info
}
