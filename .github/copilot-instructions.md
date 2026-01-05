# Copilot / AI Agent Instructions for Peril

This repo is a small Pub/Sub demo (RabbitMQ) with a server CLI and client CLI. Keep guidance focused on concrete patterns, run steps, and integration points so an AI agent can be productive immediately.

- **Big picture**: Two CLIs communicate over RabbitMQ using two exchanges:
  - Direct control messages: `peril_direct` (see [internal/routing/routing.go](internal/routing/routing.go))
  - Topic messages: `peril_topic` for game events like army moves
  - DLX: `peril_dlx` is set as the dead-letter exchange in queue declarations

- **Primary components**:
  - Server: [cmd/server/main.go](cmd/server/main.go) — connects to RabbitMQ, publishes pause/resume messages, and declares `game_logs` queue.
  - Client: [cmd/client/main.go](cmd/client/main.go) — prompts username, subscribes to pause and topic queues, publishes army moves.
  - Pub/Sub helpers: [internal/pubsub/publish.go](internal/pubsub/publish.go) and [internal/pubsub/subscribe.go](internal/pubsub/subscribe.go) — single place to declare/bind queues, publish JSON, and subscribe with typed handlers.
  - Routing models and constants: [internal/routing](internal/routing) — canonical routing keys and exchange names.
  - Game logic and models: [internal/gamelogic](internal/gamelogic) — `GameState`, `ArmyMove`, spawn/move commands, and CLI helpers.

- **How to run locally (dev)**:
  - Start RabbitMQ (management image recommended):
    ```bash
    docker run -it --rm --name rabbit -p 5672:5672 -p 15672:15672 rabbitmq:3-management
    ```
  - Run server: `go run ./cmd/server`
  - Run client: `go run ./cmd/client`

- **Message shapes and keys**:
  - Pause messages: `routing.PlayingState` (see [internal/routing/models.go](internal/routing/models.go)), routing key `pause` → queue names use `pause.<username>` for client bindings.
  - Army moves: `gamelogic.ArmyMove` (see [internal/gamelogic/gamedata.go](internal/gamelogic/gamedata.go)), published on `peril_topic` with keys like `army_moves.<username>`; subscribers bind to `army_moves.*`.

- **Patterns an AI should follow when editing code**:
  - Use `pubsub.PublishJSON[T]` and `pubsub.SubscribeJSON[T]` for all JSON messages to keep serialization consistent.
  - Handlers passed to `SubscribeJSON` must be of type `func(T) pubsub.AckType` and return one of `pubsub.Ack`, `pubsub.NackRequeue`, or `pubsub.NackDiscard` (see [internal/pubsub/subscribe.go](internal/pubsub/subscribe.go)).
  - Queue lifetimes: client queues use `pubsub.Transient` (auto-delete, exclusive) while server logs may be `pubsub.Durable`.
  - `DeclareAndBind` sets a dead-letter exchange `peril_dlx`. Be careful when changing queue args.

- **Conventions and gotchas**:
  - `SubscribeJSON` spawns a goroutine that processes deliveries and acks/nacks — handlers must be fast and should not block long-running work.
  - Queue names are constructed in code (examples in [cmd/client/main.go](cmd/client/main.go)) — prefer the existing fmt.Sprintf patterns (e.g., `fmt.Sprintf("army_moves.%s", username)`).
  - Tests and CI: none present. Keep changes small and runnable locally using the Docker RabbitMQ above.

- **When changing messaging behaviour**:
  - Update both producer and consumer code paths: check [internal/pubsub] helpers, `cmd/*` usages, and types under [internal/routing] and [internal/gamelogic].

- **Examples to copy/paste**:
  - Subscribe to moves: `pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, fmt.Sprintf("army_moves.%s", username), "army_moves.*", pubsub.Transient, handlerMove(gs))`
  - Publish a pause: `pubsub.PublishJSON(chan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})`

If any section is unclear or you want more explicit handler examples or CI/run scripts, tell me which area to expand.
