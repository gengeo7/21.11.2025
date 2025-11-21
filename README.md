# Workmate test

## Запуск 

```sh
git clone https://github.com/gengeo7/21.11.2025.git
cd 21.11.2025
go mod download
go run cmd/main.go
```

## Описание

При реализации использовал только одну библиотеку: gopdf

Api доступен по аддресу: localhost:3000

Маршрут проверки ссылок: 
```
POST localhost:3000/status
Content-Type: application/json

Body {
    "links": ["google.com", "malformedlink.gg"]
}
```

Маршрут для получения отчета: 
```
POST localhost:3000/pdf
Content-Type: application/json

Body {
    "links_list": [1, 2]
}
```

Для хранения испозовал простые json файлы. При запросе статус сохроняется в json документ в диреторию status_data.
При остановки сервера все запросы возвращают ошибку 503, сохранняя при этом данные запроса в директорию tasks_data.
При повторном запуске сервера все сохраненные данные считываются и обрабатываются. 
При запросе о статусе сервера, сервис запускает worker pool для проверки статусов.

