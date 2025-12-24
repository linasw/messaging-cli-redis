# Messaging-cli 

Simple CLI to send messages to Redis and retrieve the result.

Using flags to get the arguments.

Using Redis as a messaging tool.

Using Watermill to send and retrieve the messages.

## How to run
1. `docker compose up` to run Redis locally
2. `go run cmd/consumer/main.go` to start the consumer
3. `go build cmd/cli/main.go` to build the main producer
   1. `./main.exe order-created -orderId "ord1" -productIds "prod1" -productIds "prod2"`
   2. `./main.exe order-completed -orderId "ord1"`