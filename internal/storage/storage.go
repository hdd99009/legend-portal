package storage

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type StoredFile struct {
	Key string
	URL string
}

type FileStorage interface {
	SaveImage(src io.Reader, ext string) (StoredFile, error)
	Delete(key string) error
	PublicURL(key string) string
	LocalPath(key string) (string, bool)
}

type LocalStorage struct {
	rootDir      string
	publicPrefix string
}

func NewLocalStorage(rootDir, publicPrefix string) *LocalStorage {
	if publicPrefix == "" {
		publicPrefix = "/uploads"
	}
	return &LocalStorage{
		rootDir:      rootDir,
		publicPrefix: strings.TrimRight(publicPrefix, "/"),
	}
}

func (s *LocalStorage) SaveImage(src io.Reader, ext string) (StoredFile, error) {
	datePath := time.Now().Format("2006/01")
	savedName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	key := filepath.ToSlash(filepath.Join(datePath, savedName))
	target := filepath.Join(s.rootDir, filepath.FromSlash(key))

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return StoredFile{}, err
	}

	dst, err := os.Create(target)
	if err != nil {
		return StoredFile{}, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return StoredFile{}, err
	}

	return StoredFile{
		Key: key,
		URL: s.PublicURL(key),
	}, nil
}

func (s *LocalStorage) Delete(key string) error {
	target := filepath.Join(s.rootDir, filepath.FromSlash(key))
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *LocalStorage) PublicURL(key string) string {
	escaped := (&url.URL{Path: filepath.ToSlash(key)}).String()
	return s.publicPrefix + "/" + strings.TrimPrefix(escaped, "/")
}

func (s *LocalStorage) LocalPath(key string) (string, bool) {
	return filepath.Join(s.rootDir, filepath.FromSlash(key)), true
}
