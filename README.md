# Go gRPC PubSub Service

## Описание

Этот сервис реализует модель публикации-подписки (PubSub) на Go с использованием gRPC.  
Клиенты могут подписываться на события по ключу и получать уведомления, когда другие публикуют события с тем же ключом.

Реализация основывается на:

- Чистой архитектуре
- Каналах и горутинах
- gRPC API со статус-кодами (`codes.InvalidArgument`, `codes.Unavailable`, и др.)
- Graceful shutdown
- Подключаемой конфигурации

## Конфигурация

Параметры конфигурации устанавливаются через переменные окружения или берутся по умолчанию:

| Переменная         | Назначение                      | Значение по умолчанию |
|--------------------|----------------------------------|------------------------|
| `GRPC_PORT`        | Порт для основного сервиса       | `50051`                |
| `TEST_GRPC_ADDR1`  | Адрес для первого теста gRPC     | `localhost:50052`      |
| `TEST_GRPC_ADDR2`  | Адрес для второго теста gRPC     | `localhost:50053`      |

## Сборка и запуск

### Зависимости

Для работы и сборки проекта необходимы:

- **Go** 1.20 или выше
- **Protocol Buffers Compiler** (`protoc`)
- Плагины для генерации Go-кода из `.proto`:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Установка зависимостей

Выполните в корне проекта:

```bash
go mod tidy
```
### Сборка и запуск gRPC-сервера
```bash
go run main.go
```

По умолчанию сервер запускается на :50051. Порт можно переопределить переменной окружения:
```bash
GRPC_PORT=50505 go run main.go
```

## Тесты

Включают:

- Проверку корректной работы `Publish/Subscribe` по gRPC
- Проверку graceful shutdown
- Юнит-тесты pubsub логики

Запуск всех тестов:
```bash
go test -v ./...
```

## Использованные паттерны

- **Dependency Injection**: передача `subpub` в конструктор gRPC сервера
- **Graceful Shutdown**: `os.Signal`, `GracefulStop`
- **Config management**: чтение из `os.Getenv` с дефолтами
- **Structured Logging**: уровни логов, структура сообщений

## Как это работает

1. Клиенты отправляют запрос `Subscribe`, указывая ключ — подписываются на события по нему.
2. Любой клиент может вызвать `Publish`, указав ключ и сообщение.
3. Сервер доставляет сообщение всем подписанным клиентам (асинхронно, без блокировок).
4. Каждый подписчик обрабатывается в отдельной горутине, не мешая остальным.
5. Сервер использует `GracefulStop()` для корректного завершения и закрытия подписчиков через `Close(ctx)`.
6. Внутренние ошибки возвращаются с использованием gRPC `codes`, а логгирование помогает отладке.

## Структура проекта

```
Go-Practice-VK/
├── main.go               # Точка входа
├── config/               # Конфигуратор
├── subpub/               # Логика подписки
├── server/               # gRPC-реализация
├── gen/                  # Сгенерированные protobuf-файлы
├── proto/                # Исходный `.proto` файл
├── test/                 # Интеграционные и unit-тесты
```

## Протокол

Файл `proto/pubsub.proto` определяет интерфейс gRPC сервиса. Генерация:
```bash
protoc -I proto --go_out=gen --go-grpc_out=gen proto/pubsub.proto
```

// Author: Yvetlin [Gloria] \
// Date: 2025
