package db

import (
	"github.com/Nikita-Boyarskikh/DB/models"
	"github.com/jackc/pgx"
	"github.com/mailru/easyjson/opt"
)

func GetForumBySlug(tx *pgx.Tx, slug string) (int, models.Forum, error) {
	var (
		id       int
		posts    int64
		threads  int32
		title    string
		nickname string
	)

	if err := tx.QueryRow(`SELECT id, posts, slug, threads, title, userID FROM forums WHERE slug = $1`, slug).
		Scan(&id, &posts, &slug, &threads, &title, &nickname); err != nil {
		return -1, models.Forum{}, err
	}

	return id, models.Forum{
		Posts:   opt.OInt64(posts),
		Slug:    slug,
		Threads: opt.OInt32(threads),
		Title:   title,
		User:    nickname,
	}, nil
}

func GetForumSlug(tx *pgx.Tx, slug string) (string, error) {
	if err := tx.QueryRow(`SELECT slug FROM forums WHERE slug = $1`, slug).Scan(&slug); err != nil {
		return "", err
	}

	return slug, nil
}

func CreateForum(tx *pgx.Tx, f models.Forum) (int, error) {
	var id int
	if err := tx.QueryRow(`INSERT INTO forums(slug, title, userID) VALUES ($1, $2, $3) RETURNING id`,
		f.Slug, f.Title, f.User).Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}
