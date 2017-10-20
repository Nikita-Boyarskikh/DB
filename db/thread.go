package db

import (
	"time"

	"strconv"

	"github.com/Nikita-Boyarskikh/DB/config"
	"github.com/Nikita-Boyarskikh/DB/models"
	"github.com/jackc/pgx"
	"github.com/mailru/easyjson/opt"
)

// TODO: Refactor this method!
func CreateThread(t models.Thread) (models.Thread, error) {
	var (
		id      int32
		created time.Time
	)

	if t.Created.Defined {
		var err error
		created, err = time.Parse(config.TimestampInLayout, t.Created.V)
		if err != nil {
			return models.Thread{}, err
		}

		log.Println(t.Created.V, config.TimestampInLayout, created)

		if t.Slug.Defined {
			log.Printf(`INSERT INTO threads(authorID, created, forumID, message, title, slug)
				VALUES (%s, %s, %s, %s, %s, %s)`,
				t.Author, t.Created.V, t.Forum.V, t.Message, t.Title, t.Slug.V)
			if err := conn.QueryRow(`INSERT INTO threads(authorID, created, forumID, message, title, slug)
				VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created AT TIME ZONE 'UTC'`,
				t.Author, created, t.Forum.V, t.Message, t.Title, t.Slug.V).
				Scan(&id, &created); err != nil {
				return models.Thread{}, err
			}
		} else {
			log.Printf(`INSERT INTO threads(authorID, created, forumID, message, title) VALUES (%s, %s, %s, %s, %s)`,
				t.Author, t.Created.V, t.Forum.V, t.Message, t.Title)
			if err := conn.QueryRow(`INSERT INTO threads(authorID, created, forumID, message, title)
				VALUES ($1, $2, $3, $4, $5) RETURNING id, created AT TIME ZONE 'UTC'`,
				t.Author, created, t.Forum.V, t.Message, t.Title).Scan(&id, &created); err != nil {
				return models.Thread{}, err
			}
		}
	} else {
		if t.Slug.Defined {
			log.Printf(`INSERT INTO threads(authorID, forumID, message, title, slug) VALUES (%s, %s, %s, %s, %s)`,
				t.Author, t.Forum.V, t.Message, t.Title, t.Slug.V)
			if err := conn.QueryRow(`INSERT INTO threads(authorID, forumID, message, title, slug)
				VALUES ($1, $2, $3, $4, $5) RETURNING id, created AT TIME ZONE 'UTC'`,
				t.Author, t.Forum.V, t.Message, t.Title, t.Slug.V).Scan(&id, &created); err != nil {
				return models.Thread{}, err
			}
		} else {
			log.Printf(`INSERT INTO threads(authorID, forumID, message, title) VALUES (%s, %s, %s, %s)`,
				t.Author, t.Forum.V, t.Message, t.Title)
			if err := conn.QueryRow(`INSERT INTO threads(authorID, forumID, message, title)
				VALUES ($1, $2, $3, $4) RETURNING id, created AT TIME ZONE 'UTC'`,
				t.Author, t.Forum.V, t.Message, t.Title).Scan(&id, &created); err != nil {
				return models.Thread{}, err
			}
		}
	}

	t.Votes = opt.OInt32(0)
	t.Created = opt.OString(created.Format(config.TimestampOutLayout))
	t.ID = opt.OInt32(id)
	if err := NewThread(t.Forum.V); err != nil {
		return models.Thread{}, nil
	}
	return t, nil
}

func GetThreadBySlugOrID(slug string, ID int32) (models.Thread, error) {
	var (
		authorID string
		created  time.Time
		forumID  string
		message  string
		title    string
		votes    int32
	)

	log.Printf(`SELECT id, authorID, created AT TIME ZONE 'UTC', forumID, message, title, slug, votes FROM threads
		WHERE slug = %s OR id = %d`, slug, ID)
	if err := conn.QueryRow(`SELECT id, authorID, created AT TIME ZONE 'UTC', forumID, message, title, slug, votes FROM threads
		WHERE slug = $1 OR id = $2`, slug, ID).
		Scan(&ID, &authorID, &created, &forumID, &message, &title, &slug, &votes); err != nil {
		return models.Thread{}, err
	}

	return models.Thread{
		ID:      opt.OInt32(ID),
		Author:  authorID,
		Forum:   opt.OString(forumID),
		Created: opt.OString(created.Format(config.TimestampOutLayout)),
		Message: message,
		Title:   title,
		Slug:    opt.OString(slug),
		Votes:   opt.OInt32(votes),
	}, nil
}

