-include .env
export

run:
	docker compose --env-file .env up -d --build

migration-up:
	@migrate -database $(DB_LOCAL_URL) -path $(MIGRATIONS_DIR) up

migration-down:
	@migrate -database $(DB_LOCAL_URL) -path $(MIGRATIONS_DIR) down

migration-force:
	@migrate -database $(DB_LOCAL_URL) -path $(MIGRATIONS_DIR) force 1

migration-version:
	@migrate -database $(DB_LOCAL_URL) -path $(MIGRATIONS_DIR) version

migration-create:
	@if [ -z "$(name)" ]; then \
		echo "Need argument 'name=...'" && exit 1; \
	fi
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

build-image:
	docker buildx build --platform linux/amd64 -t iamvkosarev/book-shelf:latest --push .

push-image:
	docker push $(USER_NAME)/book-shelf:latest

apply-prod:
	kubectl apply -k k8s/overlays/prod
	kubectl -n book-shelf rollout restart deployment/book-shelf

apply-local:
	kubectl apply -k k8s/overlays/local
	kubectl -n book-shelf rollout restart deployment/book-shelf