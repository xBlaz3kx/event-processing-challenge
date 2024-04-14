.PHONY: all up migrate generate

all: up migrate

# Run the solution with observability
up:
	docker compose -f ./deployments/docker/docker-compose.yml -f ./deployments/docker/docker-compose.observability.yaml up -d

migrate:
	docker compose exec database sh -c 'psql -U casino < /db/migrations/00001.create_base.sql'

generator:
	docker compose -f ./deployments/docker/docker-compose.yml run --rm generator

