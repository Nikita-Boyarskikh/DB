package db

import (
	"fmt"

	"sort"
	"strings"

	"github.com/Nikita-Boyarskikh/DB/models"
	"github.com/jackc/pgx"
)

var (
	status      = models.Status{}
	initialized = false
)

func GetStatus() (models.Status, error) {
	var result error = nil

	if !initialized {
		tx, err := conn.Begin()
		if err != nil {
			return status, err
		}
		defer tx.Rollback()

		sql := `SELECT COUNT(nickname) FROM users`
		if err := tx.QueryRow(sql).Scan(&status.User); err != nil {
			return status, err
		}

		sql = `SELECT COUNT(ID) FROM forums`
		if err := tx.QueryRow(sql).Scan(&status.Forum); err != nil {
			return status, err
		}

		sql = `SELECT COUNT(ID) FROM threads`
		if err := tx.QueryRow(sql).Scan(&status.Thread); err != nil {
			return status, err
		}

		sql = `SELECT COUNT(ID) FROM posts`
		if err := tx.QueryRow(sql).Scan(&status.Post); err != nil {
			return status, err
		}

		result = tx.Commit()
	}

	initialized = true
	return status, result
}

func Clear() error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}

	if !initialized || status.User > 0 {
		if err := truncate(tx, "users"); err != nil {
			return err
		}
	}
	if !initialized || status.Forum > 0 {
		if err := truncate(tx, "forums"); err != nil {
			return err
		}
	}
	if !initialized || status.Thread > 0 {
		if err := truncate(tx, "threads"); err != nil {
			return err
		}
	}
	if !initialized || status.Post > 0 {
		if err := truncate(tx, "posts"); err != nil {
			return err
		}
	}

	status.User = 0
	status.Forum = 0
	status.Thread = 0
	status.Post = 0

	return tx.Commit()
}

func NewUser() {
	status.User++
}

func NewForum() {
	status.Forum++
}

func NewThread(tx *pgx.Tx, forumID string, userID string) error {
	if _, err := tx.Exec(`UPDATE forums SET threads = threads + 1 WHERE slug = $1`, forumID); err != nil {
		return err
	}

	if _, err := tx.Exec(`INSERT INTO forum_users (forumID, userID) VALUES ($1, $2) ON CONFLICT DO NOTHING`, forumID, userID); err != nil {
		return err
	}

	status.Thread++
	return nil
}

func NewPosts(tx *pgx.Tx, posts models.Posts) error {
	postsCount := make(map[string]int)
	args := make([]interface{}, len(posts)*2)
	sqlArgs := make([]string, len(posts))
	sortedPosts := make(models.Posts, len(posts))
	copy(sortedPosts, posts)
	sort.Sort(sortedPosts)
	for i, post := range sortedPosts {
		postsCount[post.Forum.V]++
		args[2*i] = post.Forum.V
		args[2*i+1] = post.Author
		sqlArgs[i] = fmt.Sprintf(`($%d, $%d)`, 2*i+1, 2*i+2)
	}

	for slug, n := range postsCount {
		sql := `UPDATE forums SET posts = posts + $1 WHERE slug = $2`
		if _, err := tx.Exec(sql, n, slug); err != nil {
			return err
		}
	}

	sql := `INSERT INTO forum_users (forumID, userID) VALUES %s ON CONFLICT DO NOTHING`
	if _, err := tx.Exec(fmt.Sprintf(sql, strings.Join(sqlArgs, ", ")), args...); err != nil {
		return err
	}

	status.Post += int64(len(posts))
	return nil
}

func truncate(tx *pgx.Tx, table string) error {
	sql := fmt.Sprintf(`TRUNCATE TABLE %s CASCADE`, table)
	if _, err := tx.Exec(sql); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return nil
}
