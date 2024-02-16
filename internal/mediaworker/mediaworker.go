package mediaworker

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"os"
	"strconv"
)

type MediaWorker interface {
	SaveFile(img image.Image) (string, error)
	LoadFile(path string) (image.Image, error)
}

type SimpleFileWorker struct {
	fileDir string
	lastInd int
}

func NewMediaWorker(fileDir string, lastInd int) MediaWorker {
	return &SimpleFileWorker{fileDir: fileDir, lastInd: lastInd}
}

func (s SimpleFileWorker) SaveFile(img image.Image) (string, error) {
	path := s.fileDir + string(os.PathSeparator) + strconv.Itoa(s.lastInd+1) + ".jpeg"
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

func (s SimpleFileWorker) LoadFile(path string) (image.Image, error) {
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
