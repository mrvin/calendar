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

#### Вход пользователя
```bash
curl -i -X POST 'http://localhost:8080/api/auth/login' \
-H "Content-Type: application/json" \
-d '{
  "username": "Bob",
  "password": "qwerty"
}'
```

#### Получения информации о пользователе
```bash
curl -i -X GET 'http://localhost:8080/api/auth/me' \
-H "Authorization: Bearer <token>"
```

#### Удаление пользователя
```bash
curl -i -X DELETE 'http://localhost:8080/api/auth/me' \
-H "Authorization: Bearer <token>"
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

#### Получить событие
```bash
curl -i -X GET 'http://localhost:8080/api/events/{id}' \
-H "Authorization: Bearer <token>"
```

#### Удаление события
```bash
curl -i -X DELETE 'http://localhost:8080/api/events/{id}' \
-H "Authorization: Bearer <token>"
```

#### Список cобытий (на день/на неделю/нa месяц)
```bash
curl -i -X GET 'http://localhost:8080/api/events?date=2022-05-25' \
-H "Authorization: Bearer <token>"
```