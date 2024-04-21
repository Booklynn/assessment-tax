dockerup:
	docker-compose up --build

dockerdown:
	docker-compose down

test:
	go test -v -cover ./...

.PHONY: dockerup dockerdown test
