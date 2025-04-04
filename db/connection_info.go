package db

import "fmt"

type ConnectionInfo struct {
	Host          string
	Port          string
	Username      string
	Password      string
	Database      string
	IsNewDatabase bool
}

func (dbConnInfo *ConnectionInfo) String() string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbConnInfo.Username,
		dbConnInfo.Password,
		dbConnInfo.Host,
		dbConnInfo.Port,
		dbConnInfo.Database,
	)
}

func (dbConnInfo *ConnectionInfo) ServerString() string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s?sslmode=disable",
		dbConnInfo.Username,
		dbConnInfo.Password,
		dbConnInfo.Host,
		dbConnInfo.Port,
	)
}

func (dbConnInfo *ConnectionInfo) WithDatabase(database string) *ConnectionInfo {
	return &ConnectionInfo{
		Host:     dbConnInfo.Host,
		Port:     dbConnInfo.Port,
		Username: dbConnInfo.Username,
		Password: dbConnInfo.Password,
		Database: database,
	}
}
