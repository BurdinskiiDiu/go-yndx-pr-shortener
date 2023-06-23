package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type ClientDB interface {
	Create() (*sql.DB, error)
	Ping(context.Context) error
	Close()
}

type ClientDBStruct struct {
	db     *pgx.Conn
	logger *zap.Logger
	cf     *config.Config
}

func NewClientDBStruct(logger *zap.Logger, cf *config.Config) *ClientDBStruct {
	return &ClientDBStruct{
		db:     new(pgx.Conn),
		logger: logger,
		cf:     cf,
	}
}

func (cDBS *ClientDBStruct) Create(parentCtx context.Context) error {
	var err error
	cDBS.logger.Info("cDBS.dsn: " + cDBS.cf.DBdsn)
	cDBS.db, err = pgx.Connect(parentCtx, cDBS.cf.DBdsn)
	if err != nil {
		cDBS.logger.Error("creating db method, error while creating new db", zap.Error(err))
		return err
	}
	cDBS.logger.Info("db is successfuly created")

	ctx, cansel := context.WithTimeout(parentCtx, 100*time.Second)
	defer cansel()
	res, err := cDBS.db.Exec(ctx, `CREATE TABLE IF NOT EXISTS urlstorage("id" INTEGER, "short_url" TEXT, "long_url" TEXT, UNIQUE(long_url))`)
	if err != nil {
		cDBS.logger.Error("creating db method, error while creating new table", zap.Error(err))
		return err
	}
	cDBS.logger.Info("table is successfuly created")
	rows := res.RowsAffected()
	cDBS.logger.Info("Rows affected when creating table: ", zap.Int64("raws num", rows))
	return nil
}

func (cDBS *ClientDBStruct) Close(parentCtx context.Context) {
	cDBS.db.Close(parentCtx)
}

func (cDBS *ClientDBStruct) Ping() error {
	ctxPar := context.TODO()
	ctx, canselFunc := context.WithTimeout(ctxPar, 30*time.Second)
	defer canselFunc()
	err := cDBS.db.Ping(ctx)
	if err != nil {
		cDBS.logger.Error("db ping error")
		return err
	}
	cDBS.logger.Info("db ping success")
	return nil
}

func (cDBS *ClientDBStruct) PostShortURL(shortURL, longURL string, uuid int32) (string, error) {
	cDBS.logger.Info("new shrtURL is: " + shortURL)
	ctxPar := context.TODO()
	ctx, canselCtx := context.WithTimeout(ctxPar, 1*time.Minute)
	defer canselCtx()
	var shURL, lnURL string
	//var pgErr *pgconn.PgError
	err := cDBS.db.QueryRow(ctx, `SELECT long_url FROM urlstorage WHERE short_url=$1`, shortURL).Scan(&lnURL)
	if err != nil {
		//if !(errors.As(err, &pgErr) && pgErr.Code == pgerrcode.NoDataFound) {
		//	return "", errors.New("postShortURL db method, err while selecting short url: " + err.Error())
		//}
		if !strings.Contains(err.Error(), "no rows in result set") {
			return "", errors.New("postShortURL db method, err while selecting short url: " + err.Error())
		}
	}

	if lnURL != "" {
		return "", errors.New("shortURL is already exist")
	}

	row := cDBS.db.QueryRow(ctx,
		`INSERT INTO urlstorage(id, short_url, long_url)
		 VALUES ($1, $2, $3) 
		 ON CONFLICT(long_url) 
		 DO UPDATE SET 
		 long_url=EXCLUDED.long_url
		 RETURNING (short_url)`, uuid, shortURL, longURL)
	err = row.Scan(&shURL)
	cDBS.logger.Info("returned shrtURL is: " + shURL)
	if err != nil {
		cDBS.logger.Error("insert data error", zap.Error(err))
		return "", err
	}
	if shURL != shortURL && shURL != "" {
		err := errors.New("longURL is already exist")
		return shURL, err
	}
	return shURL, nil
}

func (cDBS *ClientDBStruct) GetLongURL(shortURL string) (string, error) {
	ctxPar := context.TODO()
	ctx, canselCtx := context.WithTimeout(ctxPar, 1*time.Minute)
	defer canselCtx()

	row := cDBS.db.QueryRow(ctx, `SELECT long_url FROM urlstorage WHERE short_url=$1`, shortURL)
	var longURL string
	err := row.Scan(&longURL)
	if err != nil {
		cDBS.logger.Info("getLongURL metod, getting longURL error" + longURL)
		return "", errors.New("getLongURL metod, getting longURL error:" + err.Error())
	}
	return longURL, nil
}

type DBRowStrct struct {
	ID       int
	ShortURL string
	LongURL  string
}

func (cDBS *ClientDBStruct) PostURLBatch(URLarr []DBRowStrct) ([]string, error) {
	ctxPar := context.TODO()
	ctx, canselCtx := context.WithTimeout(ctxPar, 1*time.Minute)
	defer canselCtx()
	btch := new(pgx.Batch)
	for _, v := range URLarr {
		btch.Queue(`INSERT INTO urlstorage(id, short_url, long_url)
		 VALUES ($1, $2, $3) 
		 ON CONFLICT(long_url) 
		 DO UPDATE SET 
		 long_url=EXCLUDED.long_url
		 RETURNING (short_url)`, v.ID, v.ShortURL, v.LongURL)
	}
	retShrtURL := make([]string, 0)
	br := cDBS.db.SendBatch(ctx, btch)
	defer br.Close()
	shortid := ""
	for i, v := range URLarr {
		br.QueryRow().Scan(&shortid)
		if shortid != "" {
			retShrtURL = append(retShrtURL, shortid)
		} else {
			retShrtURL = append(retShrtURL, URLarr[i].ShortURL)
		}
		fmt.Println("new short url: " + v.ShortURL)
		fmt.Println("returned short url: " + shortid)

	}
	return retShrtURL, nil
}
