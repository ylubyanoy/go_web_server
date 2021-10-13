all:
	echo "Enter command"

# Docker compose
docker-dev-up:
	docker-compose -f docker-compose.dev.yml up --build -d

docker-dev-down:
	docker-compose -f docker-compose.dev.yml down

docker-prod-up:
	docker-compose up --build -d

docker-prod-down:
	docker-compose down

# Generate PKC#1 certs
gen-cert-priv-priv:
	openssl genrsa -f4 -out private 4096

gen-cert-priv-pub:
	openssl rsa -in private -outform PEM -pubout -out private.pub

gen-cert-refresh-priv:
	openssl genrsa -f4 -out refresh 4096

gen-cert-refresh-pub:
	openssl rsa -in refresh -outform PEM -pubout -out refresh.pub

# Migrationns
migrate-up:
	migrate -path migrations -database "postgres://postgres:12345@localhost:5432/postgres?sslmode=disable" -verbose up

migrate-down:
	migrate -path migrations -database "postgres://postgres:12345@localhost:5432/postgres?sslmode=disable" -verbose up
