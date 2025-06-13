# minecraft-server-controller

Skeleton project for managing a Minecraft server with a Go backend.

## Structure

```
cmd/server           - program entrypoint
internal/router      - route definitions
internal/controllers - HTTP handlers
internal/services    - placeholder business logic
internal/middleware  - authentication middleware
```

## Requirements

- Go 1.20+

## Run

```
cd cmd/server
go run .
```

This starts a Gin server on `:8080` with stub endpoints for managing the Minecraft server.

