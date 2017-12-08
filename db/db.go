package db

import (
	"github.com/jackc/pgx"
	"github.com/mailru/easyjson/opt"
)

var conn *pgx.ConnPool

func Init(config pgx.ConnPoolConfig) error {
	var err error
	conn, err = pgx.NewConnPool(config)
	if err != nil {
		return err
	}

	// Test db is alive
	res := conn.QueryRow("SELECT 1")
	var one int
	if err := res.Scan(&one); err != nil {
		return err
	}

	return nil
}

func Vacuum() {
	conn.Exec("VACUUM ANALYZE")
}

func GetConn() *pgx.ConnPool {
	return conn
}

func Close() {
	conn.Close()
}

func opt2string(o opt.String, placeholder string) string {
	if o.Defined {
		placeholder = o.V
	}
	return placeholder
}
