# 📦 zip-downloader

[![Tests](https://github.com/DeneesK/zip-downloader-11-07-2025/actions/workflows/zip-downloader-tests.yml/badge.svg)](https://github.com/DeneesK/zip-downloader-11-07-2025/actions/workflows/zip-downloader-tests.yml)

![Go version](https://img.shields.io/badge/go-1.22-blue)

Сервис для скачивания файлов по ссылкам и архивации их в ZIP.

Принимает ссылки на `.pdf`, `.jpeg`, `.jpg` файлы, скачивает их и архивирует. Поддерживает ограничение по количеству одновременных задач и файлов в задаче. Предназначен для демонстрации слоистой архитектуры, очередей задач и in-memory-хранилища.

---

## ⚙️ Конфигурация и запуск

Сервис можно конфигурировать как через переменные окружения, так и через флаги командной строки.

### ✅ Переменные окружения

| Переменная         | Описание                              | Значение по умолчанию      |
|--------------------|----------------------------------------|-----------------------------|
| `SERVER_ADDRESS`   | Адрес и порт сервера                   | `localhost:8080`            |
| `ENV`              | Среда выполнения (`dev` или `prod`)    | `dev`                       |
| `ARCHIVE_DIR`      | Путь для хранения архивов              | `static/archives`           |
| `MAX_ACTIVE_TASKS` | Максимум активных задач                | `3`                         |
| `MAX_LINKS_PER_TASK` | Максимум ссылок в одной задаче       | `3`                         |

> Переменные окружения имеют приоритет над флагами.

---

### 🚀 Пример запуска

```bash
# Установить переменные окружения
export SERVER_ADDRESS=":8080"
export ENV="prod"
export ARCHIVE_DIR="./static/archives"
export MAX_ACTIVE_TASKS=3
export MAX_LINKS_PER_TASK=3

go run ./cmd/main.go
```

### или запуск с флагами

```
go run ./cmd/main.go \
  -a ":8080" \
  -env "dev" \
  -dir "./static/archives" \
  -tasks 3 \
  -links 3
```

## 🚀 Возможности

- Создание задачи на скачивание файлов
- Добавление до **3 ссылок** на `.pdf`, `.jpeg`, `.jpg` файлы
- Фоновая обработка задач с очередью
- Скачивание доступных файлов, упаковка в `.zip`
- Поддержка **до 3 активных задач одновременно**
- Информативный статус задачи: `created`, `running`, `done`, `failed`
- In-memory хранилище (без БД или Docker)

---

Паттерны:
- **Dependency Injection** через конструкторы
- **Интерфейсы для хранения**, логирования и архивирования
- **Очередь задач** (`chan string`)
- **Worker loop** для `processTask`


## 📡 API

Сервер работает на порту `:8080`. Все запросы и ответы — в формате `application/json`.

---

### 1. `POST /task` — создать задачу

Создает новую задачу на архивирование.

**Пример запроса:**

```bash
curl -X POST http://localhost:8080/task
```

### Ответ:

```
{
  "task_id": "b2f3f3f8-9234-4f5b-9f2c-1a6e4b3a1c12",
  "status": "created"
}
```

### Коды ответа:

- 201	Задача создана
- 429	Превышен лимит активных задач

### 2. PATCH /task/{id} — добавить ссылки
Добавляет ссылки (до 3) к задаче. Поддерживаются только .pdf, .jpeg, .jpg.

**Пример запроса:**

```bash
curl -X PATCH http://localhost:8080/task/{task_id} \
  -H "Content-Type: application/json" \
  -d '{
    "links": [
      "https://example.com/file1.pdf",
      "https://example.com/image.jpeg"
    ]
  }'
```
### Коды ответа:

- 200	Ссылки добавлены
- 400	Недопустимые типы файлов или превышен лимит
- 404	Задача не найдена

### 2. GET /task/{id} — получить статус задачи
Возвращает текущий статус задачи. Если задача завершена — возвращает ссылку на архив.

**Пример запроса:**

```bash
curl http://localhost:8080/task/{task_id}
```

### Ответ:

```
{
  "task_id": "b2f3f3f8-9234-4f5b-9f2c-1a6e4b3a1c12",
  "status": "done",
  "archive": "http://localhost:8080/static/1a6e4b3a1c12.zip"
}
```

### Пример ошибки

```
{
  "task_id": "b2f3f3f8-9234-4f5b-9f2c-1a6e4b3a1c12",
  "status": "failed",
  "failed_files": {
    "https://example.com/broken.pdf": "failed to download: 404 Not Found"
  }
}
```

### Коды ответа:

- 200	Задача найдена
- 404	Задача не найдена