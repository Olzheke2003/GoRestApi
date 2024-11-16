package tasks

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

// FileInfo содержит информацию о файле в архиве
type FileInfo struct {
	FilePath string  `json:"file_path"`
	Size     float64 `json:"size"`
	MimeType string  `json:"mimetype"`
}

// ArchiveInfo содержит информацию об архиве
type ArchiveInfo struct {
	FileName    string     `json:"filename"`
	ArchiveSize float64    `json:"archive_size"`
	TotalSize   float64    `json:"total_size"`
	TotalFiles  int        `json:"total_files"`
	Files       []FileInfo `json:"files"`
}

func HandleArchiveInformation(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Читаем файл из multipart/form-data
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Сохраняем файл временно на диск
	tempFile, err := os.CreateTemp("", "uploaded-*.zip")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create temp file: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name()) // Удаляем файл после обработки
	defer tempFile.Close()

	// Копируем содержимое загруженного файла во временный файл
	if _, err := io.Copy(tempFile, file); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
		return
	}

	// Проверяем, является ли файл ZIP-архивом
	zipReader, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		http.Error(w, "Uploaded file is not a valid ZIP archive", http.StatusBadRequest)
		return
	}
	defer zipReader.Close()

	// Собираем информацию об архиве
	var archiveInfo ArchiveInfo
	archiveInfo.FileName = header.Filename
	archiveInfo.ArchiveSize = float64(header.Size)
	archiveInfo.Files = make([]FileInfo, 0)

	var totalSize float64
	for _, file := range zipReader.File {
		fileInfo := FileInfo{
			FilePath: file.Name,
			Size:     float64(file.UncompressedSize64),
			MimeType: detectMimeType(file),
		}
		archiveInfo.Files = append(archiveInfo.Files, fileInfo)
		totalSize += fileInfo.Size
	}

	archiveInfo.TotalFiles = len(archiveInfo.Files)
	archiveInfo.TotalSize = totalSize

	// Возвращаем информацию в формате JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(archiveInfo)
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

	return mime.TypeByExtension(filepath.Ext(file.Name))
}
