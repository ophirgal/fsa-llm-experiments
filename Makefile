.PHONY: up down logs psql clean build-prod

up:
	docker compose up --build -d

down:
	docker compose down -v

logs:
	docker compose logs -f

psql:
	docker compose exec postgres psql -U $${POSTGRES_USER:-postgres} -d $${POSTGRES_DB:-appdb}

clean:
	docker compose down -v

build-prod:
	docker compose -f docker-compose.prod.yml up --build -d
