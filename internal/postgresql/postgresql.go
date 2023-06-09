package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
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
	ctx    context.Context
	cf     *config.Config
}

func NewClientDBStruct(ctx context.Context, dsn string, logger *zap.Logger, cf *config.Config) *ClientDBStruct {
	return &ClientDBStruct{
		db:     new(sql.DB),
		dsn:    dsn,
		logger: logger,
		ctx:    ctx,
		cf:     cf,
	}
}

// ///!!!!!!!!!!!!!!!!!!убрать create, усли это не нужно из конфига
func (cDBS *ClientDBStruct) Create() error {
	var err error

	if cDBS.cf.StoreType == 1 {
		return errors.New("db is not necessary")
	}
	cDBS.db, err = sql.Open("pgx", cDBS.dsn)
	if err != nil {
		cDBS.logger.Error("creating db method, error while creating new db", zap.Error(err))
		return err
	}
	cDBS.db.SetMaxOpenConns(20)
	cDBS.db.SetMaxIdleConns(20)
	cDBS.db.SetConnMaxLifetime(time.Minute * 5)

	query := `CREATE TABLE IF NOT EXIST URLStorage(URL_id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY, short_URL text, long_URL text)`
	ctx, cansel := context.WithTimeout(cDBS.ctx, 5*time.Second)
	defer cansel()

	res, err := cDBS.db.ExecContext(ctx, query)
	if err != nil {
		cDBS.logger.Error("creating db method, error while creating new table", zap.Error(err))
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		cDBS.logger.Error("Error %s when getting rows affected", zap.Error(err))
		return err
	}
	cDBS.logger.Info("Rows affected when creating table: ", zap.Int64("raws num", rows))
	return nil
}

func (cDBS *ClientDBStruct) Close() {
	cDBS.db.Close()
}

func (cDBS *ClientDBStruct) Ping() error {
	ctx, canselFunc := context.WithTimeout(cDBS.ctx, 5*time.Second)
	defer canselFunc()

	err := cDBS.db.PingContext(ctx)
	if err != nil {
		cDBS.logger.Error("db ping error")
		return err
	}
	cDBS.logger.Info("db ping success")
	return nil
}

func (cDBS *ClientDBStruct) PostShortURL(shortURL, longURL string) error {
	ctx1, canselFunc1 := context.WithTimeout(cDBS.ctx, 3*time.Second)
	defer canselFunc1()
	row := cDBS.db.QueryRowContext(ctx1, `SELECT long_URL FROM URLStorage WHERE short_URL=$N`, shortURL)

	var checkURL string
	err := row.Scan(&checkURL)
	if err != nil {
		if err != sql.ErrNoRows {
			cDBS.logger.Error("insertUTL method, error while scaning", zap.Error(err))
			return err
		}
		ctx2, canselFunc2 := context.WithTimeout(cDBS.ctx, 3*time.Second)
		defer canselFunc2()
		_, err := cDBS.db.ExecContext(ctx2, `INSERT INTO URLStorage(short_URL, long_URL) VALUES ($N, $N)`, shortURL, longURL)
		if err != nil {
			cDBS.logger.Error("insertURL method, inserting new row error", zap.Error(err))
			return err
		}
	}
	cDBS.logger.Info("this short url is already involved")
	return errors.New("this short url is already involved")
}

func (cDBS *ClientDBStruct) GetLongURL(shortURL string) (string, error) {
	ctx, canselFunc := context.WithTimeout(cDBS.ctx, 3*time.Second)
	defer canselFunc()

	row := cDBS.db.QueryRowContext(ctx, `SELECT long_URL FROM URLStorage WHERE short_URL=$N`, shortURL)
	var longURL string
	err := row.Scan(&longURL)
	if err != nil {
		cDBS.logger.Error("getLongURL metod, getting longURL error", zap.Error(err))
		return "", errors.New("getLongURL metod, getting longURL error:" + err.Error())
	}
	return longURL, nil
}
