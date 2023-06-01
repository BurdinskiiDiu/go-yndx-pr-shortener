package store

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	filestore "github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/fileStore"
	"go.uber.org/zap"
)

type URLStorage struct {
	URLStr  map[string]string
	mutex   *sync.Mutex
	uuid    int
	fileInf *filestore.FileExst
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		URLStr:  make(map[string]string),
		mutex:   new(sync.Mutex),
		uuid:    0,
		fileInf: new(filestore.FileExst),
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
	logger.Info("storefile addr from post req", zap.String("path", uS.fileInf.FileName))
	err := uS.FileFilling(shortURL, longURL, logger)
	if err != nil {
		logger.Info("file filling error")
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
	uS.fileInf.FileName = cf.FileStorePath

	if _, err := os.Stat(uS.fileInf.FileName); err != nil {
		if os.IsNotExist(err) {
			logger.Info("store file is not exist. creating file")
		}
	} else {
		uS.fileInf.Existed = true
	}

	logger.Info("storefile addr from createfile", zap.String("path", uS.fileInf.FileName))

	if !uS.fileInf.Existed {
		file, err := os.Create(cf.FileStorePath)
		if err != nil {
			//return fmt.Errorf("creating store_file err: %w", err)
			return err
		}
		file.Close()
	}

	file, err := os.OpenFile(uS.fileInf.FileName, os.O_RDONLY, 0777)
	if err != nil {
		//return fmt.Errorf("open store_file error: %w", err)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	urlDataStr := new(URLDataStruct)
	var raw string
	for scanner.Scan() {
		raw = scanner.Text()
		err := json.Unmarshal([]byte(raw), urlDataStr)
		if err != nil {
			//return fmt.Errorf("unmarhalling store_file error: %w", err)
			return err
		}
		uS.URLStr[urlDataStr.ShrtURL] = urlDataStr.LngURL
	}
	if urlDataStr.UUID != "" {
		uS.uuid, err = strconv.Atoi(urlDataStr.UUID)
		if err != nil {
			//logger.Info("gettitng last uuid error")
			return err
		}
	}
	return nil
}

func (uS *URLStorage) FileFilling(shrtURL, lngURL string, logger *zap.Logger) error {
	if _, err := os.Stat(uS.fileInf.FileName); err != nil {
		if os.IsNotExist(err) {
			logger.Info("there are no backUpFile to fill with new data")
			return errors.New("filling filestore error")
		}
	}
	logger.Info("storefile addr from fillins method", zap.String("path", uS.fileInf.FileName))
	file, err := os.OpenFile(uS.fileInf.FileName, os.O_RDWR|os.O_APPEND, 0777)
	if err != nil {
		return err
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
		return err
	}
	if _, err := writer.Write(raw); err != nil {
		return err
	}
	if err := writer.WriteByte('\n'); err != nil {
		return err
	}
	return writer.Flush()
}
