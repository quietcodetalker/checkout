export MY_IP=0.0.0.0
up:
	docker-compose -f ./deployments/docker-compose.yml up -d
down:
	docker-compose -f ./deployments/docker-compose.yml down
ps:
	docker-compose -f ./deployments/docker-compose.yml ps

stock_migrate_up:
	goose \
	-dir ./db/migrations/stock/ \
	postgres \
	"user=postgres password=postgres dbname=stock sslmode=disable host=localhost port=5433" \
	up

stock_migrate_down:
	goose \
	-dir ./db/migrations/stock/ \
	postgres \
	"user=postgres password=postgres dbname=stock sslmode=disable host=localhost port=5433" \
	down

billing_migrate_up:
	goose \
	-dir ./db/migrations/billing/ \
	postgres \
	"user=postgres password=postgres dbname=billing sslmode=disable host=localhost port=5434" \
	up

billing_migrate_down:
	goose \
	-dir ./db/migrations/billing/ \
	postgres \
	"user=postgres password=postgres dbname=billing sslmode=disable host=localhost port=5434" \
	down

orders_migrate_up:
	goose \
	-dir ./db/migrations/orders/ \
	postgres \
	"user=postgres password=postgres dbname=orders sslmode=disable host=localhost port=5435" \
	up

orders_migrate_down:
	goose \
	-dir ./db/migrations/orders/ \
	postgres \
	"user=postgres password=postgres dbname=orders sslmode=disable host=localhost port=5435" \
	down

notifications_migrate_up:
	goose \
	-dir ./db/migrations/notifications/ \
	postgres \
	"user=postgres password=postgres dbname=notifications sslmode=disable host=localhost port=5436" \
	up

notifications_migrate_down:
	goose \
	-dir ./db/migrations/notifications/ \
	postgres \
	"user=postgres password=postgres dbname=notifications sslmode=disable host=localhost port=5436" \
	down

migrate_up:
	make orders_migrate_up && \
	make stock_migrate_up && \
	make billing_migrate_up && \
	make notifications_migrate_up

migrate_down:
	make orders_migrate_down && \
	make stock_migrate_down && \
	make billing_migrate_down && \
	make notifications_migrate_down

stock_run:
	go run ./cmd/stock

billing_run:
	go run ./cmd/billing

orders_run:
	go run ./cmd/orders

notifications_run:
	go run ./cmd/notifications

seed:
	go run ./cmd/seed

build_stock_image:
	DOCKER_BUILDKIT=0 docker build \
	-t gitlab-registry.ozon.dev/unknownspacewalker/homework3/stock:latest \
	--tag stock:latest \
	-f ./deployments/stock/Dockerfile .

build_billing_image:
	DOCKER_BUILDKIT=0 docker build \
	-t gitlab-registry.ozon.dev/unknownspacewalker/homework3/billing:latest \
	--tag billing:latest \
	-f ./deployments/billing/Dockerfile .

build_orders_image:
	DOCKER_BUILDKIT=0 docker build \
	-t gitlab-registry.ozon.dev/unknownspacewalker/homework3/orders:latest \
	--tag orders:latest \
	-f ./deployments/orders/Dockerfile .

build_notifications_image:
	DOCKER_BUILDKIT=0 docker build \
	-t gitlab-registry.ozon.dev/unknownspacewalker/homework3/notifications:latest \
	--tag notifications:latest \
	-f ./deployments/notifications/Dockerfile .

build_ntrigger_image:
	DOCKER_BUILDKIT=0 docker build \
	-t gitlab-registry.ozon.dev/unknownspacewalker/homework3/ntrigger:latest \
	--tag ntrigger:latest \
	-f ./deployments/ntrigger/Dockerfile .

build_all_images:
	make build_orders_image && \
	make build_stock_image && \
	make build_billing_image && \
	make build_notifications_image

TOPIC=foo
create_topic:
	docker run \
	--net=host \
	--rm \
	confluentinc/cp-kafka:5.0.0 kafka-topics \
	--create \
	--topic ${TOPIC} \
	--partitions 4 \
	--replication-factor 2 \
	--if-not-exists \
	--zookeeper localhost:32181
