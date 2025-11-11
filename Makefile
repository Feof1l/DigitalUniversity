ifeq ($(OS),Windows_NT)
    DOCKER_COMPOSE = docker-compose
else
    ifeq ($(shell command -v docker-compose 2> /dev/null),)
        DOCKER_COMPOSE = docker compose
    else
        DOCKER_COMPOSE = docker-compose
    endif
endif

.PHONY: format down up reup help

format:
	@echo "Formatting code with gofumpt..."
	gofumpt -w .
	@echo "Done!"

down:
	@echo "Stopping containers..."
	$(DOCKER_COMPOSE) down
	@echo "Done!"

up:
	@echo "Starting containers with build..."
	$(DOCKER_COMPOSE) up --build

reup:
	@echo "Removing containers and volumes..."
	$(DOCKER_COMPOSE) down -v
	@echo "Starting containers with build..."
	$(DOCKER_COMPOSE) up --build

clean:
	@echo "Stopping all containers..."
	-docker stop $$(docker ps -a -q) 2>/dev/null || true
	@echo "Removing all containers..."
	-docker rm $$(docker ps -a -q) 2>/dev/null || true
	@echo "Removing all images..."
	-docker rmi $$(docker images -q) 2>/dev/null || true
	@echo "Pruning Docker system..."
	docker system prune -a --volumes -f
	@echo "Docker cleanup complete!"

help:
	@echo "Available targets:"
	@echo "  format  - Format Go code with gofumpt"
	@echo "  down    - Stop and remove containers"
	@echo "  up      - Start containers with build"
	@echo "  reup    - Full restart (remove volumes + rebuild)"
	@echo "  clean   - Remove ALL Docker containers, images, and volumes"
	@echo "  help    - Show this help message"
