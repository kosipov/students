# Деплой в Dokploy

Dokploy сам забирает код из GitHub (ветка `main`), собирает образ по `Dockerfile` и перезапускает приложение при каждом пуше.

## Что нужно приложению

| Переменная | Обязательна | Значение |
|---|---|---|
| `MYSQL_HOST` | да | `host:port` базы, для базы в Dokploy — Internal Host и порт `3306` |
| `MYSQL_USER` | да | пользователь базы |
| `MYSQL_PASSWORD` | да | пароль пользователя |
| `MYSQL_DBNAME` | да | имя базы |
| `GIN_MODE` | нет | в образе по умолчанию `release` |
| `PORT` | нет | порт HTTP-сервера, по умолчанию `8080` |

Если обязательная переменная не задана, приложение падает при старте с `Invalid type assertion with key <NAME>`.

`config/config.yml` лежит в образе. `auth.hash_salt` оттуда нельзя менять после создания пользователей: пароли хэшируются с этой солью, и после смены никто не сможет войти.

Таблицы создаются автоматически при старте (`AutoMigrate`).

## 1. Подключить GitHub

1. Dokploy → **Settings → Git** → GitHub → **Create Github App**.
2. После создания нажать **Install**, выбрать репозиторий `kosipov/students` → **Install & Authorize**.

## 2. Создать проект и базу

1. **Create Project**, например `students`.
2. В проекте **Create Service → Database → MySQL**. Образ `mysql:8.0`: с ним приложение проверено.
3. Задать Database Name, User, Password → **Deploy**.
4. Со страницы базы понадобится **Internal Host** (для `MYSQL_HOST`). **External Port** не открывать: приложению он не нужен.
5. Настроить бэкапы: **Settings → Destinations** (S3-бакет), затем на странице базы **Backups**: destination, cron, например `0 3 * * *`. Нажать **Test** и убедиться, что файл появился в бакете.

## 3. Создать приложение

В том же проекте **Create Service → Application**.

**General → Provider: GitHub**
- Repository: `kosipov/students`
- Branch: `main`
- Build Path: `/`

**Build Type: Dockerfile**
- Docker File: `Dockerfile`
- Docker Context Path: `.`
- Docker Build Stage: пусто (нужен последний этап)

**Environment** — переменные из таблицы выше, например:

```
MYSQL_HOST=<Internal Host базы>:3306
MYSQL_USER=<user>
MYSQL_PASSWORD=<password>
MYSQL_DBNAME=<database>
```

**Advanced → Cluster Settings → Swarm Settings**

- **Health Check** — оставить пустым. В образе уже есть `HEALTHCHECK` на `/healthz` через `wget`, Swarm использует его. Пример из документации Dokploy с `curl` здесь не подойдёт: `curl` в образе нет. Если нужно задать явно:

  ```json
  {
    "Test": ["CMD-SHELL", "wget -q -O /dev/null http://127.0.0.1:8080/healthz || exit 1"],
    "Interval": 15000000000,
    "Timeout": 3000000000,
    "StartPeriod": 10000000000,
    "Retries": 3
  }
  ```

- **Update Config** — сначала поднимать новый контейнер, и только когда он станет healthy, гасить старый. Если новый не поднялся, откатываться:

  ```json
  {
    "Parallelism": 1,
    "Delay": 5000000000,
    "FailureAction": "rollback",
    "Monitor": 20000000000,
    "Order": "start-first"
  }
  ```

Replicas: `1`. Приложение хранит сессии в cookie, так что несколько реплик тоже будут работать, но пока в этом нет необходимости.

Нажать **Deploy** и проверить в **Deployments**, что сборка прошла, а в **Logs** — что приложение стартовало без ошибок.

## 4. Домен

1. Сначала проверить на временном домене: **Domains → traefik.me** (только HTTP), Container Port `8080`.
2. Боевой домен: A-запись `student.kosipov.ru` → IP сервера Dokploy (и `www.student.kosipov.ru`, если нужен).
3. **Domains → Add Domain**: Host `student.kosipov.ru`, Path `/`, Container Port `8080`, HTTPS включён, Certificate `letsencrypt`.

Сертификат выпускается, когда DNS уже указывает на сервер Dokploy. Порты 80 и 443 на сервере должны быть открыты.

## 5. Первый администратор

Регистрации в приложении нет, пользователь создаётся командой в контейнере приложения.

Через терминал контейнера в Dokploy:

```sh
./create-admin -username admin
```

Или по SSH на сервере Dokploy:

```sh
docker exec -it $(docker ps -q -f name=<appName> | head -n1) ./create-admin -username admin
```

`<appName>` — App Name приложения в Dokploy.

Пароль спрашивается без отображения ввода (не короче 8 символов). Команда откажет, если пользователь с таким логином уже есть.

## 6. Проверка

- `https://student.kosipov.ru/healthz` → `200`
- главная открывается, стили подгружены
- вход в `/auth/sign-in`, создание группы, дисциплины и задания, редактирование и удаление задания
- пуш в `main` запускает новый деплой (**Deployments**)

## Задания из OneDrive

Ссылки на markdown-файлы в OneDrive (`1drv.ms`, `onedrive.live.com`) открываются на сайте как страница `/tasks/:id`, остальные ссылки — как есть. Настраивать ничего не нужно, приложению нужен только исходящий HTTPS к `api-badgerp.svc.ms` и `api.onedrive.com`.

- Файл скачивается в фоне при создании задания или смене ссылки. Копия хранится в базе и проверяется на изменения не чаще раза в 5 минут.
- Загрузка идёт через недокументированный API, которым пользуется веб-версия OneDrive, поэтому Microsoft может его сломать. Тогда студенты продолжат видеть последние скачанные версии.
- Статус загрузки и ошибка видны в админке у каждого задания, кнопка «Обновить» скачивает файл заново. Ошибки также пишутся в лог: `Failed to fetch content of subject object`.

## Откат

- Сборка упала — старый контейнер продолжает работать, ничего делать не нужно.
- Новый контейнер не стал healthy — Swarm откатится сам (`FailureAction: rollback`).
- Ошибка в коде, которая проявилась позже, — `git revert` проблемного коммита и пуш в `main`.
