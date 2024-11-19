package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
)

func Hello(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join("templates", "hello.html")

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Не удалось загрузить шаблон", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Ошибка при обработке шаблона", http.StatusInternalServerError)
	}
}
