package db

import (
	"fmt"
	"strings"

	"strconv"

	"github.com/Nikita-Boyarskikh/DB/models"
	"github.com/mailru/easyjson/opt"
	"github.com/pkg/errors"
)

// TODO: Refactor this method!
func UpdateUser(u models.User) (models.User, error) {
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
		if err := conn.QueryRow(
			`SELECT nickname, fullname, email, about FROM users WHERE nickname = $1`, nickname,
		).Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
			return models.User{}, err
		}
	} else {
		if u.Fullname == "" && u.Email == "" {
			if err := conn.QueryRow(`UPDATE users SET about = $1 WHERE nickname = $2
				RETURNING nickname, fullname, email, about`, about, nickname).
				Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
				return models.User{}, err
			}
		} else {
			if u.Fullname == "" {
				if about == "" {
					if err := conn.QueryRow(`UPDATE users SET email = $1 WHERE nickname = $2
				RETURNING nickname, fullname, email, about`,
						u.Email, nickname).
						Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
						return models.User{}, err
					}
				} else {
					if err := conn.QueryRow(`UPDATE users SET email = $1, about = $2 WHERE nickname = $3
				RETURNING nickname, fullname, email, about`,
						u.Email, about, nickname).
						Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
						return models.User{}, err
					}
				}
			} else {
				if about == "" {
					if u.Email == "" {
						if err := conn.QueryRow(`UPDATE users SET fullname = $1 WHERE nickname = $2
				RETURNING nickname, fullname, email, about`,
							u.Fullname, nickname).
							Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
							return models.User{}, err
						}
					} else {
						if err := conn.QueryRow(`UPDATE users SET fullname = $1, email = $2 WHERE nickname = $3
				RETURNING nickname, fullname, email, about`,
							u.Fullname, u.Email, nickname).
							Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
							return models.User{}, err
						}
					}
				} else {
					if u.Email == "" {
						if err := conn.QueryRow(`UPDATE users SET fullname = $1, about = $2 WHERE nickname = $3
				RETURNING nickname, fullname, email, about`,
							u.Fullname, about, nickname).
							Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
							return models.User{}, err
						}
					} else {
						if err := conn.QueryRow(`UPDATE users SET fullname = $1, email = $2, about = $3 WHERE nickname = $4
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

func CreateUser(u models.User) error {
	about := opt2string(u.About, "")

	if _, err := conn.Exec(
		`INSERT INTO users(nickname, fullname, email, about) VALUES ($1, $2, $3, $4)`,
		u.Nickname.V, u.Fullname, u.Email, about); err != nil {
		return err
	}

	NewUser()
	return nil
}

func GetUserByNickname(nickname string) (models.User, error) {
	var (
		fullname string
		email    string
		about    string
	)

	if err := conn.QueryRow(
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

func GetUsersByEmailAndNickname(email, nickname string) (models.Users, error) {
	rows, err := conn.Query(
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

func CheckAllUsersExists(nicknames []string) (bool, error) {
	count := 0
	uniqNicknamesMap := make(map[string]bool)
	for _, n := range nicknames {
		uniqNicknamesMap[n] = true
	}

	var uniqNicknames []string
	for n := range uniqNicknamesMap {
		uniqNicknames = append(uniqNicknames, n)
	}

	err := conn.QueryRow(fmt.Sprintf(`SELECT COUNT(nickname) FROM users WHERE nickname IN ('%s')`,
		strings.Join(uniqNicknames, "', '"))).Scan(&count)
	if err != nil {
		return false, err
	}

	return count == len(uniqNicknames), nil
}

func GetUsersByForumSlug(slug string, since string, limit int, desc bool) (models.Users, error) {
	sqlPattern := `SELECT u.nickname, u.fullname, u.email, u.about FROM (
		SELECT authorID, forumID FROM posts UNION SELECT authorID, forumID FROM threads
	) AS ids JOIN users AS u ON (ids.authorID = u.nickname) WHERE ids.forumID = %s ORDER BY u.nickname %s`

	args := make([]string, 2)
	args[0] = "'" + slug + "'"

	if since != "" {
		if desc {
			args[0] += " AND nickname < '" + since + "'"
			args[1] = "DESC"
		} else {
			args[0] += " AND nickname > '" + since + "'"
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
	rows, err := conn.Query(sql)
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
