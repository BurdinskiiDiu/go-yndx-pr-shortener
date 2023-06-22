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

/*
type ClientDB interface {
	Create() (*sql.DB, error)
	Ping(context.Context) error
	Close()
}*/
//////////////////////old part
/*type ClientDBStruct struct {
	db     *sql.DB
	logger *zap.Logger
	//ctx    context.Context
	cf *config.Config
}*/
/*
//func NewClientDBStruct( /*ctx context.Context, logger *zap.Logger, cf *config.Config) *ClientDBStruct {
	return &ClientDBStruct{
		db:     new(sql.DB),
		logger: logger,
		//ctx:    ctx,
		cf: cf,
	}
}
*/ /*
func (cDBS *ClientDBStruct) Create() error {
	var err error
	///*cDBS.logger.Info("cDBS.cf.StoreType: " + strconv.Itoa(cDBS.cf.StoreType))

	//if cDBS.cf.StoreType == 0 {
	//	return errors.New("db is not necessary")
	//}*/ /*
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
	ctxPar := context.TODO()
	ctx, cansel := context.WithTimeout( /*cDBS.ctx*/ /* ctxPar, 100*time.Second)
defer cansel()

res, err := cDBS.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS urlstorage("id" INTEGER, "short_url" TEXT, "long_url" TEXT, UNIQUE(long_url))`)

if err != nil {
	cDBS.logger.Error("creating db method, error while creating new table", zap.Error(err))
	return err
}
cDBS.logger.Info("table is successfuly created")
rows, err := res.RowsAffected()
if err != nil {
	cDBS.logger.Error("Error %s when getting rows affected", zap.Error(err))
	return err
}
cDBS.logger.Info("Rows affected when creating table: ", zap.Int64("raws num", rows))
///*ctx2, cansel2 := context.WithTimeout(cDBS.ctx, 100*time.Second)
//defer cansel2()
//_, err = cDBS.db.ExecContext(ctx2, `ALTER TABLE urlstorage ADD CONSTRAINT longurl_id UNIQUE (long_url)`)
//if err != nil {
//	cDBS.logger.Error("creating db method, error while creating unique addition", zap.Error(err))
//}*/
/*
	return nil
}*/

// /////////////////new part
type ClientDB interface {
	Create() (*sql.DB, error)
	Ping(context.Context) error
	Close()
}

// ////////////////////old part
type ClientDBStruct struct {
	db     *pgx.Conn
	logger *zap.Logger
	//ctx    context.Context
	cf *config.Config
}

func NewClientDBStruct( /*ctx context.Context,*/ logger *zap.Logger, cf *config.Config) *ClientDBStruct {
	return &ClientDBStruct{
		db:     new(pgx.Conn),
		logger: logger,
		//ctx:    ctx,
		cf: cf,
	}
}

func (cDBS *ClientDBStruct) Create(parentCtx context.Context) error {
	var err error
	///*cDBS.logger.Info("cDBS.cf.StoreType: " + strconv.Itoa(cDBS.cf.StoreType))

	//if cDBS.cf.StoreType == 0 {
	//	return errors.New("db is not necessary")
	//}*/
	cDBS.logger.Info("cDBS.dsn: " + cDBS.cf.DBdsn)
	//ctxPar := context.TODO()
	cDBS.db, err = pgx.Connect(parentCtx, cDBS.cf.DBdsn)
	if err != nil {
		cDBS.logger.Error("creating db method, error while creating new db", zap.Error(err))
		return err
	}
	cDBS.logger.Info("db is successfuly created")
	/*cDBS.db.SetMaxOpenConns(20)
	cDBS.db.SetMaxIdleConns(20)
	cDBS.db.SetConnMaxLifetime(time.Minute * 5)*/
	//ctxPar := context.TODO()
	ctx, cansel := context.WithTimeout( /*cDBS.ctx*/ parentCtx, 100*time.Second)
	defer cansel()
	//time.Sleep(10 * time.Second)
	res, err := cDBS.db.Exec(ctx, `CREATE TABLE IF NOT EXISTS urlstorage("id" INTEGER, "short_url" TEXT, "long_url" TEXT, UNIQUE(long_url))`)
	if err != nil {
		cDBS.logger.Error("creating db method, error while creating new table", zap.Error(err))
		return err
	}
	cDBS.logger.Info("table is successfuly created")
	rows := res.RowsAffected()
	cDBS.logger.Info("Rows affected when creating table: ", zap.Int64("raws num", rows))
	///*ctx2, cansel2 := context.WithTimeout(cDBS.ctx, 100*time.Second)
	//defer cansel2()
	//_, err = cDBS.db.ExecContext(ctx2, `ALTER TABLE urlstorage ADD CONSTRAINT longurl_id UNIQUE (long_url)`)
	//if err != nil {
	//	cDBS.logger.Error("creating db method, error while creating unique addition", zap.Error(err))
	//}*/
	return nil
}

// ////////////////////////////
// //////////////////////////
func (cDBS *ClientDBStruct) Close(parentCtx context.Context) {
	cDBS.db.Close(parentCtx)
}

