package repository

import (
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
	memoryStorage map[string]model.ShortURLDto
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

func (f *FileRepository) Save(shortURL string, url string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.memoryStorage[shortURL]; exists {
		return ErrConflict
	}
	shortURLDto, err := f.saveShortURLDtoToStorage(shortURL, url)
	if err != nil {
		return err
	}
	f.memoryStorage[shortURL] = *shortURLDto
	return nil
}

func (f *FileRepository) GetByShortURL(shortURL string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	storedShortURLDto, exists := f.memoryStorage[shortURL]
	if !exists {
		return "", ErrNotFound
	}
	return storedShortURLDto.OriginalURL, nil
}

func (f *FileRepository) Close() error {
	return f.fileStorage.Close()
}

func (f *FileRepository) Ping() error {
	return nil
}

func (f *FileRepository) saveShortURLDtoToStorage(shortURL string, originalURL string) (*model.ShortURLDto, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	shortURLDto := model.ShortURLDto{
		ID:          id,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
	err = f.encoder.Encode(&shortURLDto)
	if err != nil {
		return nil, err
	}
	return &shortURLDto, nil
}

func loadFileData(decoder json.Decoder) (map[string]model.ShortURLDto, error) {
	memoryStorage := make(map[string]model.ShortURLDto)
	for {
		var dto model.ShortURLDto
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
