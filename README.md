### Прокси сервер

Команда для запуска:

`make start`

- генерирует сертификат и ключи
- собирает образ с приложением
- запускает контейнеры:
  - Postgres на 5432 порту
  - прокси-сервер на 8080 порту и веб-апи на 8000 порту


После этого нужно добавить в свою систему сертификат certs/ca.crt

Команды для Ubuntu:

```
sudo apt-get install -y ca-certificates
sudo cp certs/ca.crt /usr/local/share/ca-certificates
sudo update-ca-certificates
```
