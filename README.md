# news-hunter
Full-text search engine for exploring multilingual news headlines and articles
## Migrations
- Install golang-migrate
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```
## Schema gen
- Api
```bash
swag init -g cmd/news_search/main.go -o ./api/openapi-spec
```
