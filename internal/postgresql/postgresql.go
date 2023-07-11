package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
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
	tm     time.Duration
}

func NewClientDBStruct(logger *zap.Logger, cf *config.Config, tm time.Duration) *ClientDBStruct {
	return &ClientDBStruct{
		db:     new(pgxpool.Pool),
		logger: logger,
		cf:     cf,
		tm:     tm,
	}
}

func (cDBS *ClientDBStruct) Create(parentCtx context.Context) error {
	cDBS.logger.Debug("cDBS.dsn: " + cDBS.cf.DBdsn)
	cf, err := pgxpool.ParseConfig(cDBS.cf.DBdsn)
	cf.MaxConns = 10
	cf.MaxConnIdleTime = 120 * time.Second
	cf.MaxConnLifetime = 360 * time.Second
	if err != nil {
		return errors.New("error while parsing db config, " + err.Error())
	}
	cDBS.db, err = pgxpool.NewWithConfig(parentCtx, cf)
	if err != nil {
		return errors.New("error while creatin db connection pool, " + err.Error())
	}

	ctx, canselCtx := context.WithTimeout(parentCtx, cDBS.tm)
	defer canselCtx()
	res, err := cDBS.db.Exec(ctx, `CREATE TABLE IF NOT EXISTS urlstorage("id" INTEGER, "user_id" TEXT, "short_url" TEXT, "long_url" TEXT, "is_deleted" BOOLEAN DEFAULT false,  CONSTRAINT uniq_key PRIMARY KEY("user_id", "long_url"))` /*UNIQUE(user_id, long_url))*/)
	if err != nil {
		return errors.New("creating db method, error while creating new table, " + err.Error())
	}
	cDBS.logger.Debug("table is successfuly created")
	rows := res.RowsAffected()
	cDBS.logger.Debug("Rows affected when creating table: ", zap.Int64("raws num", rows))
	return nil
}

func (cDBS *ClientDBStruct) Close() {
	cDBS.db.Close()
}

func (cDBS *ClientDBStruct) Ping(ctxPar context.Context) error {
	ctx, canselCtx := context.WithTimeout(ctxPar, cDBS.tm)
	defer canselCtx()
	err := cDBS.db.Ping(ctx)
	if err != nil {
		return errors.New("db ping error" + err.Error())
	}
	cDBS.logger.Debug("db ping success")
	return nil
}

func (cDBS *ClientDBStruct) PostShortURL(shortURL, longURL, userID string, uuid int32) (string, error) {
	cDBS.logger.Debug("new shrtURL is: " + shortURL)
	ctxPar := context.TODO()
	ctx, canselCtx := context.WithTimeout(ctxPar, cDBS.tm)
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
	cDBS.logger.Debug("what we insert: " + userID + " " + shortURL + " " + longURL)
	row := tx.QueryRow(ctx,
		`INSERT INTO urlstorage(id, user_id, short_url, long_url)
		 VALUES ($1, $2, $3, $4) 
		 ON CONFLICT 
		 ON CONSTRAINT uniq_key
		 DO UPDATE SET 
		 long_url=EXCLUDED.long_url
		 RETURNING (short_url)`, uuid, userID, shortURL, longURL)
	err = row.Scan(&shURL)
	cDBS.logger.Debug("returned shrtURL is: " + shURL)
	if err != nil {
		tx.Rollback(ctx)
		return "", errors.New("postShortURL db method, insert data error: " + err.Error())
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
	ctx, canselCtx := context.WithTimeout(ctxPar, cDBS.tm)
	defer canselCtx()
	cDBS.logger.Debug("shortURL for getting", zap.String("shtURL", shortURL))
	row := cDBS.db.QueryRow(ctx, `SELECT long_url, is_deleted  FROM urlstorage WHERE short_url=$1`, shortURL)
	var longURL string
	var isDeleted bool
	err := row.Scan(&longURL, &isDeleted)
	if err != nil {
		return "", errors.New("getLongURL metod, getting longURL error:" + err.Error())
	}
	if isDeleted {
		return "", nil
	}
	return longURL, nil
}

type DBRowStrct struct {
	ID       int
	ShortURL string
	LongURL  string
}

func (cDBS *ClientDBStruct) ReturnAllUserReq(ctxPar context.Context, userID string) (map[string]string, error) {
	ctx, canselCtx := context.WithTimeout(ctxPar, cDBS.tm)
	defer canselCtx()
	ans := make(map[string]string, 0)

	rows, err := cDBS.db.Query(ctx, `SELECT short_url, long_url FROM urlstorage WHERE user_id = $1`, userID)

	if err != nil {
		return nil, errors.New("getting all user requests error, " + err.Error())
	}
	defer rows.Close()
	var shrtURL, lngURL string
	for rows.Next() {
		err := rows.Scan(&shrtURL, &lngURL)
		if err != nil {
			cDBS.logger.Error("error while scaning user requests", zap.Error(err))
		}
		ans[lngURL] = shrtURL
	}
	return ans, nil
}

func (cDBS *ClientDBStruct) PostURLBatch(URLarr []DBRowStrct, userID string) ([]string, error) {
	ctxPar := context.TODO()
	ctx, canselCtx := context.WithTimeout(ctxPar, cDBS.tm)
	defer canselCtx()
	btch := new(pgx.Batch)
	for _, v := range URLarr {
		cDBS.logger.Debug("what we insert: " + userID + " " + v.ShortURL + " " + v.LongURL)
		btch.Queue(`INSERT INTO urlstorage(id, user_id, short_url, long_url)
		 VALUES ($1, $2, $3, $4) 
		 ON CONFLICT 
		 ON CONSTRAINT uniq_key
		 DO UPDATE SET 
		 long_url=EXCLUDED.long_url
		 RETURNING (short_url)`, v.ID, userID, v.ShortURL, v.LongURL)
	}
	retShrtURL := make([]string, 0)
	br := cDBS.db.SendBatch(ctx, btch)
	defer br.Close()
	shortid := ""
	for i := range URLarr {
		br.QueryRow().Scan(&shortid)
		if shortid != "" {
			retShrtURL = append(retShrtURL, shortid)
		} else {
			retShrtURL = append(retShrtURL, URLarr[i].ShortURL)
		}
	}
	return retShrtURL, nil
}

type URLsForDel struct {
	UserID   string
	ShortURL string
}

func (cDBS *ClientDBStruct) DeleteUserURLS(ctxPar context.Context, urlsArr []URLsForDel) (err error) {
	fmt.Println("start deleting")
	ctx, canselCtx := context.WithTimeout(ctxPar, cDBS.tm)
	defer canselCtx()
	btch := new(pgx.Batch)
	for _, s := range urlsArr {
		btch.Queue(`UPDATE urlstorage SET is_deleted = true WHERE user_id = $1 AND short_url = $2`, s.UserID, s.ShortURL)
	}
	btchRes := cDBS.db.SendBatch(ctx, btch)
	for i := range urlsArr {
		_, err = btchRes.Exec()
		if err != nil {
			cDBS.logger.Error("batch row err, row N is "+strconv.Itoa(i)+" /", zap.Error(err))
		}
	}
	fmt.Println("finish deleting")
	return
}
