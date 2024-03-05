package mediaworker

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type MediaWorker interface {
	SaveFile(img image.Image) (string, error)
	LoadFile(path string) (image.Image, error)
}

type SimpleFileWorker struct {
	fileDir string
	lastInd int
}

func NewMediaWorker(fileDir string) MediaWorker {
	mxInd := -1
	err := filepath.Walk(fileDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() == false {
			inds := strings.Split(info.Name(), ".")[0]
			ind, _ := strconv.Atoi(inds)
			if ind > mxInd {
				mxInd = ind
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal("cant init file saver")
	}
	return &SimpleFileWorker{fileDir: fileDir, lastInd: mxInd + 1}
}

func (s *SimpleFileWorker) SaveFile(img image.Image) (string, error) {
	path := s.fileDir + string(os.PathSeparator) + strconv.Itoa(s.lastInd) + ".jpeg"
	file, err := os.Create(path)
	defer func() { _ = file.Close() }()
	if err != nil {
		return "", err
	}
	err = jpeg.Encode(file, img, nil)
	if err != nil {
		return "", err
	}
	s.lastInd++
	return path, nil
}

func (s *SimpleFileWorker) LoadFile(path string) (image.Image, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	defer func() { _ = file.Close() }()
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, err
}

func ImageToString(img image.Image) string {
	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, nil)
	if err != nil {
		return ""
	}
	byteArr := buf.Bytes()
	return base64.StdEncoding.EncodeToString(byteArr)
}
