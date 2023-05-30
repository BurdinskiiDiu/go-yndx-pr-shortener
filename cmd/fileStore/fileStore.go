package filestore

import (
	"os"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/cmd/config"
	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/logger"
)

type FileExst struct {
	FileName string
	Existed  bool
}

func NewFileExist() *FileExst {
	return &FileExst{
		FileName: "",
		Existed:  false,
	}
}

func CreateFileStore(cf config.Config) *FileExst {
	fE := NewFileExist()

	if _, err := os.Stat(cf.FileStorePath); err != nil {
		if os.IsNotExist(err) {
			logger.Log.Info("store file is not exist. creating file")
		}
	} else {
		fE.Existed = true
		return fE
	}

	if !fE.Existed {
		file, err := os.Create(cf.FileStorePath)
		if err != nil {
			logger.Log.Info("creating store file err: " + err.Error())
			return nil
		}
		fE.FileName = cf.FileStorePath
		defer file.Close()
	}

	return fE
}
