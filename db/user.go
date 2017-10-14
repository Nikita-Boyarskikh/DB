package db

import (
	"github.com/mailru/easyjson/opt"
	"github.com/pkg/errors"
)

//easyjson:json
type User struct {
	Nickname opt.String
	Fullname string
	Email    string
	About    opt.String
}

//easyjson:json
type Users []User

// TODO: Refactor this method!
func UpdateUser(u User) (User, error) {
	var (
		result   User
		nickname = u.Nickname.V
		about    = opt2string(u.About, "")
	)
	if !u.Nickname.Defined {
		log.Println("User nickname id not defined")
		return User{}, errors.New("User nickname id not defined")
	}

	if u.Fullname == "" && u.Email == "" && about == "" {
		log.Printf(`SELECT nickname, fullname, email, about FROM users WHERE nickname = %s`, nickname)
		if err := conn.QueryRow(
			`SELECT nickname, fullname, email, about FROM users WHERE nickname = $1`, nickname,
		).Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
			return User{}, err
		}
	} else {
		if u.Fullname == "" && u.Email == "" {
			log.Printf(`UPDATE users SET about = %s WHERE nickname = %s`, about, nickname)
			if err := conn.QueryRow(`UPDATE users SET about = $1 WHERE nickname = $2
				RETURNING nickname, fullname, email, about`, about, nickname).
				Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
				return User{}, err
			}
		} else {
			if u.Fullname == "" {
				if about == "" {
					log.Printf(`UPDATE users SET email = %s WHERE nickname = %s`,
						u.Email, nickname)
					if err := conn.QueryRow(`UPDATE users SET email = $1 WHERE nickname = $2
				RETURNING nickname, fullname, email, about`,
						u.Email, nickname).
						Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
						return User{}, err
					}
				} else {
					log.Printf(`UPDATE users SET email = %s, about = %s WHERE nickname = %s`,
						u.Email, about, nickname)
					if err := conn.QueryRow(`UPDATE users SET email = $1, about = $2 WHERE nickname = $3
				RETURNING nickname, fullname, email, about`,
						u.Email, about, nickname).
						Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
						return User{}, err
					}
				}
			} else {
				if about == "" {
					if u.Email == "" {
						log.Printf(`UPDATE users SET fullname = %s WHERE nickname = %s`,
							u.Fullname, nickname)
						if err := conn.QueryRow(`UPDATE users SET fullname = $1 WHERE nickname = $2
				RETURNING nickname, fullname, email, about`,
							u.Fullname, nickname).
							Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
							return User{}, err
						}
					} else {
						log.Printf(`UPDATE users SET fullname = %s, email = %s WHERE nickname = %s`,
							u.Fullname, u.Email, nickname)
						if err := conn.QueryRow(`UPDATE users SET fullname = $1, email = $2 WHERE nickname = $3
				RETURNING nickname, fullname, email, about`,
							u.Fullname, u.Email, nickname).
							Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
							return User{}, err
						}
					}
				} else {
					if u.Email == "" {
						log.Printf(`UPDATE users SET fullname = %s, about = %s WHERE nickname = %s`,
							u.Fullname, about, nickname)
						if err := conn.QueryRow(`UPDATE users SET fullname = $1, about = $2 WHERE nickname = $3
				RETURNING nickname, fullname, email, about`,
							u.Fullname, about, nickname).
							Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
							return User{}, err
						}
					} else {
						log.Printf(`UPDATE users SET fullname = %s, email = %s, about = %s WHERE nickname = %s`,
							u.Fullname, u.Email, about, nickname)
						if err := conn.QueryRow(`UPDATE users SET fullname = $1, email = $2, about = $3 WHERE nickname = $4
				RETURNING nickname, fullname, email, about`,
							u.Fullname, u.Email, about, nickname).
							Scan(&nickname, &result.Fullname, &result.Email, &about); err != nil {
							return User{}, err
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

func CreateUser(u User) error {
	about := opt2string(u.About, "")

	log.Printf(`INSERT INTO users(nickname, fullname, email, about) VALUES (%s, %s, %s, %s)`,
		u.Nickname.V, u.Fullname, u.Email, u.About)
	if _, err := conn.Exec(
		`INSERT INTO users(nickname, fullname, email, about) VALUES ($1, $2, $3, $4)`,
		u.Nickname.V, u.Fullname, u.Email, about); err != nil {
		return err
	}

	return nil
}

func GetUserByNickname(nickname string) (User, error) {
	var (
		fullname string
		email    string
		about    string
	)

	log.Printf(`SELECT nickname, fullname, email, about FROM users WHERE nickname = %s`, nickname)
	if err := conn.QueryRow(
		`SELECT nickname, fullname, email, about FROM users WHERE nickname = $1`, nickname).
		Scan(&nickname, &fullname, &email, &about); err != nil {
		return User{}, err
	}

	return User{
		Nickname: opt.OString(nickname),
		Fullname: fullname,
		Email:    email,
		About:    opt.OString(about),
	}, nil
}

func GetUsersByEmailAndNickname(email, nickname string) (Users, error) {
	log.Printf(
		`SELECT nickname, fullname, email, about FROM users WHERE email = %s OR nickname = %s`,
		email, nickname)
	rows, err := conn.Query(
		`SELECT nickname, fullname, email, about FROM users WHERE email = $1 OR nickname = $2`,
		email, nickname)
	if err != nil {
		return Users{}, err
	}

	var users Users
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return Users{}, err
		}

		users = append(users, User{
			Nickname: opt.OString(vals[0].(string)),
			Fullname: vals[1].(string),
			Email:    vals[2].(string),
			About:    opt.OString(vals[3].(string)),
		})
	}
	return users, nil
}
