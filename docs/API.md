#### Регистрация пользователя
```bash
curl -i -X POST 'http://localhost:8080/api/auth/register' \
-H "Content-Type: application/json" \
-d '{
  "username": "Bob",
  "password": "qwerty",
  "email": "bob@mail.com"
}'
```
```bash
grpcurl -plaintext \
-d '{
  "username": "Alice",
  "password": "qwerty",
  "email": "alice@example.com"
}' \
localhost:50051 calendar.CalendarService/Register
```

#### Вход пользователя
```bash
curl -i -X POST 'http://localhost:8080/api/auth/login' \
-H "Content-Type: application/json" \
-d '{
  "username": "Bob",
  "password": "qwerty"
}'
```
```bash
grpcurl -plaintext \
-d '{
  "username": "Alice",
  "password": "qwerty"
}' \
localhost:50051 calendar.CalendarService/Login
```

#### Получения информации о пользователе
```bash
curl -i -X GET 'http://localhost:8080/api/auth/me' \
-H "Authorization: Bearer <token>"
```
```bash
grpcurl -plaintext \
-H "Authorization: Bearer <token>" \
-d '{}' \
localhost:50051 calendar.CalendarService/GetUser
```

#### Удаление пользователя
```bash
curl -i -X DELETE 'http://localhost:8080/api/auth/me' \
-H "Authorization: Bearer <token>"
```
```bash
grpcurl -plaintext \
-H "Authorization: Bearer <token>" \
-d '{}' \
localhost:50051 calendar.CalendarService/DeleteUser
```

#### Добавление события
```bash
curl -i -X POST 'http://localhost:8080/api/events' \
-H "Content-Type: application/json" \
-H "Authorization: Bearer <token>" \
-d '{
	"title":"Aliss Birthday",
	"description":"Birthday April 12, 1996. House party",
	"start_time":"2022-05-25T10:41:31Z",
	"end_time":"2022-05-25T14:41:31Z",
	"notify_before":3600000000000
}'
```
```bash
grpcurl -plaintext \
-H "Authorization: Bearer <token>" \
-d '{
  "title":"Bobs Birthday",
  "description":"Birthday April 12, 1996. House party",
  "start_time":"2022-05-25T10:41:31Z",
  "end_time":"2022-05-25T14:41:31Z",
  "notify_before":"3600s"
}' \
localhost:50051 calendar.CalendarService/CreateEvent
```

#### Получить событие
```bash
curl -i -X GET 'http://localhost:8080/api/events/{id}' \
-H "Authorization: Bearer <token>"
```
```bash
grpcurl -plaintext \
-H "Authorization: Bearer <token>" \
-d '{
  "id":"<id>"
}' \
localhost:50051 calendar.CalendarService/GetEvent
```

#### Список cобытий (на день/на неделю/нa месяц)
```bash
curl -i -X GET 'http://localhost:8080/api/events?date=2022-05-25' \
-H "Authorization: Bearer <token>"
```
```bash
grpcurl -plaintext \
-H "Authorization: Bearer <token>" \
-d '{
  "start_time":"2022-05-25T00:00:00Z",
  "end_time":"2022-05-26T00:00:00Z"
}' \
localhost:50051 calendar.CalendarService/ListEvents
```

#### Обновление события
```bash
grpcurl -plaintext \
-H "Authorization: Bearer <token>" \
-d '{
  "id":"<id>",
  "title":"Bills Birthday",
  "description":"House party",
  "start_time":"2026-02-15T10:41:31Z",
  "end_time":"2026-02-16T14:41:31Z",
  "notify_before":"20s"
}' \
localhost:50051 calendar.CalendarService/UpdateEvent
```

#### Удаление события
```bash
curl -i -X DELETE 'http://localhost:8080/api/events/{id}' \
-H "Authorization: Bearer <token>"
```
```bash
grpcurl -plaintext \
-H "Authorization: Bearer <token>" \
-d '{
  "id":"<id>"
}' \
localhost:50051 calendar.CalendarService/DeleteEvent
```
