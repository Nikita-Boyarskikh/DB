package db

import (
	"time"

	"strconv"

	"github.com/Nikita-Boyarskikh/DB/config"
	"github.com/Nikita-Boyarskikh/DB/models"
	"github.com/jackc/pgx"
	"github.com/mailru/easyjson/opt"
)

func CreateThread(tx *pgx.Tx, t models.Thread) (models.Thread, error) {
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

		if t.Slug.Defined {
			if err := tx.QueryRow(`INSERT INTO threads(authorID, created, forumID, message, title, slug)
				VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created AT TIME ZONE 'UTC'`,
				t.Author, created, t.Forum.V, t.Message, t.Title, t.Slug.V).
				Scan(&id, &created); err != nil {
				return models.Thread{}, err
			}
		} else {
			if err := tx.QueryRow(`INSERT INTO threads(authorID, created, forumID, message, title)
				VALUES ($1, $2, $3, $4, $5) RETURNING id, created AT TIME ZONE 'UTC'`,
				t.Author, created, t.Forum.V, t.Message, t.Title).Scan(&id, &created); err != nil {
				return models.Thread{}, err
			}
		}
	} else {
		if t.Slug.Defined {
			if err := tx.QueryRow(`INSERT INTO threads(authorID, forumID, message, title, slug)
				VALUES ($1, $2, $3, $4, $5) RETURNING id, created AT TIME ZONE 'UTC'`,
				t.Author, t.Forum.V, t.Message, t.Title, t.Slug.V).Scan(&id, &created); err != nil {
				return models.Thread{}, err
			}
		} else {
			if err := tx.QueryRow(`INSERT INTO threads(authorID, forumID, message, title)
				VALUES ($1, $2, $3, $4) RETURNING id, created AT TIME ZONE 'UTC'`,
				t.Author, t.Forum.V, t.Message, t.Title).Scan(&id, &created); err != nil {
				return models.Thread{}, err
			}
		}
	}

	t.Votes = opt.OInt32(0)
	t.Created = opt.OString(created.Format(config.TimestampOutLayout))
	t.ID = opt.OInt32(id)
	if err := NewThread(tx, t.Forum.V, t.Author); err != nil {
		return models.Thread{}, err
	}
	return t, nil
}

func GetThreadBySlugOrID(tx *pgx.Tx, slug string, ID int32) (models.Thread, error) {
	var (
		authorID string
		created  time.Time
		forumID  string
		message  string
		title    string
		votes    int32
		resSlug  *string
	)

	if err := tx.QueryRow(`SELECT ID, authorID, created AT TIME ZONE 'UTC', forumID, message, title, slug, votes FROM threads
		WHERE slug = $1 OR ID = $2`, slug, ID).
		Scan(&ID, &authorID, &created, &forumID, &message, &title, &resSlug, &votes); err != nil {
		return models.Thread{}, err
	}

	var optSlug opt.String
	if resSlug != nil {
		optSlug = opt.OString(*resSlug)
	}

	return models.Thread{
		ID:      opt.OInt32(ID),
		Author:  authorID,
		Forum:   opt.OString(forumID),
		Created: opt.OString(created.Format(config.TimestampOutLayout)),
		Message: message,
		Title:   title,
		Slug:    optSlug,
		Votes:   opt.OInt32(votes),
	}, nil
}

func GetThreadIDAndForumBySlugOrID(tx *pgx.Tx, slug string, ID int32) (int32, string, error) {
	if err := tx.QueryRow(`SELECT ID, forumID FROM threads WHERE slug = $1 OR ID = $2`, slug, ID).
		Scan(&ID, &slug); err != nil {
		return -1, "", err
	}
	return ID, slug, nil
}

func GetThreadIDBySlugOrID(tx *pgx.Tx, slug string, ID int32) (int32, error) {
	if err := tx.QueryRow(`SELECT ID FROM threads WHERE slug = $1 OR ID = $2`, slug, ID).Scan(&ID); err != nil {
		return -1, err
	}
	return ID, nil
}

