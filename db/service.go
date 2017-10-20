package db

import (
	"fmt"

	"github.com/Nikita-Boyarskikh/DB/models"
	"github.com/jackc/pgx"
)

var (
	status      = models.Status{}
	initialized = false
)

func GetStatus() (models.Status, error) {
	if !initialized {
		sql := `SELECT COUNT(nickname) FROM users`
		log.Printf(sql)
		if err := conn.QueryRow(sql).Scan(&status.User); err != nil {
			return status, err
		}

		sql = `SELECT COUNT(ID) FROM forums`
		log.Printf(sql)
		if err := conn.QueryRow(sql).Scan(&status.Forum); err != nil {
			return status, err
		}

		sql = `SELECT COUNT(ID) FROM threads`
		log.Printf(sql)
		if err := conn.QueryRow(sql).Scan(&status.Thread); err != nil {
			return status, err
		}

		sql = `SELECT COUNT(ID) FROM posts`
		log.Printf(sql)
		if err := conn.QueryRow(sql).Scan(&status.Post); err != nil {
			return status, err
		}
	}

	initialized = true
	return status, nil
}

func Clear() error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}

	if !initialized || status.User > 0 {
		truncate(tx, "users")
	}
	if !initialized || status.Forum > 0 {
		truncate(tx, "forums")
	}
	if !initialized || status.Thread > 0 {
		truncate(tx, "threads")
	}
	if !initialized || status.Post > 0 {
		truncate(tx, "posts")
	}

	if err := tx.Commit(); err != nil {
		log.Println("\t500:\t", err)
		return err
	}

	status.User = 0
	status.Forum = 0
	status.Thread = 0
	status.Post = 0

	return nil
}

func NewUser() {
	status.User++
}

func NewForum() {
	status.Forum++
}

func NewThread(slug string) error {
	sql := `UPDATE forums SET threads = threads + 1 WHERE slug = `
	log.Printf(sql+"%s", slug)
	if _, err := conn.Exec(sql+"$1", slug); err != nil {
		return err
	}
	status.Thread++
	return nil
}

func NewPosts(posts models.Posts) error {
	postsCount := make(map[string]int)
	for _, post := range posts {
		postsCount[post.Forum.V]++
	}

	for slug, n := range postsCount {
		logSql := `UPDATE forums SET posts = posts + %d WHERE slug = %s`
		sql := `UPDATE forums SET posts = posts + $1 WHERE slug = $2`
		log.Printf(logSql, n, slug)
		if _, err := conn.Exec(sql, n, slug); err != nil {
			return err
		}
	}

	status.Post += int64(len(posts))
	return nil
}

func truncate(tx *pgx.Tx, table string) error {
	sql := fmt.Sprintf(`TRUNCATE TABLE %s`, table)
	log.Printf(sql)
	if _, err := conn.Exec(sql); err != nil {
		if err := tx.Rollback(); err != nil {
			log.Println("\t500:\t", err)
			return err
		}
		return err
	}
	return nil
}
