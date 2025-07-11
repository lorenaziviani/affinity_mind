# Makefile para AffinityMind

.PHONY: all build test docker-backend docker-embedding docker-vector-db run-backend run-embedding run-vector-db

all: build

build:
	cd cmd/backend && go build -o backend main.go

# Tests Go
backend-test: # All projects must be running
	docker-compose down
	cd cmd/backend && go test -v


# Tests Python
embedding-test:
	docker-compose up -d embedding-server
	docker exec affinitymind-embedding pytest test_embedding.py
vector-db-test:
	docker-compose up -d vector-db
	docker exec affinitymind-vector-db pytest

test: backend-test embedding-test vector-db-test

docker-backend:
	cd cmd/backend && docker build -t affinitymind-backend .

docker-embedding:
	cd ml/embedding-server && docker build -t affinitymind-embedding .

docker-vector-db:
	cd infra/vector-db && docker build -t affinitymind-vector-db .

run-backend:
	cd cmd/backend && PORT=8080 go run main.go

run-embedding:
	cd ml/embedding-server && venv/bin/uvicorn main:app --reload --host 0.0.0.0 --port 5001

run-vector-db:
	cd infra/vector-db && venv/bin/uvicorn main:app --reload --host 0.0.0.0 --port 8001

