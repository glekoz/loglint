# loglint

Static analysis linter for Go that enforces logging message conventions.

## Rules

### 1. Lowercase first letter

Log messages must start with a lowercase letter.

```go
// ❌ Wrong
log.Info("Starting server on port 8080")
slog.Error("Failed to connect to database")

// ✅ Correct
log.Info("starting server on port 8080")
slog.Error("failed to connect to database")
```

### 2. English only

Log messages must be written in English.

```go
// ❌ Wrong
log.Info("запуск сервера")
log.Error("ошибка подключения к базе данных")

// ✅ Correct
log.Info("starting server")
log.Error("failed to connect to database")
```

### 3. No special symbols or emoji

Log messages must not contain punctuation, special characters, or emoji.

```go
// ❌ Wrong
log.Info("server started!🚀")
log.Error("connection failed!!!")
log.Warn("warning: something went wrong...")

// ✅ Correct
log.Info("server started")
log.Error("connection failed")
log.Warn("something went wrong")
```

### 4. No sensitive data

Log messages and concatenated variables must not contain potentially sensitive information (passwords, tokens, keys, etc.).

```go
// ❌ Wrong
log.Info("user password: " + password)
log.Debug("api_key=" + apiKey)
log.Info("token: " + token)

// ✅ Correct
log.Info("user authenticated successfully")
log.Debug("api request completed")
log.Info("token validated")
```

## Supported loggers

| Package | Methods |
|---------|---------|
| `log/slog` | `Debug`, `DebugContext`, `Error`, `ErrorContext`, `Info`, `InfoContext`, `Warn`, `WarnContext` |
| `go.uber.org/zap` | `Debug`, `DPanic`, `Error`, `Fatal`, `Info`, `Panic`, `Warn` |

Custom loggers can be added via [config file](#configuration-file).

---

## Installation

### Standalone binary

```bash
go install github.com/glekoz/loglint/cmd/loglint@latest
```

### As a golangci-lint plugin (Module Plugin System)

golangci-lint supports the **Module Plugin System** which embeds your linter directly into a custom-built golangci-lint binary — no `.so` files, works on all platforms including Windows.

**Requirements:** Go, git, `golangci-lint` installed.

---

**1. Add `.custom-gcl.yml` to your project:**

```yaml
version: v2.1.6   # your golangci-lint version

plugins:
  - module: "github.com/glekoz/loglint"
    import: "github.com/glekoz/loglint/plugin"
    version: v1.0.0   # or use `path: ./` for a local checkout
```

**2. Configure `.golangci.yml`:**

```yaml
linters-settings:
  custom:
    loglint:
      type: "module"
      description: Checks logging message conventions
      original-url: github.com/glekoz/loglint
      settings:
        config: .loglint.yml   # optional: path to loglint config

linters:
  enable:
    - loglint
```

**3. Build the custom binary (once, or after updating the plugin):**

```bash
golangci-lint custom
# generates ./custom-gcl
```

**4. Run:**

```bash
./custom-gcl run ./...
```

---

## Usage

### Standalone

```bash
# Analyse a package
loglint ./...

# With config file
loglint -loglint.config=.loglint.yml ./...
```

### golangci-lint (Module Plugin)

```bash
# Build the custom binary (once)
golangci-lint custom

# Run
./custom-gcl run ./...
```

To pass the config file path to the plugin, set the `settings.config` key in `.golangci.yml`:

```yaml
linters-settings:
  custom:
    loglint:
      type: "module"
      description: Checks logging message conventions
      original-url: github.com/glekoz/loglint
      settings:
        config: .loglint.yml
```

---

## Configuration file

Create a `.loglint.yml` file in the root of your project.

```yaml
rules:
  # Rule 1: log message must start with a lowercase letter (default: true)
  lowercase: true

  # Rule 2: log message must be in English only (default: true)
  english_only: true

  # Rule 3: log message must not contain special symbols or emoji (default: true)
  no_special_symbols: true

  # Rule 4: log message must not contain sensitive data (default: true)
  no_sensitive_data: true

# Override the list of sensitive keywords (default list shown below)
sensitive_keywords:
  - password
  - secret
  - token
  - key
  - credential
  - auth
  - login
  - pass
  - pwd

# Words that contain a sensitive keyword but are safe to use
keywords_whitelist:
  - passwordik

# Single characters that are allowed in log messages despite being non-alphanumeric
symbols_whitelist:
  - "-"
  - "."

# Override the set of recognised logger packages and their methods
loggers:
  "log/slog":
    - Debug
    - DebugContext
    - Error
    - ErrorContext
    - Info
    - InfoContext
    - Warn
    - WarnContext
  "go.uber.org/zap":
    - Debug
    - DPanic
    - Error
    - Fatal
    - Info
    - Panic
    - Warn
  # Add your own logger
  "github.com/myorg/mylogger":
    - Info
    - Error
    - Warn
```

All keys are optional. If a key is omitted, the default value is used.

---

## Development

```bash
# Run tests
go test ./...

# Build standalone binary
go build -o loglint ./cmd/loglint/

# Build custom golangci-lint binary with loglint embedded
golangci-lint custom          # reads .custom-gcl.yml, outputs ./custom-gcl
```
