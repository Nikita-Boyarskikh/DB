package config

import (
	"time"

	"flag"

	"io/ioutil"

	"fmt"

	"github.com/Nikita-Boyarskikh/DB/models"
	"github.com/jackc/pgx"
	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
)

var (
	TimestampOutLayout string
	TimestampInLayout  string
	URI                string
	DBConnection       pgx.ConnPoolConfig
	Server             fasthttp.Server
)

func Init() error {
	configFile := *flag.String("conf", "config/config.json", "Path to config file")
	flag.Parse()

	raw, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	config := models.Config{}
	if err := easyjson.Unmarshal(raw, &config); err != nil {
		return err
	}

	Server.Concurrency = config.Server.Concurrency
	Server.ReadBufferSize = config.Server.ReadBufferSize
	Server.WriteBufferSize = config.Server.WriteBufferSize
	Server.MaxRequestBodySize = config.Server.MaxRequestBodySize
	Server.LogAllErrors = config.Server.LogAllErrors
	Server.DisableHeaderNamesNormalizing = config.Server.DisableHeaderNamesNormalizing
	Server.ReadTimeout = time.Duration(config.Server.ReadTimeout) * time.Millisecond
	Server.WriteTimeout = time.Duration(config.Server.WriteTimeout) * time.Millisecond

	TimestampInLayout = config.TimestampInLayout
	TimestampOutLayout = config.TimestampOutLayout

	URI = fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)

	// PGHOST PGPORT PGDATABASE PGUSER PGPASSWORD
	var conf pgx.ConnConfig
	conf, err = pgx.ParseEnvLibpq()
	if err != nil {
		return err
	}
	DBConnection.ConnConfig = conf
	DBConnection.MaxConnections = config.DB.MaxConnections
	DBConnection.AcquireTimeout = time.Duration(config.DB.AcquireTimeout) * time.Millisecond

	return nil
}