func (cDBS *ClientDBStruct) Ping() error {
	ctxPar := context.TODO()
	ctx, canselFunc := context.WithTimeout( /*cDBS.ctx*/ ctxPar, 30*time.Second)
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
	fmt.Println("now we are here place1")
	cDBS.logger.Info("new shrtURL is: " + shortURL)
	ctxPar := context.TODO()
	ctx, canselCtx := context.WithTimeout( /*cDBS.ctx*/ ctxPar, 1*time.Minute)
	defer canselCtx()
	var shURL, lnURL string
	err := cDBS.db.QueryRow(ctx, `SELECT long_url FROM urlstorage WHERE short_url=$1`, shortURL).Scan(&lnURL)
	if err != nil {
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
	fmt.Println("now we are here place2")
	cDBS.logger.Info("returned shrtURL is: " + shURL)
	if err != nil {
		cDBS.logger.Error("insert data error", zap.Error(err))
		//if shURL == shortURL {
		//err := errors.New("shortURL is already exist")
		//	return "", err

		return "", err
	}
	if shURL != shortURL && shURL != "" {
		err := errors.New("longURL is already exist")
		return shURL, err
	}
	return shURL, nil
}

/*
	if
	//row := cDBS.db.QueryRowContext(ctx1, `SELECT long_url FROM urlstorage WHERE short_url=$1`, shortURL)
	var checkURL string
	err := row.Scan(&checkURL)
	cDBS.logger.Info("this short url from request " + shortURL)
	cDBS.logger.Info("checked url from db " + checkURL)
	if err != nil {
		if err != sql.ErrNoRows {
			cDBS.logger.Error("insertURL method, error while scaning", zap.Error(err))
			cDBS.logger.Info("gotted checkURL is" + checkURL)
			return err
		}
		cDBS.logger.Info("checking short url, it is not exist, shortURL: " + checkURL)
		ctx2, canselFunc2 := context.WithTimeout(ctxPar, 1*time.Minute)
		defer canselFunc2()
		_, err := cDBS.db.ExecContext(ctx2, `INSERT INTO urlstorage(id, short_url, long_url) VALUES ($1, $2, $3)`, uuid, shortURL, longURL)
		if err != nil {
			cDBS.logger.Error("insertURL method, inserting new row error", zap.Error(err))
			if strings.Contains(err.Error(), "duplicate key value violates") {
				cDBS.logger.Info("catch this error!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
			}
			return err
		}
		return nil
	}
	cDBS.logger.Info("this short url is already involved")
	return errors.New("this short url is already involved")*/

func (cDBS *ClientDBStruct) GetLongURL(shortURL string) (string, error) {
	ctxPar := context.TODO()
	ctx, canselFunc := context.WithTimeout( /*cDBS.ctx*/ ctxPar, 1*time.Minute)
	defer canselFunc()

	row := cDBS.db.QueryRow(ctx, `SELECT long_url FROM urlstorage WHERE short_url=$1`, shortURL)
	var longURL string
	err := row.Scan(&longURL)
	if err != nil {
		cDBS.logger.Error("getLongURL metod, getting longURL error", zap.Error(err))
		cDBS.logger.Info("getLongURL metod, getting longURL error" + longURL)
		return "", errors.New("getLongURL metod, getting longURL error:" + err.Error())
	}
	return longURL, nil
}

/*
func (cDBS *ClientDBStruct) GetShortURL(longURL string) (string, error) {
	ctxPar := context.TODO()
	ctx, canselFunc := context.WithTimeout( /*cDBS.ctx*/ /*ctxPar, 1*time.Minute)
	defer canselFunc()

	row := cDBS.db.QueryRowContext(ctx, `SELECT short_url FROM urlstorage WHERE long_url=$1`, longURL)
	var shortURL string
	err := row.Scan(&shortURL)
	if err != nil {
		cDBS.logger.Error("getLongURL metod, getting longURL error", zap.Error(err))
		cDBS.logger.Info("getLongURL metod, getting longURL error" + shortURL)
		return "", errors.New("getLongURL metod, getting longURL error:" + err.Error())
	}
	return shortURL, nil
}
*/

type DBRowStrct struct {
	ID       int
	ShortURL string
	LongURL  string
}

func (cDBS *ClientDBStruct) PostURLBatch(URLarr []DBRowStrct) ([]string, error) {
	ctxPar := context.TODO()
	ctx, canselCtx := context.WithTimeout( /*cDBS.ctx*/ ctxPar, 1*time.Minute)
	defer canselCtx()
	btch := new(pgx.Batch)
	/*query := `INSERT INTO urlstorage(id, short_url, long_url) VALUES(@ID, @shortURL, @longURL)`
	for _, v := range URLarr {
		args := pgx.NamedArgs{
			"ID":       v.ID,
			"shortURL": v.ShortURL,
			"longURL":  v.LongURL,
		}
		btch.Queue(query, args)
	}*/

	for _, v := range URLarr {
		btch.Queue( /*`INSERT INTO urlstorage(id, short_url, long_url)
			VALUES ($1, $2, $3)`*/`INSERT INTO urlstorage(id, short_url, long_url)
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
		//cT, err := br.Exec()
		br.QueryRow().Scan(&shortid)
		if shortid != "" {
			retShrtURL = append(retShrtURL, shortid)
		} else {
			retShrtURL = append(retShrtURL, URLarr[i].ShortURL)
		}
		fmt.Println("new short url: " + v.ShortURL)
		fmt.Println("returned short url: " + shortid)
		/*if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				cDBS.logger.Info("short url is already exist", zap.String("short_url:", row.ShortURL))
				continue
			}
			return errors.New("unable to insert row:" + err.Error())
		}*/
	}
	br.Close()
	return retShrtURL, nil
	//return nil
}
