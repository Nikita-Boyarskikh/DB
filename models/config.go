package models

//easyjson:json
type Config struct {
	Server struct {
		Host                          string
		Port                          int
		Concurrency                   int
		ReadBufferSize                int
		WriteBufferSize               int
		ReadTimeout                   int
		WriteTimeout                  int
		MaxRequestBodySize            int
		LogAllErrors                  bool
		DisableHeaderNamesNormalizing bool
	}
	DB struct {
		AcquireTimeout int
		MaxConnections int
	}
	TimestampOutLayout string
	TimestampInLayout  string
}
