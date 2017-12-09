package db

import (
	"fmt"

	"strconv"

	"github.com/Nikita-Boyarskikh/DB/models"
	"github.com/jackc/pgx"
	"github.com/mailru/easyjson/opt"
	"github.com/pkg/errors"
)

// TODO: Refactor this method!
func UpdateUser(tx *pgx.Tx, u models.User) (models.User, error) {
	var (
		result   models.User
		nickname = u.Nickname.V
		about    = opt2string(u.About, "")
	)
	if !u.Nickname.Defined {
		message := "User nickname id not defined"
		return models.User{}, errors.New(message)
	}

	if u.Fullname == "" && u.Email == "" && about == "" {
		if err := tx.QueryRow(
			`SELECT nickname, fullname, email, about FROM users WHERE nickname = $1`, nickname,
		).Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
			return models.User{}, err
		}
	} else {
		if u.Fullname == "" && u.Email == "" {
			if err := tx.QueryRow(`UPDATE users SET about = $1 WHERE nickname = $2
				RETURNING nickname, fullname, email, about`, about, nickname).
				Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
				return models.User{}, err
			}
		} else {
			if u.Fullname == "" {
				if about == "" {
					if err := tx.QueryRow(`UPDATE users SET email = $1 WHERE nickname = $2
				RETURNING nickname, fullname, email, about`,
						u.Email, nickname).
						Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
						return models.User{}, err
					}
				} else {
					if err := tx.QueryRow(`UPDATE users SET email = $1, about = $2 WHERE nickname = $3
				RETURNING nickname, fullname, email, about`,
						u.Email, about, nickname).
						Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
						return models.User{}, err
					}
				}
			} else {
				if about == "" {
					if u.Email == "" {
						if err := tx.QueryRow(`UPDATE users SET fullname = $1 WHERE nickname = $2
				RETURNING nickname, fullname, email, about`,
							u.Fullname, nickname).
							Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
							return models.User{}, err
						}
					} else {
						if err := tx.QueryRow(`UPDATE users SET fullname = $1, email = $2 WHERE nickname = $3
				RETURNING nickname, fullname, email, about`,
							u.Fullname, u.Email, nickname).
							Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
							return models.User{}, err
						}
					}
				} else {
					if u.Email == "" {
						if err := tx.QueryRow(`UPDATE users SET fullname = $1, about = $2 WHERE nickname = $3
				RETURNING nickname, fullname, email, about`,
							u.Fullname, about, nickname).
							Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
							return models.User{}, err
						}
					} else {
						if err := tx.QueryRow(`UPDATE users SET fullname = $1, email = $2, about = $3 WHERE nickname = $4
				RETURNING nickname, fullname, email, about`,
							u.Fullname, u.Email, about, nickname).
							Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
							return models.User{}, err
						}
					}
				}
			}
		}
	}

	result.Nickname = opt.OString(nickname)
	result.About = opt.OString(about)
	return result, nil
}

func CreateUser(tx *pgx.Tx, u models.User) error {
	about := opt2string(u.About, "")

	if _, err := tx.Exec(
		`INSERT INTO users(nickname, fullname, email, about) VALUES ($1, $2, $3, $4)`,
		u.Nickname.V, u.Fullname, u.Email, about); err != nil {
		return err
	}

	NewUser()
	return nil
}

func GetUserByNickname(tx *pgx.Tx, nickname string) (models.User, error) {
	var (
		fullname string
		email    string
		about    string
	)

	if err := tx.QueryRow(
		`SELECT nickname, fullname, email, about FROM users WHERE nickname = $1`, nickname).
		Scan(&nickname, &fullname, &email, &about); err != nil {
		return models.User{}, err
	}

	return models.User{
		Nickname: opt.OString(nickname),
		Fullname: fullname,
		Email:    email,
		About:    opt.OString(about),
	}, nil
}

func GetUserNickname(tx *pgx.Tx, nickname string) (string, error) {
	if err := tx.QueryRow(
		`SELECT nickname FROM users WHERE nickname = $1`, nickname).Scan(&nickname); err != nil {
		return "", err
	}
	return nickname, nil
}

func GetUsersByEmailAndNickname(tx *pgx.Tx, email, nickname string) (models.Users, error) {
	rows, err := tx.Query(
		`SELECT nickname, fullname, email, about FROM users WHERE email = $1 OR nickname = $2`,
		email, nickname)
	if err != nil {
		return models.Users{}, err
	}

	var users models.Users
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return models.Users{}, err
		}

		users = append(users, models.User{
			Nickname: opt.OString(vals[0].(string)),
			Fullname: vals[1].(string),
			Email:    vals[2].(string),
			About:    opt.OString(vals[3].(string)),
		})
	}
	return users, nil
}

func CheckAllUsersExists(tx *pgx.Tx, nicknames map[string]bool) (bool, error) {
	n := -4
	for k := range nicknames {
		n += len(k) + 4
	}
	uniqNicknames := make([]byte, n)
	first := true
	bp := 0
	for k := range nicknames {
		if !first {
			bp += copy(uniqNicknames[bp:], "', '")
		}
		bp += copy(uniqNicknames[bp:], k)
		first = false
	}

	tag, err := tx.Exec(`SELECT nickname FROM users WHERE nickname IN ('` + string(uniqNicknames) + `')`)
	if err != nil {
		return false, err
	}

	return tag.RowsAffected() == int64(len(nicknames)), nil
}

func GetUsersByForumSlug(tx *pgx.Tx, slug string, since string, limit int, desc bool) (models.Users, error) {
	sqlPattern := `SELECT u.nickname, u.fullname, u.email, u.about FROM forum_users AS fu
		JOIN users AS u ON (u.nickname = fu.userID) WHERE fu.forumID = %s ORDER BY u.nickname %s`

	args := make([]string, 2)
	args[0] = "'" + slug + "'"

	if since != "" {
		if desc {
			args[0] += " AND u.nickname < '" + since + "'"
			args[1] = "DESC"
		} else {
			args[0] += " AND u.nickname > '" + since + "'"
			args[1] = "ASC"
		}
	} else {
		if desc {
			args[1] = "DESC"
		} else {
			args[1] = "ASC"
		}
	}

	if limit != -1 {
		args[1] += " LIMIT " + strconv.Itoa(limit)
	}

	sql := fmt.Sprintf(sqlPattern, args[0], args[1])
	rows, err := tx.Query(sql)
	if err != nil {
		return models.Users{}, err
	}

	var users models.Users
	for rows.Next() {
		val, err := rows.Values()
		if err != nil {
			return models.Users{}, err
		}

		users = append(users, models.User{
			Nickname: opt.OString(val[0].(string)),
			Fullname: val[1].(string),
			Email:    val[2].(string),
			About:    opt.OString(val[3].(string)),
		})
	}

	return users, nil
}
