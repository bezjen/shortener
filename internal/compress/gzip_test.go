package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNewGzipWriter проверяет создание GzipWriter и базовую запись
func TestNewGzipWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	gzWriter := NewGzipWriter(recorder)
	defer gzWriter.Close()

	// Проверяем что заголовки клонируются
	gzWriter.Header().Set("Test-Header", "test-value")

	if gzWriter.header.Get("Test-Header") != "test-value" {
		t.Error("Expected header to be set in cloned header")
	}

	// Проверяем запись данных
	testData := "Hello, GZIP!"
	n, err := gzWriter.Write([]byte(testData))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
	}

	// Закрываем writer чтобы завершить сжатие
	err = gzWriter.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Проверяем что данные сжаты
	if !bytes.Equal(recorder.Body.Bytes()[:2], []byte{0x1f, 0x8b}) {
		t.Error("Response body should be gzip compressed")
	}
}

// TestGzipWriter_WriteHeader проверяет установку заголовков и кодов состояния
func TestGzipWriter_WriteHeader(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantGzip   bool
	}{
		{"Status OK should gzip", http.StatusOK, true},
		{"Status No Content should not gzip", http.StatusNoContent, false},
		{"Status Conflict should gzip", http.StatusConflict, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			gzWriter := NewGzipWriter(recorder)
			defer gzWriter.Close()

			// Устанавливаем кастомный заголовок
			gzWriter.Header().Set("X-Test", "value")

			gzWriter.WriteHeader(tt.statusCode)

			// Проверяем что заголовки применены
			if recorder.Header().Get("X-Test") != "value" {
				t.Error("Custom headers not applied")
			}

			// Проверяем заголовок Content-Encoding
			hasGzip := recorder.Header().Get("Content-Encoding") == "gzip"
			if hasGzip != tt.wantGzip {
				t.Errorf("Content-Encoding: gzip = %v, want %v", hasGzip, tt.wantGzip)
			}

			// Проверяем код состояния
			if recorder.Code != tt.statusCode {
				t.Errorf("Status code = %d, want %d", recorder.Code, tt.statusCode)
			}
		})
	}
}

// TestGzipWriter_Close проверяет корректное закрытие и возврат в пул
func TestGzipWriter_Close(t *testing.T) {
	recorder := httptest.NewRecorder()
	gzWriter := NewGzipWriter(recorder)

	// Пишем некоторые данные
	_, err := gzWriter.Write([]byte("test data"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Закрываем writer
	err = gzWriter.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Пытаемся закрыть повторно (не должно паниковать)
	err = gzWriter.Close()
	if err != nil {
		t.Fatalf("Second close failed: %v", err)
	}

	// Проверяем что writer был возвращен в пул
	// путем создания нового writer'а и проверки что пул работает
	newRecorder := httptest.NewRecorder()
	newGzWriter := NewGzipWriter(newRecorder)

	// Если пул работает, мы должны получить валидный writer
	if newGzWriter.gw == nil {
		t.Error("New writer should have gzip writer instance")
	}
	newGzWriter.Close()
}

// TestGzipWriter_WriteAutoHeader проверяет автоматический вызов WriteHeader при Write
func TestGzipWriter_WriteAutoHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	gzWriter := NewGzipWriter(recorder)
	defer gzWriter.Close()

	// Write должен автоматически вызвать WriteHeader(http.StatusOK)
	_, err := gzWriter.Write([]byte("test"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if recorder.Code != http.StatusOK {
		t.Errorf("Auto status code = %d, want %d", recorder.Code, http.StatusOK)
	}
}

// TestNewGzipReader проверяет создание GzipReader и чтение данных
func TestNewGzipReader(t *testing.T) {
	// Создаем тестовые сжатые данные
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	originalText := "Hello, GZIP Reader!"
	_, err := gz.Write([]byte(originalText))
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	gz.Close()

	// Создаем GzipReader
	reader := io.NopCloser(&buf)
	gzReader, err := NewGzipReader(reader)
	if err != nil {
		t.Fatalf("NewGzipReader failed: %v", err)
	}
	defer gzReader.Close()

	// Читаем и проверяем данные
	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if string(decompressed) != originalText {
		t.Errorf("Decompressed data = %s, want %s", string(decompressed), originalText)
	}
}

// TestGzipReader_Close проверяет закрытие reader'а и возврат в пул
func TestGzipReader_Close(t *testing.T) {
	// Создаем минимальные валидные gzip данные
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Close()

	reader := io.NopCloser(&buf)
	gzReader, err := NewGzipReader(reader)
	if err != nil {
		t.Fatalf("NewGzipReader failed: %v", err)
	}

	// Закрываем reader
	err = gzReader.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Пытаемся закрыть повторно
	err = gzReader.Close()
	if err != nil {
		t.Fatalf("Second close failed: %v", err)
	}
}

// TestGzipReader_InvalidData проверяет обработку невалидных данных
func TestGzipReader_InvalidData(t *testing.T) {
	// Создаем невалидные gzip данные
	reader := io.NopCloser(strings.NewReader("invalid gzip data"))
	_, err := NewGzipReader(reader)

	if err == nil {
		t.Error("Expected error with invalid gzip data")
	}
}

// TestGzipWriter_Pool проверяет использование пула writer'ов
func TestGzipWriter_Pool(t *testing.T) {
	// Создаем несколько writer'ов чтобы проверить работу пула
	recorders := make([]*httptest.ResponseRecorder, 3)
	writers := make([]*GzipWriter, 3)

	for i := 0; i < 3; i++ {
		recorders[i] = httptest.NewRecorder()
		writers[i] = NewGzipWriter(recorders[i])

		// Проверяем что writer создан
		if writers[i].gw == nil {
			t.Errorf("Writer %d has no gzip writer", i)
		}
	}

	// Закрываем всех writer'ов (возвращаем в пул)
	for i := 0; i < 3; i++ {
		err := writers[i].Close()
		if err != nil {
			t.Errorf("Close writer %d failed: %v", i, err)
		}
	}
}

// TestGzipReader_Pool проверяет использование пула reader'ов
func TestGzipReader_Pool(t *testing.T) {
	// Создаем валидные gzip данные
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte("test data"))
	gz.Close()

	// Создаем несколько reader'ов
	readers := make([]*GzipReader, 3)
	for i := 0; i < 3; i++ {
		reader := io.NopCloser(bytes.NewReader(buf.Bytes()))
		gzReader, err := NewGzipReader(reader)
		if err != nil {
			t.Fatalf("NewGzipReader %d failed: %v", i, err)
		}
		readers[i] = gzReader
	}

	// Закрываем всех reader'ов (возвращаем в пул)
	for i := 0; i < 3; i++ {
		err := readers[i].Close()
		if err != nil {
			t.Errorf("Close reader %d failed: %v", i, err)
		}
	}
}

// TestGzipWriter_Concurrent проверяет конкурентное использование
func TestGzipWriter_Concurrent(t *testing.T) {
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			recorder := httptest.NewRecorder()
			gzWriter := NewGzipWriter(recorder)

			// Пишем данные
			testData := []byte("concurrent test data")
			_, err := gzWriter.Write(testData)
			if err != nil {
				t.Errorf("Concurrent write failed: %v", err)
			}

			// Закрываем writer
			err = gzWriter.Close()
			if err != nil {
				t.Errorf("Concurrent close failed: %v", err)
			}

			done <- true
		}(i)
	}

	// Ждем завершения всех горутин
	for i := 0; i < concurrency; i++ {
		<-done
	}
}