func GetThreadsBySlug(slug string, limit int, since time.Time, desc bool) (models.Threads, error) {
	var (
		orderDir string
		sign     string
	)

	if desc {
		sign = "<="
		orderDir = "DESC"
	} else {
		sign = ">="
		orderDir = "ASC"
	}

	log.Printf(`SELECT id, authorID, created AT TIME ZONE 'UTC', forumID, message, title, slug, votes FROM threads
		WHERE forumID = %s AND created `+sign+` %s ORDER BY created `+orderDir+` LIMIT %s`, slug, since, limit)
	rows, err := conn.Query(`SELECT id, authorID, created AT TIME ZONE 'UTC', forumID, message, title, slug, votes FROM threads
		WHERE forumID = $1 AND created `+sign+` $2 ORDER BY created `+orderDir+` LIMIT $3`, slug, since, limit)
	if err != nil {
		return models.Threads{}, err
	}

	var threads models.Threads
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return models.Threads{}, err
		}

		threads = append(threads, models.Thread{
			ID:      opt.OInt32(vals[0].(int32)),
			Author:  vals[1].(string),
			Created: opt.OString(vals[2].(time.Time).Format(config.TimestampOutLayout)),
			Forum:   opt.OString(vals[3].(string)),
			Message: vals[4].(string),
			Title:   vals[5].(string),
			Slug:    opt.OString(vals[6].(string)),
			Votes:   opt.OInt32(vals[7].(int32)),
		})
	}

	return threads, nil
}

// TODO: Refactor this method!
func UpdateThreadBySlugOrID(slugOrID string, t models.PatchThread) (models.Thread, error) {
	var (
		result  models.Thread
		ID      int32
		created time.Time
		forumID string
		slug    string
		votes   int32
	)

	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		ID = 0
	} else {
		ID = int32(id)
	}

	if t.Title.Defined && t.Message.Defined {
		log.Printf(`UPDATE threads SET title=%s, message=%s WHERE ID=%d OR slug=%s`, t.Title.V, t.Message.V, ID, slugOrID)
		if err := conn.QueryRow(`UPDATE threads SET title=$1, message=$2 WHERE ID=$3 OR slug=$4
			RETURNING ID, authorID, created AT TIME ZONE 'UTC', forumID, slug, votes`, t.Title.V, t.Message.V, ID, slugOrID).
			Scan(&ID, &result.Author, &created, &forumID, &slug, &votes); err != nil {
			return models.Thread{}, err
		}

		result.Message = t.Message.V
		result.Title = t.Title.V
	} else if t.Title.Defined {
		log.Printf(`UPDATE threads SET title=%s WHERE ID=%d OR slug=%s`, t.Title.V, ID, slugOrID)
		if err := conn.QueryRow(`UPDATE threads SET title=$1 WHERE ID=$2 OR slug=$3
			RETURNING ID, authorID, created AT TIME ZONE 'UTC', forumID, message, slug, votes`, t.Title.V, ID, slugOrID).
			Scan(&ID, &result.Author, &created, &forumID, &result.Message, &slug, &votes); err != nil {
			return models.Thread{}, err
		}

		result.Title = t.Title.V
	} else if t.Message.Defined {
		log.Printf(`UPDATE threads SET message=%s WHERE ID=%d OR slug=%s`, t.Message.V, ID, slugOrID)
		if err := conn.QueryRow(`UPDATE threads SET message=$1 WHERE ID=$2 OR slug=$3
			RETURNING ID, authorID, created AT TIME ZONE 'UTC', forumID, title, slug, votes`, t.Message.V, ID, slugOrID).
			Scan(&ID, &result.Author, &created, &forumID, &result.Title, &slug, &votes); err != nil {
			return models.Thread{}, err
		}

		result.Message = t.Message.V
	} else {
		if result, err = GetThreadBySlugOrID(slugOrID, ID); err != nil {
			return models.Thread{}, err
		}

		return result, nil
	}

	result.ID = opt.OInt32(ID)
	result.Created = opt.OString(created.Format(config.TimestampOutLayout))
	result.Forum = opt.OString(forumID)
	result.Slug = opt.OString(slug)
	result.Votes = opt.OInt32(votes)

	return result, nil
}

