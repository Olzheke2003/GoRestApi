package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
)

// Проверка MIME-типа файла
func isValidFileType(mimeType string) bool {
	validMimeTypes := []string{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document", // .docx
		"application/pdf", // .pdf
	}
	for _, validType := range validMimeTypes {
		if mimeType == validType {
			return true
		}
	}
	return false
}

// Обработка запроса для отправки файла
func HandleFileAndEmails(w http.ResponseWriter, r *http.Request) {
	// Ограничиваем размер файла (например, 10 MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Printf("Ошибка при парсинге формы: %v", err)
		http.Error(w, "Ошибка при парсинге формы", http.StatusBadRequest)
		return
	}

	// Извлекаем файл из формы
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Ошибка при получении файла: %v", err)
		http.Error(w, "Ошибка при получении файла", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Проверяем MIME-тип файла (читаем первые 512 байт)
	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil && err != io.EOF {
		log.Printf("Ошибка при чтении файла: %v", err)
		http.Error(w, "Ошибка при чтении файла", http.StatusInternalServerError)
		return
	}
	mimeType := http.DetectContentType(buf)
	log.Printf("Определенный MIME-тип: %s", mimeType)

	if !isValidFileType(mimeType) {
		log.Printf("Неверный MIME-тип файла: %s", mimeType)
		http.Error(w, "Неверный MIME-тип файла", http.StatusBadRequest)
		return
	}

	// Извлекаем список email-адресов
	emails := r.FormValue("emails")
	emailList := strings.Split(emails, ",")
	if len(emailList) == 0 || emailList[0] == "" {
		log.Println("Список email-адресов пуст")
		http.Error(w, "Список email-адресов пуст", http.StatusBadRequest)
		return
	}

	// Считываем весь файл в память
	file.Seek(0, io.SeekStart) // Возвращаемся в начало файла
	fileContent, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Ошибка при чтении содержимого файла: %v", err)
		http.Error(w, "Ошибка при чтении содержимого файла", http.StatusInternalServerError)
		return
	}

	// Отправляем файл по почте
	err = sendEmailWithAttachment(emailList, fileContent, header.Filename, mimeType)
	if err != nil {
		log.Printf("Ошибка при отправке письма: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка при отправке письма: %v", err), http.StatusInternalServerError)
		return
	}

	log.Println("Файл успешно отправлен")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Файл успешно отправлен"))
}

func sendEmailWithAttachment(to []string, fileContent []byte, filename, mimeType string) error {
	from := os.Getenv("EMAIL_USER")
	password := os.Getenv("EMAIL_PASSWORD")

	// SMTP сервер и порт
	smtpHost := os.Getenv("SMTP_PORT")
	smtpPort := os.Getenv("SMTP_SERVER")

	// Кодируем файл в Base64
	encodedFile := base64.StdEncoding.EncodeToString(fileContent)

	// Формируем сообщение
	subject := "Subject: Отправка файла\n"
	mime := "MIME-Version: 1.0\n"
	contentType := "Content-Type: multipart/mixed; boundary=boundary\n\n"
	body := "--boundary\n" +
		"Content-Type: text/plain; charset=UTF-8\n\n" +
		"Здравствуйте!\nПрикрепленный файл.\n\n" +
		"--boundary\n" +
		fmt.Sprintf("Content-Type: %s; name=\"%s\"\n", mimeType, filename) +
		fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\n", filename) +
		"Content-Transfer-Encoding: base64\n\n" +
		encodedFile + "\n" +
		"--boundary--"

	message := []byte(subject + mime + contentType + body)

	// Аутентификация и отправка письма
	auth := smtp.PlainAuth("", from, password, smtpHost)
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
}
