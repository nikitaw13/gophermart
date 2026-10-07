# Локальный запуск тестов

## 1. statictest

Аналог `go vet` с дополнительными проверками Яндекса.

```bash
go install github.com/Yandex-Practicum/go-autotests/cmd/statictest@latest
go vet -vettool=$(which statictest) ./...
```

## 2. gophermarttest

Интеграционный автотест Яндекса (то же, что запускает `.github/workflows/gophermart.yml`).

### Подготовка

1. Postgres:

   ```bash
   docker run -d --name praktikum -p 5432:5432 \
     -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=praktikum postgres
   ```

2. Бинарарь gophermart:

   ```bash
   (cd cmd/gophermart && go build -o gophermart .)
   ```

3. Скачать бинарники из [go-autotests releases](https://github.com/Yandex-Practicum/go-autotests/releases)
   и положить в `.tools/` (в `cmd/accrual/` уже лежит свой `accrual_linux_amd64`):

   - `gophermarttest`
   - `random`

   ```bash
   TAG=$(curl -s https://api.github.com/repos/Yandex-Practicum/go-autotests/releases/latest \
     | grep -oP '"tag_name":\s*"\K[^"]+')

   cd .tools
   for f in gophermarttest random; do
     curl -LO "https://github.com/Yandex-Practicum/go-autotests/releases/download/$TAG/$f"
   done
   chmod +x *
   cd ..
   ```

### Запуск

```bash
.tools/gophermarttest \
  -test.v -test.run=^TestGophermart$ \
  -gophermart-binary-path=cmd/gophermart/gophermart \
  -gophermart-host=localhost \
  -gophermart-port=8080 \
  -gophermart-database-uri="postgresql://postgres:postgres@localhost:5432/praktikum?sslmode=disable" \
  -accrual-binary-path=cmd/accrual/accrual_linux_amd64 \
  -accrual-host=localhost \
  -accrual-port=$(.tools/random unused-port) \
  -accrual-database-uri="postgresql://postgres:postgres@localhost:5432/praktikum?sslmode=disable"
```

### Заметки

- Порт 8080 должен быть свободен.
- Перед каждым повторным прогоном пересоздавай базу:
  автотест не переносит остатки данных.

  ```bash
  docker rm -f praktikum
  docker run -d --name praktikum -p 5432:5432 \
    -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=praktikum postgres
  ```
