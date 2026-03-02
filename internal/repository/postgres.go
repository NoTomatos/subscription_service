package repository

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func NewPostgresConnection(connString string) (*sql.DB, error) {
	logrus.WithField("dsn", connString).Info("Connecting to database")

	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	for i := 0; i < 3; i++ {
		if err := db.Ping(); err != nil {
			logrus.WithError(err).Warnf("Failed to ping database (attempt %d/3)", i+1)
			time.Sleep(2 * time.Second)
		} else {
			logrus.Info("Successfully connected to PostgreSQL")
			return db, nil
		}
	}

	db.Close()
	return nil, fmt.Errorf("failed to ping database after 3 attempts")
}

func CloseConnection(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close database connection")
		} else {
			logrus.Info("Database connection closed")
		}
	}
}
