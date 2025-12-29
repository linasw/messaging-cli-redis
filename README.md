# Messaging-cli 

Simple CLI to send messages to Redis and retrieve the result.

Using flags to get the arguments.

Using Redis as a messaging tool.

Using Postgres as DB to store orders.

Using Watermill to send and retrieve the messages.

## How to run
1. `docker compose up` to run Redis & Postgres locally
2. `docker exec -i messaging-cli-postgres psql -U orderuser -d orderdb < migrations/001_create_orders.up.sql` to setup the DB
3. `go run cmd/consumer/main.go` to start the consumer
4. `go build cmd/cli/main.go` to build the main producer
   1. `./main.exe order-created -orderId "ord1" -productIds "prod1" -productIds "prod2"` to create a message in 'order-created' topic
   2. `./main.exe order-completed -orderId "ord1"` to create a message within 'order-completed' topic

## How to test
1. `docker compose up` to run Redis & Postgres locally
2. `docker exec -i messaging-cli-postgres psql -U orderuser -d orderdb < migrations/001_create_orders.up.sql` to setup the DB
3. `go test -v ./test/order_repository_test.go` to run repo's unit test
4. `go test -v ./test/integration_test.go` to run integration tests (NOTE: The tests fail if run with this command. Need to fix, but as a workaround, you can run the test one by one and then they work:)
   1. `go test -v ./test -run TestOrderCreatedFlow`
   2. `go test -v ./test -run TestOrderCompletedFlow`
   3. `go test -v ./test -run TestFullOrderFlow`