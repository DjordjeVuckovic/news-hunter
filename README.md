# news-hunter
Full-text search engine for exploring multilingual news headlines and articles
## Migrations
- Install golang-migrate
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```
- Run migrations
```bash
migrate -path db/migrations -database "postgres://username:password@localhost:5432/news_db?sslmode=disable" up
```
- Local Postgres with Docker
```bash
migrate -path db/migrations -database "postgres://news_user:news_password@localhost:54320/news_db?sslmode=disable" up
```
## OpenAPI Schema gen
- Install swag CLI
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```
- Generate OpenAPI spec
```bash
swag init -g cmd/news_search/main.go -o ./api/openapi-spec
```
## Run tests
```bash
go test ./...
```
## Run linter
```bash
golangci-lint run
```
