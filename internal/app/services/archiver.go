package services

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

const filePerm = 0755

type zipService struct {
	archiveDir string
}

func NewZipService(archiveDir string) *zipService {
	os.MkdirAll(archiveDir, filePerm)
	return &zipService{archiveDir: archiveDir}
}

func (s *zipService) CreateZipArchive(files []string) (string, error) {
	archiveName := filepath.Join(s.archiveDir, uuid.NewString()+".zip")
	out, err := os.Create(archiveName)
	if err != nil {
		return "", err
	}
	defer out.Close()

	zipWriter := zip.NewWriter(out)
	defer zipWriter.Close()

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			continue
		}
		defer f.Close()

		w, err := zipWriter.Create(filepath.Base(file))
		if err != nil {
			continue
		}
		io.Copy(w, f)
	}

	return archiveName, nil
}
