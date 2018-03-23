package db

import "fmt"

type DBConn struct {
	DBUser  string
	DBPass  string
	DBHost  string
	DBPort  string
	DBName  string
	SSLMode string
}

func (c *DBConn) DbUrl() string {
	return fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v?sslmode=%v",
		c.DBUser,
		c.DBPass,
		c.DBHost,
		c.DBPort,
		c.DBName,
		c.SSLMode,
	)
}
