package filestore

/*
import (
	"os"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/config"
	"go.uber.org/zap"
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

func CreateFileStore(cf config.Config, logger *zap.Logger) *FileExst {
	fE := NewFileExist()
	fE.FileName = cf.FileStorePath
	if _, err := os.Stat(cf.FileStorePath); err != nil {
		if os.IsNotExist(err) {
			logger.Info("store file is not exist. creating file")
		}
	} else {
		fE.Existed = true
		return fE
	}

	if !fE.Existed {
		file, err := os.Create(cf.FileStorePath)
		if err != nil {
			logger.Info("creating store file err: " + err.Error())
			return nil
		}
		defer file.Close()
	}

	return fE
}
*/
