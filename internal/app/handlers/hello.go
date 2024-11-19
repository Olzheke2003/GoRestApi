package handler

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func Hello(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join("templates", "hello.html")

	// Парсим шаблон
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Не удалось загрузить шаблон", http.StatusInternalServerError)
		return
	}

	// Рендерим шаблон
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Ошибка при обработке шаблона", http.StatusInternalServerError)
	}
}
