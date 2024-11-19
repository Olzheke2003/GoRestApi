package handlers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	model "github.com/Olzheke2003/GoRestApi/internal/app/model"
)

// FileInfo содержит информацию о файле в архиве

// HandleArchiveInformation обрабатывает POST-запрос с архивом и возвращает информацию о нем
func HandleArchiveInformation(w http.ResponseWriter, r *http.Request) {
	// Логирование заголовков запроса
	log.Println("Received request with headers:")
	for name, values := range r.Header {
		for _, value := range values {
			log.Printf("%s: %s\n", name, value)
		}
	}

	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ограничиваем размер загружаемого файла (например, до 50MB)
	r.ParseMultipartForm(50 << 20) // 50MB

	// Читаем файл из multipart/form-data
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusBadRequest)
		log.Printf("Error reading file: %v\n", err)
		return
	}
	defer file.Close()

	// Логирование информации о файле
	log.Printf("Processing file: %s, Size: %d bytes\n", header.Filename, header.Size)

	// Проверяем, является ли файл ZIP-архивом
	if !isArchive(file) {
		http.Error(w, "File is not a valid archive", http.StatusBadRequest)
		log.Println("Uploaded file is not a valid archive")
		return
	}

	// Сбросить курсор файла для дальнейшего чтения
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, fmt.Sprintf("Failed to seek file: %v", err), http.StatusInternalServerError)
		log.Printf("Error seeking file: %v\n", err)
		return
	}

	// Сохраняем файл временно на диск
	tempFile, err := os.CreateTemp("", "uploaded-*.zip")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create temp file: %v", err), http.StatusInternalServerError)
		log.Printf("Error creating temp file: %v\n", err)
		return
	}
	defer os.Remove(tempFile.Name()) // Удаляем файл после обработки
	defer tempFile.Close()

	// Копируем содержимое загруженного файла во временный файл
	if _, err := io.Copy(tempFile, file); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		log.Printf("Error saving file: %v\n", err)
		return
	}

	// Проверяем, является ли файл ZIP-архивом
	zipReader, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		http.Error(w, "Uploaded file is not a valid ZIP archive", http.StatusBadRequest)
		log.Printf("Error opening zip file: %v\n", err)
		return
	}
	defer zipReader.Close()

	// Собираем информацию об архиве
	// Создаем переменную для информации об архиве
	var archiveInfo model.ArchiveInfo
	archiveInfo.FileName = header.Filename
	archiveInfo.ArchiveSize = float64(header.Size)
	archiveInfo.Files = make([]model.FileInfo, 0)

	var totalSize float64
	for _, f := range zipReader.File {
		fileInfo := model.FileInfo{
			FilePath: f.Name,
			Size:     float64(f.UncompressedSize64),
			MimeType: detectMimeType(f),
		}
		archiveInfo.Files = append(archiveInfo.Files, fileInfo)
		totalSize += fileInfo.Size
	}

	archiveInfo.TotalFiles = len(archiveInfo.Files)
	archiveInfo.TotalSize = totalSize

	log.Println("Form data:", r.MultipartForm)

	// Возвращаем информацию в формате JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(archiveInfo); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		log.Printf("Error encoding response: %v\n", err)
	}
}

// isArchive проверяет, начинается ли файл с магии ZIP
func isArchive(file io.Reader) bool {
	buf := make([]byte, 4)
	_, err := file.Read(buf)
	if err != nil {
		return false
	}

	// Проверка на начало ZIP файла (PK\x03\x04)
	return strings.HasPrefix(string(buf), "PK\x03\x04")
}

// detectMimeType определяет mime-тип файла в архиве
func detectMimeType(file *zip.File) string {
	reader, err := file.Open()
	if err != nil {
		return "unknown"
	}
	defer reader.Close()

	buffer := make([]byte, 512)
	_, err = reader.Read(buffer)
	if err != nil && err != io.EOF {
		return "unknown"
	}

	// Используем mime-тип на основе расширения файла
	return mime.TypeByExtension(filepath.Ext(file.Name))
}
