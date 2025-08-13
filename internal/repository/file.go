package repository

import (
	"context"
	"encoding/json"
	"github.com/bezjen/shortener/internal/config"
	"github.com/bezjen/shortener/internal/model"
	"github.com/google/uuid"
	"io"
	"os"
	"sync"
)

type FileRepository struct {
	fileStorage   os.File
	encoder       json.Encoder
	decoder       json.Decoder
	memoryStorage map[string]model.ShortURLFileDto
	mu            sync.RWMutex
}

func NewFileRepository(cfg config.Config) (*FileRepository, error) {
	fileStorage, err := os.OpenFile(cfg.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	decoder := *json.NewDecoder(fileStorage)
	memoryStorage, err := loadFileData(decoder)
	if err != nil {
		return nil, err
	}
	return &FileRepository{
		fileStorage:   *fileStorage,
		memoryStorage: memoryStorage,
		encoder:       *json.NewEncoder(fileStorage),
		decoder:       decoder,
	}, nil
}

func (f *FileRepository) Save(_ context.Context, url model.URL) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.memoryStorage[url.ShortURL]; exists {
		return ErrConflict
	}
	shortURLDto, err := f.saveShortURLDtoToStorage(url)
	if err != nil {
		return err
	}
	f.memoryStorage[url.ShortURL] = *shortURLDto
	return nil
}

func (f *FileRepository) GetByShortURL(_ context.Context, shortURL string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	storedShortURLDto, exists := f.memoryStorage[shortURL]
	if !exists {
		return "", ErrNotFound
	}
	return storedShortURLDto.OriginalURL, nil
}

func (f *FileRepository) Ping(_ context.Context) error {
	return nil
}

func (f *FileRepository) Close() error {
	return f.fileStorage.Close()
}

func (f *FileRepository) saveShortURLDtoToStorage(url model.URL) (*model.ShortURLFileDto, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	shortURLDto := model.ShortURLFileDto{
		ID:          id,
		ShortURL:    url.ShortURL,
		OriginalURL: url.OriginalURL,
	}
	err = f.encoder.Encode(&shortURLDto)
	if err != nil {
		return nil, err
	}
	return &shortURLDto, nil
}

func loadFileData(decoder json.Decoder) (map[string]model.ShortURLFileDto, error) {
	memoryStorage := make(map[string]model.ShortURLFileDto)
	for {
		var dto model.ShortURLFileDto
		err := decoder.Decode(&dto)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		memoryStorage[dto.ShortURL] = dto
	}
	return memoryStorage, nil
}
