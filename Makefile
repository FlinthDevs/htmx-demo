default: help

start: # Start the application
	@docker compose up -d
	@echo "Website accessible at http://htmx.docker.localhost"

stop: # Stop the app
	@docker compose stop

refresh: # Reload only go.
	@docker compose stop go
	@docker compose start go

restart: # Restart the app
	@make stop
	@make start

debug: # Starts container attached.
	@docker compose up

start-offline: # Starts the application offline.
	@OFFLINE=1 docker compose up -d
	@echo "Website accessible at http://htmx.docker.localhost"

logs: # Show docker logs
	@docker compose logs -f --since=1h --timestamps go

docker-remove: # Remove the app
	@docker compose down --remove-orphans -v

db: # Connects to db
	@docker compose exec -it db mysql -uhtmx -phtmx htmx

db-root: # Connects to db
	@docker compose exec -it db mysql -uroot -proot

help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: start stop restart refresh debug start-offline logs docker-remove db db-root help