// TODO: Refactor this method!
func VoteForThread(ID int32, vote models.Vote) (models.Thread, error) {
	var (
		result  models.Thread
		created time.Time
		forumID string
		slug    string
		votes   int32
		voice   int32
	)

	log.Printf(`SELECT voice FROM voices WHERE threadID=%d AND userID=%s`, ID, vote.Nickname)
	if err := conn.QueryRow(`SELECT voice FROM voices WHERE threadID=$1 AND userID=$2`, ID, vote.Nickname).
		Scan(&voice); err != nil && err != pgx.ErrNoRows {
		return models.Thread{}, err
	} else if err == nil && voice == vote.Voice {
		log.Printf(`SELECT authorID, created AT TIME ZONE 'UTC', forumID, message, title, slug, votes FROM threads
			WHERE ID=%d`, ID)
		if err := conn.QueryRow(`SELECT authorID, created AT TIME ZONE 'UTC', forumID, message, title, slug, votes FROM threads
			WHERE ID=$1`, ID).
			Scan(&result.Author, &created, &forumID, &result.Message, &result.Title, &slug, &votes); err != nil {
			return models.Thread{}, err
		}
	} else if err == nil {
		log.Printf(`UPDATE voices SET voice=%d WHERE threadID=%d AND userID=%s`, vote.Voice, ID, vote.Nickname)
		_, err := conn.Exec(`UPDATE voices SET voice=$1 WHERE threadID=$2 AND userID=$3`, vote.Voice, ID, vote.Nickname)
		if err != nil {
			return models.Thread{}, err
		}

		log.Printf(`UPDATE threads SET votes=votes+2*(%d) WHERE ID=%d`, vote.Voice, ID)
		if err := conn.QueryRow(`UPDATE threads SET votes=votes+2*($1) WHERE ID=$2
		RETURNING authorID, created AT TIME ZONE 'UTC', forumID, message, title, slug, votes`, vote.Voice, ID).
			Scan(&result.Author, &created, &forumID, &result.Message, &result.Title, &slug, &votes); err != nil {
			return models.Thread{}, err
		}
	} else {
		log.Printf(`INSERT INTO voices(threadID, userID, voice) VALUES (%d, %s, %d)`, ID, vote.Nickname, vote.Voice)
		_, err := conn.Exec(`INSERT INTO voices(threadID, userID, voice) VALUES ($1, $2, $3)`, ID, vote.Nickname, vote.Voice)
		if err != nil {
			return models.Thread{}, err
		}

		log.Printf(`UPDATE threads SET votes=votes+(%d) WHERE ID=%d`, vote.Voice, ID)
		if err := conn.QueryRow(`UPDATE threads SET votes=votes+($1) WHERE ID=$2
		RETURNING authorID, created AT TIME ZONE 'UTC', forumID, message, title, slug, votes`, vote.Voice, ID).
			Scan(&result.Author, &created, &forumID, &result.Message, &result.Title, &slug, &votes); err != nil {
			return models.Thread{}, err
		}
	}

	result.ID = opt.OInt32(ID)
	result.Created = opt.OString(created.Format(config.TimestampOutLayout))
	result.Forum = opt.OString(forumID)
	result.Slug = opt.OString(slug)
	result.Votes = opt.OInt32(votes)

	return result, nil
}
