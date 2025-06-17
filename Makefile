# Default service name(s) (can be left empty to act on all)
srv =

# Start containers
start:
	docker compose up -d $(srv)

# Stop containers
stop:
	docker compose down $(srv)

# Restart containers
restart: stop start

# Follow container logs
logs:
	docker compose logs -f $(srv)
