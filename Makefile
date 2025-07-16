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

gen-mocks:
	@src=$(src); \
	dir=$$(dirname $$src); \
	base=$$(basename $$src .go); \
	dest="$$dir/mock_$$base.go"; \
	pkg=$$(basename $$dir); \
	echo "Generating mock: src=$$src dest=$$dest pkg=$$pkg"; \
	mockgen -source=$$src -destination=$$dest -package=$$pkg