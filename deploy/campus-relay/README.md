# Посредник для кампуса (campus-relay)

Ходит на campus.syktsu.ru вместо сервера сайта. Нужен, если кампус не отвечает на запросы с адреса сервера. Две формы развёртывания:

- **`main.go` + `Dockerfile`** — статический бинарник (~10 МБ) для любого контейнерного хостинга. Рекомендуемый: Cloud.ru Evolution Container Apps (free tier, физлицам доступен). Пошагово — в [DEPLOY.md](../../DEPLOY.md), раздел «Если кампус не отвечает серверу».
- **`index.py`** — функция для Yandex Cloud Functions (если/когда сервис доступен).

Протокол у обоих одинаковый:

- `POST /` с телом формы (`application/x-www-form-urlencoded`);
- заголовок `X-Relay-Secret` — обязателен, совпадает с `CAMPUS_RELAY_SECRET` сайта;
- заголовок `X-Relay-Target`: `students` — список студентов группы (`get_student.php`), иначе — расписание преподавателя;
- адреса кампуса переопределяются переменными `CAMPUS_URL` и `CAMPUS_STUDENTS_URL`, порт — `PORT` (по умолчанию 8080).

Не является открытым прокси: POST только по этим двум адресам и только для тех, кто знает секрет.

Локальная проверка:

```sh
docker build -f deploy/campus-relay/Dockerfile -t campus-relay .   # из корня репозитория
docker run --rm -p 8080:8080 -e RELAY_SECRET=test campus-relay
curl -X POST -H 'X-Relay-Secret: test' -H 'X-Relay-Target: students' \
  --data-urlencode 'group=121б-пио' --data-urlencode 'reg=group' http://localhost:8080/
```
