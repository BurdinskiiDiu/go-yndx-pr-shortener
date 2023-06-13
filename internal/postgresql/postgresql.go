package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

type ClientDB interface {
	Create() (*sql.DB, error)
	Ping(context.Context) error
	Close()
}

type ClientDBStruct struct {
	db     *sql.DB
	logger *zap.Logger
	ctx    context.Context
	cf     *config.Config
}

func NewClientDBStruct(ctx context.Context, logger *zap.Logger, cf *config.Config) *ClientDBStruct {
	return &ClientDBStruct{
		db:     new(sql.DB),
		logger: logger,
		ctx:    ctx,
		cf:     cf,
	}
}

// ///!!!!!!!!!!!!!!!!!!убрать create, усли это не нужно из конфига
func (cDBS *ClientDBStruct) Create() error {
	var err error
	cDBS.logger.Info("cDBS.cf.StoreType: " + strconv.Itoa(cDBS.cf.StoreType))

	if cDBS.cf.StoreType == 0 {
		return errors.New("db is not necessary")
	}
	cDBS.logger.Info("cDBS.dsn: " + cDBS.cf.DBdsn)
	cDBS.db, err = sql.Open("pgx", cDBS.cf.DBdsn)
	if err != nil {
		cDBS.logger.Error("creating db method, error while creating new db", zap.Error(err))
		return err
	}
	cDBS.logger.Info("db is successfuly created")
	cDBS.db.SetMaxOpenConns(20)
	cDBS.db.SetMaxIdleConns(20)
	cDBS.db.SetConnMaxLifetime(time.Minute * 5)

	ctx, cansel := context.WithTimeout(cDBS.ctx, 5*time.Second)
	defer cansel()
	tx, err := cDBS.db.BeginTx(ctx, nil)
	if err != nil {
		cDBS.logger.Error("creating db method, error while creating new transaction", zap.Error(err))
	}
	defer tx.Rollback()

	//res, err := cDBS.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS urlstorage("id" INTEGER, "short_url" TEXT, "long_url" TEXT)`)
	res, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS urlstorage("id" INTEGER, "short_url" TEXT, "long_url" TEXT)`)
	if err != nil {
		cDBS.logger.Error("creating db method, error while creating new table", zap.Error(err))
		return err
	}

	//_, err = cDBS.db.Exec(`CREATE UNIQUE INDEX longurl_idx ON urlstorage (long_url)`)
	_, err = tx.Exec(`CREATE UNIQUE INDEX long_url ON urlstorage (long_url)`)
	if err != nil {
		cDBS.logger.Error("creating db method, error while creating UNIQUE INDEX", zap.Error(err))
		//return err
	}
	cDBS.logger.Info("table is successfuly created")

	rows, err := res.RowsAffected()
	if err != nil {
		cDBS.logger.Error("Error %s when getting rows affected", zap.Error(err))
		return err
	}
	cDBS.logger.Info("Rows affected when creating table: ", zap.Int64("raws num", rows))
	tx.Commit()
	return nil

}

func (cDBS *ClientDBStruct) Close() {
	cDBS.db.Close()
}

func (cDBS *ClientDBStruct) Ping() error {
	ctx, canselFunc := context.WithTimeout(cDBS.ctx, 30*time.Second)
	defer canselFunc()

	err := cDBS.db.PingContext(ctx)
	if err != nil {
		cDBS.logger.Error("db ping error")
		return err
	}
	cDBS.logger.Info("db ping success")
	return nil
}

/*
type modifErr struct {
	Err error
	Url string
}

func NewmodifErr(url string, err error) error {
	return &modifErr{
		Err: err,
		Url: url,
	}
}

func (mE *modifErr) Error() string {
	return mE.Err.Error()
}*/

func (cDBS *ClientDBStruct) PostShortURL(shortURL, longURL string, uuid int32) (string, error) {
	ctx1, canselFunc1 := context.WithTimeout(cDBS.ctx, 1*time.Minute)
	defer canselFunc1()
	row := cDBS.db.QueryRowContext(ctx1, `SELECT long_url FROM urlstorage WHERE short_url=$1`, shortURL)
	var checkURL string
	err := row.Scan(&checkURL)
	cDBS.logger.Info("this short url from request " + shortURL)
	cDBS.logger.Info("checked url from db " + checkURL)
	if err != nil {
		if /*err != sql.ErrNoRows*/ !errors.Is(err, sql.ErrNoRows) {
			cDBS.logger.Error("postShortURL to db method, error while scaning", zap.Error(err))
			cDBS.logger.Info("gotted checkURL is" + checkURL)
			return "", err
		}
		cDBS.logger.Info("checking short url, it is not exist, shortURL: " + checkURL)
		ctx2, canselFunc2 := context.WithTimeout(cDBS.ctx, 1*time.Minute)
		defer canselFunc2()

		_, err := cDBS.db.ExecContext(ctx2, `INSERT INTO urlstorage(id, short_url, long_url) VALUES ($1, $2, $3) ON CONFLICT (long_url) DO NOTHING`, uuid, shortURL, longURL)
		if err != nil {
			var srErr *pq.Error
			if errors.As(err, &srErr) {
				if srErr.Code == pgerrcode.UniqueViolation {
					ctx3, canselFunc3 := context.WithTimeout(cDBS.ctx, 1*time.Minute)
					defer canselFunc3()
					row = cDBS.db.QueryRowContext(ctx3, `SELECT short_url FROM urlstorage WHERE long_url=$1`, longURL)
					var url string
					err := row.Scan(&url)
					if err != nil {
						cDBS.logger.Error("postShortURL to db method, error while scaning", zap.Error(err))
						return "", err
					}
					return url, err
				}
			}
			cDBS.logger.Error("insertURL method, inserting new row error", zap.Error(err))
			return "", err
		}
		return "", nil
	}
	cDBS.logger.Info("this short url is already involved")
	return "", errors.New("this short url is already involved")
}

func (cDBS *ClientDBStruct) GetLongURL(shortURL string) (string, error) {
	ctx, canselFunc := context.WithTimeout(cDBS.ctx, 1*time.Minute)
	defer canselFunc()

	row := cDBS.db.QueryRowContext(ctx, `SELECT long_url FROM urlstorage WHERE short_url=$1`, shortURL)
	var longURL string
	err := row.Scan(&longURL)
	if err != nil {
		cDBS.logger.Error("getLongURL metod, getting longURL error", zap.Error(err))
		cDBS.logger.Info("getLongURL metod, getting longURL error" + longURL)
		return "", errors.New("getLongURL metod, getting longURL error:" + err.Error())
	}
	return longURL, nil
}
