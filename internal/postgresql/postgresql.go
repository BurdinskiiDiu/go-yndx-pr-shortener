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
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type ClientDB interface {
	Create() (*sql.DB, error)
	Ping(context.Context) error
	Close()
}

type ClientDBStruct struct {
	db     *pgxpool.Pool
	logger *zap.Logger
	cf     *config.Config
}

func NewClientDBStruct(logger *zap.Logger, cf *config.Config) *ClientDBStruct {
	return &ClientDBStruct{
		db:     new(pgxpool.Pool),
		logger: logger,
		cf:     cf,
	}
}

func (cDBS *ClientDBStruct) Create(parentCtx context.Context) error {
	cDBS.logger.Info("cDBS.dsn: " + cDBS.cf.DBdsn)
	cf, err := pgxpool.ParseConfig(cDBS.cf.DBdsn)
	cf.MaxConns = 10
	cf.MaxConnIdleTime = 60 * time.Second
	cf.MaxConnLifetime = 360 * time.Second
	if err != nil {
		cDBS.logger.Error("error while parsing db config", zap.Error(err))
		return err
	}
	cDBS.db, err = pgxpool.NewWithConfig(parentCtx, cf)
	if err != nil {
		cDBS.logger.Error("error while creatin db connection pool", zap.Error(err))
		return err
	}

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

func (cDBS *ClientDBStruct) Close() {
	cDBS.db.Close()
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
	tx, err := cDBS.db.Begin(ctx)
	if err != nil {
		return "", errors.New("postShortURL db method, err while creating transaction: " + err.Error())
	}

	err = tx.QueryRow(ctx, `SELECT long_url FROM urlstorage WHERE short_url=$1`, shortURL).Scan(&lnURL)
	if err != nil {

		///попытка обработать ошибку
		/*	cDBS.logger.Info("our error " + err.Error())
			if errors.As(err, &pgErr) {
				if errors.Is(pgErr, sql.ErrNoRows) {
					cDBS.logger.Info("shortURL is already exist")
				} else {
					fmt.Println("wrong type of err too")
					return "", errors.New("postShortURL db method, err while selecting short url: " + err.Error())
				}
			} else {
				fmt.Println("wrong type of err")
				return "", errors.New("postShortURL db method, err while selecting short url: " + err.Error())
			}*/
		if !strings.Contains(err.Error(), "no rows in result set") {
			return "", errors.New("postShortURL db method, err while selecting short url: " + err.Error())
		}
	}

	if lnURL != "" {
		return "", errors.New("shortURL is already exist")
	}

	row := tx.QueryRow(ctx,
		`INSERT INTO urlstorage(id, short_url, long_url)
		 VALUES ($1, $2, $3) 
		 ON CONFLICT(long_url) 
		 DO UPDATE SET 
		 long_url=EXCLUDED.long_url
		 RETURNING (short_url)`, uuid, shortURL, longURL)
	err = row.Scan(&shURL)
	cDBS.logger.Debug("returned shrtURL is: " + shURL)
	if err != nil {
		cDBS.logger.Error("insert data error", zap.Error(err))
		tx.Rollback(ctx)
		return "", err
	}
	if shURL != shortURL && shURL != "" {
		err := errors.New("longURL is already exist")
		tx.Commit(ctx)
		return shURL, err
	}
	tx.Commit(ctx)
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

/*
func (cDBS *ClientDBStruct) PrintlAllDB() {
	ctxPar := context.TODO()
	ctx, canselCtx := context.WithTimeout(ctxPar, 1*time.Minute)
	defer canselCtx()
	rows, err := cDBS.db.Query(ctx, `SELECT * FROM urlstorage`)
	if err != nil {
		cDBS.logger.Error("error while printing all db", zap.Error(err))
	}
	defer rows.Close()
	var dataRow DBRowStrct
	for rows.Next() {
		err := rows.Scan(&dataRow.ID, &dataRow.ShortURL, &dataRow.LongURL)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Println("Id: " + strconv.Itoa(dataRow.ID) + ", shkrtURL is: " + dataRow.ShortURL + " , longURL is: " + dataRow.LongURL)
	}
}*/
