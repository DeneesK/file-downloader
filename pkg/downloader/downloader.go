package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/DeneesK/file-downloader/pkg/validator"
	"github.com/google/uuid"
)

func DownloadFile(url string) (string, error) {
	if !(validator.IsValidURL(url)) {
		return "", fmt.Errorf("not valid url: %s", url)
	}
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error: %v", err)
	}
	defer resp.Body.Close()

	filename := filepath.Join(os.TempDir(), uuid.NewString()+filepath.Ext(url))
	out, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return filename, err
}
