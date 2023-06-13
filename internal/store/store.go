package store

import (
	"errors"
	"sync"

	"go.uber.org/zap"
)

type URLStorage struct {
	urlStr     map[string]string
	mutex      *sync.Mutex
	uuid       int
	dbFileName string
	logger     *zap.Logger
}

func NewURLStorageTest(us *map[string]string, logger *zap.Logger) *URLStorage {
	return &URLStorage{
		urlStr:     *us,
		mutex:      new(sync.Mutex),
		uuid:       0,
		dbFileName: "",
		logger:     logger,
	}
}

func NewURLStorage(logger *zap.Logger) *URLStorage {
	return &URLStorage{
		urlStr:     make(map[string]string),
		mutex:      new(sync.Mutex),
		uuid:       0,
		dbFileName: "",
		logger:     logger,
	}
}

func (uS *URLStorage) PostShortURL(shortURL, longURL string, uuid int32) error {
	uS.mutex.Lock()
	defer uS.mutex.Unlock()
	_, ok := uS.urlStr[shortURL]
	if ok {
		uS.logger.Info("shortURL: " + shortURL + " and longURL: " + uS.urlStr[shortURL])
		return errors.New("this short url is already involved")
	}
	uS.urlStr[shortURL] = longURL
	uS.logger.Debug("storefile addr from post req", zap.String("path", uS.dbFileName))
	/*err := uS.FileFilling(shortURL, longURL)
	if err != nil {
		uS.logger.Error("file filling error")
	}*/
	return nil
}

func (uS *URLStorage) GetLongURL(shrtURL string) (string, error) {
	uS.mutex.Lock()
	defer uS.mutex.Unlock()
	lngURL, ok := uS.urlStr[shrtURL]
	if !ok {
		return "", errors.New("wrong short url")
	}
	return lngURL, nil
}

func (uS *URLStorage) Ping() error {
	return nil
}

/*
type URLDataStruct struct {
	UUID    string `json:"uuid"`
	ShrtURL string `json:"short_url"`
	LngURL  string `json:"original_url"`
}




func (uS *URLStorage) GetStoreBackup(cf *config.Config) error {
	uS.dbFileName = cf.FileStorePath

	uS.logger.Debug("storefile addr from createfile", zap.String("path", uS.dbFileName))

	file, err := os.OpenFile(uS.dbFileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		uS.logger.Error("open storeFile error")
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
			uS.logger.Error("unmarhalling store_file error")
			return err
		}
		uS.urlStr[urlDataStr.ShrtURL] = urlDataStr.LngURL
	}
	if urlDataStr.UUID != "" {
		uS.uuid, err = strconv.Atoi(urlDataStr.UUID)
		if err != nil {
			uS.logger.Error("gettitng last uuid error, file is damaged")
		}
	}
	return nil
}

func (uS *URLStorage) FileFilling(shrtURL, lngURL string) error {
	uS.logger.Debug("storefile addr from fillins method", zap.String("path", uS.dbFileName))
	file, err := os.OpenFile(uS.dbFileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		uS.logger.Error("open db file error")
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
*/
