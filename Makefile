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
	docker buildx build --platform linux/amd64 -t $(USER_NAME)/book-shelf:latest --push .

build-migration-image:
	docker buildx build --platform linux/amd64 -t $(USER_NAME)/book-shelf-migrations:latest --push -f Dockerfile.migrations .

apply-prod:
	kubectl -n book-shelf delete job book-shelf-migrate --ignore-not-found
	kubectl apply -k k8s/overlays/prod
	kubectl -n book-shelf wait --for=condition=complete job/book-shelf-migrate --timeout=180s
	kubectl -n book-shelf rollout restart deployment/book-shelf
	kubectl -n book-shelf rollout status deployment/book-shelf

apply-local:
	kubectl apply -k k8s/overlays/local
	kubectl -n book-shelf rollout restart deployment/book-shelf

gen-secrets:
	openssl genrsa -out secrets/jwt_private.pem 2048
	openssl rsa -in secrets/jwt_private.pem -pubout -out secrets/jwt_public.pem
	awk 'BEGIN{printf "PUBLIC_KEY=\""} {gsub(/\r/,""); printf "%s\\n",$0} END{print "\""}' secrets/jwt_public.pem
	awk 'BEGIN{printf "PRIVATE_KEY=\""} {gsub(/\r/,""); printf "%s\\n",$0} END{print "\""}' secrets/jwt_private.pem