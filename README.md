# Go Backend

A lightweight backend service written in Go, created just for fun and experimentation.

## Overview

A simple backend application with the following features(P.S. All features are still under development and not yet fully functional.):
- Login authentication
- Identity management
- management
- Game server control
- Cloud storage
- I'll think about it later. ;D

## Getting Started

1. Clone the repository  
   ```bash
   git clone https://github.com/carsupper665/web-server-backend.git
   cd web-server-backend
   ```

2. Create a `.env` file in the project root (see below for details).

3. Install dependencies and run:  
   ```bash
   go mod download
   go run main.go
   ```

4. The server will start on the port you specify in your `.env` (default `8080`).

---

## `.env` Configuration

Create a file named `.env` in the project root with the following variables:

```dotenv
# URL of your frontend application (e.g. http://localhost:3000)
FRONTEND_BASE_URL=

# Port for the backend to listen on
PORT=8080

# Enable debug logging (true or false)
DEBUG=false

# Session and cryptography secrets
SESSION_SECRET=your-secret-key
CRYPTO_SECRET=your-crypto-secret

# Paths and caching
SQLITE_PATH=path/to/db.sqlite
MEMORY_CACHE_ENABLED=false

# Timing settings (in seconds)
SYNC_FREQUENCY=60
BATCH_UPDATE_INTERVAL=300
RELAY_TIMEOUT=30

# Database connection pool settings:
#   SQL_MAX_IDLE_CONNS: maximum number of idle connections retained in the pool (default: 100)
SQL_MAX_IDLE_CONNS=100

#   SQL_MAX_OPEN_CONNS: maximum number of open connections to the database at once (default: 1000)
SQL_MAX_OPEN_CONNS=1000

#   SQL_MAX_LIFETIME: maximum time in seconds a connection may be reused before being closed (default: 60)
SQL_MAX_LIFETIME=60

# Automatically create a root/admin user on startup (true or false)
CREATE_ROOT_USER=true
```

### Variable Descriptions

- **FRONTEND_BASE_URL**  
  The base URL where your frontend application is hosted (e.g. `http://localhost:3000`). This is used to configure CORS and generate links.

- **PORT**  
  The TCP port on which the Go HTTP server listens (default: `3000`).

- **DEBUG**  
  When set to `true`, enables verbose logging and debug endpoints. Use `false` in production.

- **SESSION_SECRET**  
  A secret key used to sign and validate session cookies. Keep this secure and random.

- **CRYPTO_SECRET**  
  A secret used for encrypting and decrypting sensitive data.

- **SQLITE_PATH**  
  The file path to the SQLite database. Example: `./data/db.sqlite`.

- **MEMORY_CACHE_ENABLED**  
  Enable in-memory caching (`true` or `false`).

- **SYNC_FREQUENCY**  
  How often (in seconds) background sync tasks should run.

- **BATCH_UPDATE_INTERVAL**  
  Interval (in seconds) between batch update operations.

- **RELAY_TIMEOUT**  
  Timeout (in seconds) for relay operations before giving up.

- **SQL_MAX_IDLE_CONNS**  
  The maximum number of idle (unused) connections that the database connection pool will keep open.  
  Default: `100`.

- **SQL_MAX_OPEN_CONNS**  
  The maximum total number of open connections to your database.  
  Default: `1000`.

- **SQL_MAX_LIFETIME**  
  The maximum amount of time (in seconds) a connection may be reused before being closed and replaced.  
  Default: `60`.

- **CREATE_ROOT_USER**  
  If `true`, the application will automatically create a default root (admin) user on startup when none exists.  
  Set to `false` to disable automatic user creation.

---
## References
- This project is inspired by [QuantumNous/new-api](https://github.com/QuantumNous/new-api)
