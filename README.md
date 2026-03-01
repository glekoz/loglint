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

// ✅ Correct
log.Info("user authenticated successfully")
log.Debug("api request completed")
```

## Supported loggers

| Package | Methods |
|---------|---------|
| `log/slog` | `Debug`, `DebugContext`, `Error`, `ErrorContext`, `Info`, `InfoContext`, `Warn`, `WarnContext` |
| `go.uber.org/zap` | `Debug`, `DPanic`, `Error`, `Fatal`, `Info`, `Panic`, `Warn` |

Custom loggers can be added via [config file](#loglint-yml----for-standalone-binary).

---

## Suggested fixes

For some violations loglint provides an **automatic fix** that editors and tools can apply directly.

| Rule | Fix |
|------|-----|
| Uppercase first letter | Converts the entire message to lowercase |
| Non-alphanumeric symbol | Removes all offending characters from the message |
| Sensitive variable in concatenation | Replaces the variable with the literal `"credentials removed"` |

Fixes are applied automatically in editors with `gopls` support (VS Code, GoLand, etc.) via the quick-fix menu, or from the command line:

```bash
# standalone binary
loglint -fix ./...

# custom golangci-lint binary
./custom-gcl run --fix ./...
```

Example:

```go
// Before
slog.Info("Starting server!")
// After (fix: convert first letter to lowercase + remove non-alphanumeric symbol)
slog.Info("starting server")

// Before
slog.Warn("retrying" + password)
// After (fix: remove sensitive variable)
slog.Warn("retrying" + "credentials removed")
```

> Fixes for sensitive data in string literals (e.g. `"user password: ..."`) and non-English characters are reported as diagnostics only — no automatic fix is provided, as the correct replacement depends on context.

---

## Installation

### Standalone binary

```bash
go install github.com/glekoz/loglint/cmd/loglint@latest
```

### As a golangci-lint plugin (Module Plugin System)

golangci-lint supports the **Module Plugin System** which embeds your linter directly into a custom-built golangci-lint binary.

**Requirements:** Go, git, `golangci-lint` installed.

---

**1. Add `.custom-gcl.yml` to your project:**

```yaml
version: v2.10.1   # your golangci-lint version

plugins:
  - module: "github.com/glekoz/loglint"
    import: "github.com/glekoz/loglint/plugin"
    version: v0.1.3   # actual version
```

**2. Configure `.golangci.yml`:**

```yaml
version: "2"

linters:
  default: none
  enable:
    - loglint
  settings:
    custom:
      loglint:
        type: "module"
        description: Checks logging message conventions (lowercase, English-only, no special symbols, no sensitive data)
        original-url: github.com/glekoz/loglint
        settings:
          rules:
            lowercase: true
            english_only: true
            no_special_symbols: true
            no_sensitive_data: true
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
          keywords_whitelist: []
          symbols_whitelist: []
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

### Standalone binary

```bash
# Analyse a package with default settings
loglint ./...

# With a config file
loglint -loglint.config=.loglint.yml ./...
```

### golangci-lint (Module Plugin)

```bash
# Build the custom binary (once)
golangci-lint custom

# Run
./custom-gcl run ./...
```

Settings are specified **inline** inside `.golangci.yml` under `settings.custom.loglint.settings` — no separate config file is needed. See the full example in the [Installation](#as-a-golangci-lint-plugin-module-plugin-system) section above.

> **Note:** the `loggers` option (overriding which packages and methods are checked) is only available when using the standalone binary with a `.loglint.yml` config file.

---

## Configuration

### `.loglint.yml` — for standalone binary

Pass the path with `-loglint.config=.loglint.yml`.

```yaml
rules:
  # Log message must start with a lowercase letter (default: true)
  lowercase: true

  # Log message must be in English only (default: true)
  english_only: true

  # Log message must not contain special symbols or emoji (default: true)
  no_special_symbols: true

  # Log message must not contain sensitive data (default: true)
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

### `.golangci.yml` — for golangci-lint plugin

All loglint settings are placed inline under `settings.custom.loglint.settings`.

```yaml
version: "2"

linters:
  default: none
  enable:
    - loglint
  settings:
    custom:
      loglint:
        type: "module"
        description: Checks logging message conventions (lowercase, English-only, no special symbols, no sensitive data)
        original-url: github.com/glekoz/loglint
        settings:
          rules:
            lowercase: true
            english_only: true
            no_special_symbols: true
            no_sensitive_data: true
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
          keywords_whitelist: []
          symbols_whitelist:
            - "-"
            - "."
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
