# SSO
Система единого входа в проекты ITS TECH
docker run --name=sso-db -e POSTGRES_PASSWORD='qwerty' -p 5436:5432 -d  postgres
migrate -path ./migrations -database 'postgres://postgres:qwerty@localhost:5436/postgres?sslmode=disable' up