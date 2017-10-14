package db

//easyjson:json
type Status struct {
	User   int32
	Forum  int32
	Thread int32
	Post   int64
}

var (
	status      = Status{}
	initialized = false
)

func GetStatus() (Status, error) {
	if !initialized {
		log.Printf(`SELECT COUNT(nickname) FROM users`)
		if err := conn.QueryRow(`SELECT COUNT(nickname) FROM users`).Scan(&status.User); err != nil {
			return status, err
		}

		log.Printf(`SELECT COUNT(ID) FROM forums`)
		if err := conn.QueryRow(`SELECT COUNT(ID) FROM forums`).Scan(&status.Forum); err != nil {
			return status, err
		}

		log.Printf(`SELECT COUNT(ID) FROM threads`)
		if err := conn.QueryRow(`SELECT COUNT(ID) FROM threads`).Scan(&status.Thread); err != nil {
			return status, err
		}

		log.Printf(`SELECT COUNT(ID) FROM posts`)
		if err := conn.QueryRow(`SELECT COUNT(ID) FROM posts`).Scan(&status.Post); err != nil {
			return status, err
		}
	}

	initialized = true
	return status, nil
}

// TODO: Refactor this method!
func Clear() error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}

	if !initialized || status.User > 0 {
		log.Printf(`TRUNCATE TABLE users`)
		if _, err := conn.Exec(`TRUNCATE TABLE users`); err != nil {
			if err := tx.Rollback(); err != nil {
				log.Println("\t500:\t", err)
				return err
			}
			return err
		}
	}
	if !initialized || status.Forum > 0 {
		log.Printf(`TRUNCATE TABLE forums`)
		if _, err := conn.Exec(`TRUNCATE TABLE forums`); err != nil {
			if err := tx.Rollback(); err != nil {
				log.Println("\t500:\t", err)
				return err
			}
			return err
		}
	}
	if !initialized || status.Thread > 0 {
		log.Printf(`TRUNCATE TABLE threads`)
		if _, err := conn.Exec(`TRUNCATE TABLE threads`); err != nil {
			if err := tx.Rollback(); err != nil {
				log.Println("\t500:\t", err)
				return err
			}
			return err
		}

	}
	if !initialized || status.Post > 0 {
		log.Printf(`TRUNCATE TABLE posts`)
		if _, err := conn.Exec(`TRUNCATE TABLE posts`); err != nil {
			if err := tx.Rollback(); err != nil {
				log.Println("\t500:\t", err)
				return err
			}
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Println("\t500:\t", err)
		return err
	}
	return nil
}