func GetThreadsBySlug(tx *pgx.Tx, slug string, limit int, since time.Time, desc bool) (models.Threads, error) {
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

	rows, err := tx.Query(`SELECT id, authorID, created AT TIME ZONE 'UTC', forumID, message, title, slug, votes FROM threads
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

		if vals[6] == nil {
			vals[6] = ""
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
func UpdateThreadBySlugOrID(tx *pgx.Tx, slugOrID string, t models.PatchThread) (models.Thread, error) {
	var (
		result  models.Thread
		ID      int32
		created time.Time
		forumID string
		slug    *string
		votes   int32
	)

	id, err := strconv.Atoi(slugOrID)
	if err != nil {
		ID = 0
	} else {
		ID = int32(id)
	}

	if t.Title.Defined && t.Message.Defined {
		if err := tx.QueryRow(`UPDATE threads SET title=$1, message=$2 WHERE ID=$3 OR slug=$4
			RETURNING ID, authorID, created AT TIME ZONE 'UTC', forumID, slug, votes`, t.Title.V, t.Message.V, ID, slugOrID).
			Scan(&ID, &result.Author, &created, &forumID, &slug, &votes); err != nil {
			return models.Thread{}, err
		}

		result.Message = t.Message.V
		result.Title = t.Title.V
	} else if t.Title.Defined {
		if err := tx.QueryRow(`UPDATE threads SET title=$1 WHERE ID=$2 OR slug=$3
			RETURNING ID, authorID, created AT TIME ZONE 'UTC', forumID, message, slug, votes`, t.Title.V, ID, slugOrID).
			Scan(&ID, &result.Author, &created, &forumID, &result.Message, &slug, &votes); err != nil {
			return models.Thread{}, err
		}

		result.Title = t.Title.V
	} else if t.Message.Defined {
		if err := tx.QueryRow(`UPDATE threads SET message=$1 WHERE ID=$2 OR slug=$3
			RETURNING ID, authorID, created AT TIME ZONE 'UTC', forumID, title, slug, votes`, t.Message.V, ID, slugOrID).
			Scan(&ID, &result.Author, &created, &forumID, &result.Title, &slug, &votes); err != nil {
			return models.Thread{}, err
		}

		result.Message = t.Message.V
	} else {
		if result, err = GetThreadBySlugOrID(tx, slugOrID, ID); err != nil {
			return models.Thread{}, err
		}

		return result, nil
	}

	var optSlug opt.String
	if slug != nil {
		optSlug = opt.OString(*slug)
	}

	result.ID = opt.OInt32(ID)
	result.Created = opt.OString(created.Format(config.TimestampOutLayout))
	result.Forum = opt.OString(forumID)
	result.Slug = optSlug
	result.Votes = opt.OInt32(votes)

	return result, nil
}

func VoteForThread(tx *pgx.Tx, thread *models.Thread, vote models.Vote) (*models.Thread, error) {
	var voice int32
	if err := tx.QueryRow(`SELECT voice FROM voices WHERE threadID=$1 AND userID=$2`, thread.ID.V, vote.Nickname).
		Scan(&voice); err != nil && err != pgx.ErrNoRows {
		return nil, err
	} else if err == nil && voice == vote.Voice {
		return thread, nil
	} else if err == nil {
		_, err := tx.Exec(`UPDATE voices SET voice=$1 WHERE threadID=$2 AND userID=$3`, vote.Voice, thread.ID.V, vote.Nickname)
		if err != nil {
			return nil, err
		}

		if err := tx.QueryRow(`UPDATE threads SET votes=votes+2*($1) WHERE ID=$2
		RETURNING votes`, vote.Voice, thread.ID.V).
			Scan(&thread.Votes.V); err != nil {
			return nil, err
		}
	} else {
		if _, err := tx.Exec(`INSERT INTO voices(threadID, userID, voice) VALUES ($1, $2, $3)`,
			thread.ID.V, vote.Nickname, vote.Voice); err != nil {
			return nil, err
		}

		if err := tx.QueryRow(`UPDATE threads SET votes=votes+($1) WHERE ID=$2
		RETURNING votes`, vote.Voice, thread.ID.V).
			Scan(&thread.Votes.V); err != nil {
			return nil, err
		}
	}

	return thread, nil
}
