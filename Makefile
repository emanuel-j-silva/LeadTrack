up:
	docker compose up --build

down:
	docker compose down -v

test:
	cd backend && go test ./...
	cd frontend && npm test

lint:
	cd backend && go vet ./...
	# golangci-lint run ./... (if installed)
	cd frontend && npm run lint

fmt:
	cd backend && go fmt ./...
	cd frontend && npm run format
