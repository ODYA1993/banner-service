# Баннер-сервис

Этот проект представляет собой баннер-сервис, который позволяет управлять баннерами и пользователями.

## Установка

1. Клонируйте репозиторий:

```bash
git clone https://github.com/ODYA1993/banner-service.git
```

1. Перейдите в директорию проекта: cd banner-service

## Запуск:

1. - Создайте Docker-образ проекта используя make:
: make docker build


2. - Запустите контейнеры:
: make up

## Использование

После запуска контейнеров вы можете использовать баннер-сервис, отправляя HTTP-запросы к нему.
Рекомендую использовать приложение Postman

1. Для Регистрации пользователя отправьте POST-запрос ***localhost:8080/register*** с JSON-телом:
```bash
{
"name": "admin",
"email": "admin@admin.ru",
"password": "qwerty123",
"is_admin": true
}
```
ответ:

```bash
{
"id": 1,
"name": "admin",
"email": "admin@admin.ru",
"password": "$2a$10$7Uet/qrY5pnVQ8nTZNCopOoL6jBmddHCddHqzniz/JWKf2hs1wRZK",
"is_admin": true
}
```

2. Для Авторизации пользователя отправьте POST-запрос ***localhost:8080/login*** с JSON-телом:

```bash
{
"email": "admin@admin.ru",
"password": "qwerty123"
}
```

ответ:

```bash
{
"id": 1,
"name": "admin",
"email": "admin@admin.ru",
"password": "$2a$10$7Uet/qrY5pnVQ8nTZNCopOoL6jBmddHCddHqzniz/JWKf2hs1wRZK",
"is_admin": true
}
```
знаю что пароль возвращать нельзя:) это для себя)


3. Получение баннера для пользователя, отправьте GET-запрос ***localhost:8080/user_banner?tag_id=1&feature_id=1&use_last_revision=true***.

ответ:
```bash
{
"id": 1,
"title": "Banner 1",
"text": "This is the text of Banner 1",
"url": "https://example.com/banner1",
"is_active": true,
"feature_id": {
"id": 1,
"name": "Feature 1"
},
"tags": [
{
"id": 1,
"name": "Tag 1"
}
],
"created_at": "2024-04-14T14:13:07.360317Z",
"updated_at": "2024-04-14T14:13:07.360317Z"
}
```

4. Для создания нового баннера отправьте POST-запрос ***localhost:8080/banner*** с JSON-телом:

```bash
{
"title": "createTitle",
"text": "createText",
"url": "createURL",
"is_active": true,
"feature_id": {
"id": 3
},
"tags": [{"id": 3}]
}
```

ответ:

```bash
{
"id": 11,
"title": "createTitle",
"text": "createText",
"url": "createURL",
"is_active": true,
"feature_id": {
"id": 3,
"name": ""
},
"tags": [
{
"id": 3,
"name": ""
}
],
"created_at": "0001-01-01T00:00:00Z",
"updated_at": "0001-01-01T00:00:00Z"
}
```

5. Для Получения всех баннеров c фильтрацией по фиче и/или тегу отправьте GET-запрос ***localhost:8080/banner/3/3/3/0***

ответ:
```bash
[
{
"id": 11,
"title": "createTitle",
"text": "createText",
"url": "createURL",
"is_active": true,
"feature_id": {
"id": 3,
"name": "Feature 3"
},
"tags": [
{
"id": 3,
"name": "Tag 3"
}
],
"created_at": "2024-04-14T14:18:22.143111Z",
"updated_at": "2024-04-14T14:18:22.143111Z"
}
]
```

6. Для обновления информации о баннере отправьте PUT-запрос к ***localhost:8080/banner/1*** с JSON-телом.

```bash
{
"title": "title",
"text": "text",
"url": "url",
"is_active": true,
"feature_id": {
"id": 2
},
"tags": [{"id": 2}]
}
```

ответ:
```bash
{
"id": 1,
"title": "title",
"text": "text",
"url": "url",
"is_active": true,
"feature_id": {
"id": 2,
"name": ""
},
"tags": [
{
"id": 2,
"name": ""
}
],
"created_at": "0001-01-01T00:00:00Z",
"updated_at": "0001-01-01T00:00:00Z"
}
```

7. Удаление баннера по идентификатору ***localhost:8080/delete-banner/1***.

ответ:
```bash
{
"message": "banner with ID (id 1) deleted"
}
```

## Тестирование

В проекте предусмотрены интеграционные тесты, которые проверяют работу баннер-сервиса. Для запуска теста выполните команду:

- make test

Эта команда запустит тест, который находится в директории tests/integration.

ответ:

```bash
go test -v ./tests/integration/banner_test.go
=== RUN   TestGetUserBanner
time="2024-04-14T17:22:24+03:00" level=info msg="read application configuration" func="banner-service/internal/config.GetConfig.func1()" file="config.go, 41"
time="2024-04-14T17:22:24+03:00" level=trace msg="SQL Query: \n    SELECT b.id, b.title, b.text, b.url, b.is_active, b.created_at, b.updated_at, array_agg(t.id) as tag_ids, f.id as feature_id, f.name as feature_name\n    FROM banners b\n    JOIN features f ON b.feature_id = f.id\n    JOIN banner_tags bt ON b.id = bt.banner_id\n    JOIN tags t ON bt.tag_id = t.id\n    WHERE f.id = $1 AND t.id = $2\n    GROUP BY b.id, f.id\n    ORDER BY\nb.updated_at DESC LIMIT 1" func="banner-service/internal/models/banner/dbbanner.(*bannerRepository).GetBannerFromDB()" file="postgres.go, 45"
--- PASS: TestGetUserBanner (0.08s)
PASS
ok      command-line-arguments  0.158s

```

## Конфигурация
Баннер-сервис использует конфигурационный файл config.yaml для настройки подключения к базе данных. Вы можете изменить параметры подключения, отредактировав этот файл.

## Автор
Дмитрий Одинцов - ODYA1993