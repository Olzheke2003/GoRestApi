# Doodocs Backend Challenge

## Задание
Необходимо разработать **REST API** для работы с архивами.

### 1. Роут для отображения информации по архиву
- **Метод**: `POST /api/archive/information`
- **Запрос**: принимает файл через `multipart/form-data` в поле `file`.
- **Ответ**: JSON с информацией об архиве:
    - **filename**: имя архива
    - **archive_size**: размер архива
    - **total_size**: общий размер файлов в архиве
    - **total_files**: количество файлов
    - **files**: массив файлов с полями:
        - **file_path**: путь к файлу
        - **size**: размер файла
        - **mimetype**: MIME тип файла

### 2. Роут для создания архива
- **Метод**: `POST /api/archive/files`
- **Запрос**: принимает файлы через `multipart/form-data` в поле `files[]`. Разрешенные MIME типы:
    - `application/vnd.openxmlformats-officedocument.wordprocessingml.document`
    - `application/xml`
    - `image/jpeg`
    - `image/png`
- **Ответ**: архив в формате `.zip`.
