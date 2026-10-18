# Ansible-пример деплоя

Это **справочная копия** роли `student` из боевого репозитория инфраструктуры
(`kosipov_ru_server`, каталог `ansible/`). Источник истины — тот репозиторий;
здесь файлы лежат как пример того, как сайт выкатывается на сервер, и могут
отставать от боевых версий.

Что делает роль (см. `roles/student/tasks/main.yml`):

1. Клонирует `github.com/kosipov/students` (ветка `main`) в `/opt/src/students`.
2. Раскладывает `/opt/students/.env` из шаблона `env.j2` (секреты подставляются
   из ansible-vault) и `docker-compose.yml` из шаблона.
3. Собирает образ по `Dockerfile` из корня репозитория и поднимает контейнер
   через `docker compose` (`build: always` — каждый прогон пересобирает).

Переменные, которые роль ждёт от инвентаря/group_vars:

| Переменная | Откуда | Назначение |
|---|---|---|
| `mysql_container` | vars | хост общего MySQL-контейнера |
| `vault_students_db_user` / `vault_students_db_password` | vault | доступ к базе |
| `campus_relay_url` | vars | публичный URL релея кампуса (Cloud.ru) |
| `vault_campus_relay_secret` | vault | секрет релея |
| `proxy_network` | vars | общая docker-сеть с Traefik |
| `base_domain` | vars | домен (`student.{{ base_domain }}`) |
| `github_repo_prefix` / `student_repo_path` | vars | адрес репозитория (с токеном для приватного) |

Запуск на боевом сервере (из каталога `ansible/` серверного репозитория):

```sh
ansible-playbook playbooks/site.yml --tags student
```

HTTPS, сертификат и маршрутизацию делает Traefik по лейблам из
`docker-compose.yml.j2`; само приложение слушает только HTTP :8080 во
внутренней сети.
