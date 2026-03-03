# Makefile
.PHONY: build run stop clean logs certs

build:
	docker-compose build

run:
	docker-compose up -d

stop:
	docker-compose down

logs:
	docker-compose logs -f

clean:
	docker-compose down -v
	docker system prune -f

# Генерация самоподписанных сертификатов для тестирования
certs:
	mkdir -p certs
	openssl req -x509 -newkey rsa:4096 -keyout certs/key.pem -out certs/cert.pem -days 365 -nodes -subj "/CN=localhost"
	@echo "✅ Самоподписанные сертификаты созданы в ./certs/"

# Для продакшена с Let's Encrypt
prod-certs:
	docker-compose run --rm certbot certonly --webroot --webroot-path=/var/www/html -d your-domain.com
