package photogalery

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"image"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/jinzhu/gorm"

	"github.com/disintegration/imaging"
)

const (
	tempImageName     = "temp_image."
	tempImagesDir     = "./tmp/images/temp"
	originalImagesDir = "./tmp/images/original"
	thumbnailsDir     = "./tmp/images/thumbnails"
)

// Photo is model for a photo data
type Photo struct {
	gorm.Model
	ImageType string `gorm:"not null"`
	FileName  string `gorm:"not null"`
	URL       string `gorm:"unique;not null"`
	Format    string `gorm:"not null"`
	Size      int64
	Width     int
	Height    int
}

// Service is the interface that provides photogalery methods.
type Service interface {
	Upload(file *multipart.FileHeader) (*Photo, error)
	Photos() []Photo
	Delete(id int) (*Photo, error)
}

type service struct {
	db *gorm.DB
}

// NewService returns a new instance of a photogalery Service.
func NewService(db *gorm.DB) Service {
	db.AutoMigrate(&Photo{})
	return &service{db: db}
}

func (s *service) Upload(file *multipart.FileHeader) (*Photo, error) {

	mkDirs()

	var format string

	mimeType := file.Header.Get("Content-Type")
	switch mimeType {
	case "image/jpeg":
		format = "jpg"
	case "image/png":
		format = "png"
	case "image/gif":
		format = "gif"
	default:
		return nil, errors.New("the file format is not valid")
	}

	tempFileName := tempImageName + format
	tempImagePath := filepath.Join(tempImagesDir, tempFileName)
	if _, err := saveUploadedFile(file, tempImagePath); err != nil {
		return nil, err
	}

	fhash, err := fileHash(tempImagePath)
	if err != nil {
		return nil, err
	}

	originalFileName := fhash + "." + format
	originalImagePath := filepath.Join(originalImagesDir, originalFileName)
	err = os.Rename(tempImagePath, originalImagePath)
	if err != nil {
		return nil, err
	}

	go saveThumbnail(originalFileName, originalImagePath, format, 1000, 1000, s.db)

	reader, err := os.Open(originalImagePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	stat, err := reader.Stat()
	if err != nil {
		return nil, err
	}

	im, _, err := image.DecodeConfig(reader)
	if err != nil {
		return nil, err
	}

	photo := &Photo{
		ImageType: "original",
		FileName:  originalFileName,
		URL:       "/" + originalImagePath,
		Format:    format,
		Size:      stat.Size(),
		Width:     im.Width,
		Height:    im.Height,
	}

	s.db.Create(photo)

	return photo, nil
}

func (s *service) Photos() []Photo {
	photos := make([]Photo, 0)
	s.db.Find(&photos)
	return photos
}

func (s *service) Delete(id int) (*Photo, error) {
	var original Photo
	s.db.First(&original, id)

	if original.ImageType == "thumbnail" {
		return nil, errors.New("incorrect id")
	}

	err := os.Remove("." + original.URL)
	if err != nil {
		return nil, err
	}

	var thumbnail Photo
	s.db.Where(&Photo{FileName: original.FileName, ImageType: "thumbnail"}).First(&thumbnail)

	err = os.Remove("." + thumbnail.URL)
	if err != nil {
		return nil, err
	}

	s.db.Unscoped().Where([]uint{original.ID, thumbnail.ID}).Delete(&Photo{})
	return &original, nil
}

func mkDirs() {
	os.MkdirAll(tempImagesDir, os.ModePerm)
	os.MkdirAll(originalImagesDir, os.ModePerm)
	os.MkdirAll(thumbnailsDir, os.ModePerm)
}

func saveThumbnail(srcName, srcPath, format string, width, height int, db *gorm.DB) {
	src, err := imaging.Open(srcPath)
	if err != nil {
		log.Fatalf("thumbnail: %s\n", err)
	}

	dst := imaging.Thumbnail(src, width, height, imaging.Lanczos)

	thumbnailPath := filepath.Join(thumbnailsDir, srcName)
	err = imaging.Save(dst, thumbnailPath)
	if err != nil {
		log.Fatalf("thumbnail: %s\n", err)
	}

	reader, err := os.Open(thumbnailPath)
	if err != nil {
		log.Fatalf("thumbnail: %s\n", err)
	}
	defer reader.Close()

	stat, err := reader.Stat()
	if err != nil {
		log.Fatalf("thumbnail: %s\n", err)
	}

	im, _, err := image.DecodeConfig(reader)
	if err != nil {
		log.Fatalf("thumbnail: %s\n", err)
	}

	db.Create(&Photo{
		ImageType: "thumbnail",
		FileName:  srcName,
		URL:       "/" + thumbnailPath,
		Format:    format,
		Size:      stat.Size(),
		Width:     im.Width,
		Height:    im.Height,
	})
}

func fileHash(src string) (string, error) {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return "", err
	}

	h := md5.New()
	_, err = h.Write(data)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func saveUploadedFile(file *multipart.FileHeader, dst string) (int64, error) {
	src, err := file.Open()
	if err != nil {
		return 0, err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	return io.Copy(out, src)
}
