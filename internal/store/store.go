package store

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"go.uber.org/zap"
)

type URLStorage struct {
	URLStr     map[string]string
	mutex      *sync.Mutex
	uuid       int
	dbFileName string
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		URLStr:     make(map[string]string),
		mutex:      new(sync.Mutex),
		uuid:       0,
		dbFileName: "",
	}
}

func (uS *URLStorage) PostShortURL(shortURL, longURL string, logger *zap.Logger) error {
	uS.mutex.Lock()
	defer uS.mutex.Unlock()
	_, ok := uS.URLStr[shortURL]
	if ok {
		return errors.New("this short url is already involved")
	}
	uS.URLStr[shortURL] = longURL
	logger.Info("storefile addr from post req", zap.String("path", uS.dbFileName))
	err := uS.FileFilling(shortURL, longURL, logger)
	if err != nil {
		logger.Info("file filling error")
		return errors.New("file filling error")
	}
	return nil
}

func (uS *URLStorage) GetLongURL(shrtURL string) (string, error) {
	uS.mutex.Lock()
	defer uS.mutex.Unlock()
	lngURL, ok := uS.URLStr[shrtURL]
	if !ok {
		return "", errors.New("wrong short url")
	}
	return lngURL, nil
}

type URLDataStruct struct {
	UUID    string `json:"uuid"`
	ShrtURL string `json:"short_url"`
	LngURL  string `json:"original_url"`
}

/*
	func (uS *URLStorage) GetStoreBackup(fE *filestore.FileExst, logger *zap.Logger) {
		uS.fileInf = fE
		logger.Info("storefile addr from createfile", zap.String("path", uS.fileInf.FileName))
		if !fE.Existed {
			logger.Info("there are new empty backUpFile")
			return
		}

		file, err := os.OpenFile(fE.FileName, os.O_RDONLY, 0777)
		if err != nil {
			logger.Info("open storeFile error")
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		urlDataStr := new(URLDataStruct)
		var raw string
		for scanner.Scan() {
			raw = scanner.Text()
			err := json.Unmarshal([]byte(raw), urlDataStr)
			if err != nil {
				logger.Info("unmarhalling storeFile error")
				return
			}
			uS.URLStr[urlDataStr.ShrtURL] = urlDataStr.LngURL
		}
		uS.uuid, err = strconv.Atoi(urlDataStr.UUID)

		if err != nil {
			logger.Info("gettitng last uuid error")
			return
		}
	}
*/
func (uS *URLStorage) GetStoreBackup(cf *config.Config, logger *zap.Logger) error {
	uS.dbFileName = cf.FileStorePath
	/*
		if _, err := os.Stat(uS.fileInf.FileName); err != nil {
			if os.IsNotExist(err) {
				logger.Info("store file is not exist. creating file")
			}
		} else {
			uS.fileInf.Existed = true
		}*/

	logger.Info("storefile addr from createfile", zap.String("path", uS.dbFileName))
	/*
		if !uS.fileInf.Existed {
			file, err := os.Create(cf.FileStorePath)
			if err != nil {
				logger.Info("creating store_file err")
				//return fmt.Errorf("creating store_file err: %w", err)
				return err
			}
			uS.fileInf.Existed = true
			file.Close()
		}*/

	file, err := os.OpenFile(uS.dbFileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		logger.Info("open storeFile error")
		return fmt.Errorf("open store_file error: %w", err)

	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	urlDataStr := new(URLDataStruct)
	var raw string
	for scanner.Scan() {
		raw = scanner.Text()
		err := json.Unmarshal([]byte(raw), urlDataStr)
		if err != nil {
			logger.Info("unmarhalling store_file error")
			//return fmt.Errorf("unmarhalling store_file error: %w", err)
			return err
		}
		uS.URLStr[urlDataStr.ShrtURL] = urlDataStr.LngURL
	}
	if urlDataStr.UUID != "" {
		uS.uuid, err = strconv.Atoi(urlDataStr.UUID)
		if err != nil {
			logger.Info("gettitng last uuid error, file is damaged")
		}
	}
	return nil
}

func (uS *URLStorage) FileFilling(shrtURL, lngURL string, logger *zap.Logger) error {
	/*if _, err := os.Stat(uS.fileInf.FileName); err != nil {
		if os.IsNotExist(err) {
			logger.Info("there are no backUpFile to fill with new data")
			return errors.New("filling filestore error")
		}
	}*/
	logger.Info("storefile addr from fillins method", zap.String("path", uS.dbFileName))
	file, err := os.OpenFile(uS.dbFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		logger.Error("open db file error")
		return fmt.Errorf("open db file error: %w", err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	var raw []byte
	urlDataStr := new(URLDataStruct)
	uS.uuid++
	urlDataStr.UUID = strconv.Itoa(uS.uuid)
	urlDataStr.ShrtURL = shrtURL
	urlDataStr.LngURL = lngURL
	raw, err = json.Marshal(urlDataStr)
	if err != nil {
		return fmt.Errorf("marshalling data to db file error: %w", err)
	}
	if _, err := writer.Write(raw); err != nil {
		return fmt.Errorf("writing data to db file error: %w", err)
	}
	if err := writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("making indent in db file error: %w", err)
	}
	return writer.Flush()
}
