package postgresql

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type ClientDB interface {
	Create() (*sql.DB, error)
	Ping(context.Context) error
	Close()
}

type ClientDBStruct struct {
	db     *sql.DB
	dsn    string
	logger *zap.Logger
}

func NewClientDBStruct(dsn string, logger *zap.Logger) *ClientDBStruct {
	return &ClientDBStruct{
		db:     new(sql.DB),
		dsn:    dsn,
		logger: logger,
	}
}

func (cDBS *ClientDBStruct) Create() error {
	var err error
	cDBS.db, err = sql.Open("pgx", cDBS.dsn)
	if err != nil {
		cDBS.logger.Error("creating db method, error while creating new db")
		return err
	}
	cDBS.db.SetMaxOpenConns(20)
	cDBS.db.SetMaxIdleConns(20)
	cDBS.db.SetConnMaxLifetime(time.Minute * 5)
	return nil
}

func (cDBS *ClientDBStruct) Close() {
	cDBS.db.Close()
}

func (cDBS *ClientDBStruct) Ping(parent context.Context) error {
	ctx, canselFunc := context.WithTimeout(parent, 5*time.Second)
	defer canselFunc()

	err := cDBS.db.PingContext(ctx)
	if err != nil {
		cDBS.logger.Error("db ping error")
		return err
	}
	cDBS.logger.Info("db ping success")
	return nil
}
