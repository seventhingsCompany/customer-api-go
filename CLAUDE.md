# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go SDK client library for the seventhings customer API (`github.com/SeventhingsCompany/customer-api-go`). Pure Go, zero external dependencies (only stdlib). Requires Go 1.25+.

## Commands

```sh
# Run all unit tests
go test ./...

# Run a single test
go test ./client -run TestObjectCreate

# Run integration tests (requires live seventhings instance)
SEVENTHINGS_BASE_URL=https://example.seventhings.com \
SEVENTHINGS_USERNAME=user@example.com \
SEVENTHINGS_PASSWORD=password \
SEVENTHINGS_CLIENT_ID=client-id \
go test -tags=integration -v ./...
```

## Architecture

Two packages:

- **`client/`** — `Client` struct with all API methods. Wraps `net/http` with Bearer token auth. Constructor variants: `New`, `NewWithToken`, `NewWithCredentials`. Low-level HTTP helpers (`Get`, `Post`, `Patch`, `Put`, `Delete`, `GetRaw`, `PostMultipart`) are public and reusable. Each domain (objects, tasks, files, rentals, locations, rooms, users, field definitions, circularity hub, auth, ping) is in its own file.

- **`models/`** — Request/response types, enums, and `APIError`. `ListOptions` handles pagination, sorting, and deep-object filter encoding for the PHP-style API query format.

### Key Patterns

- **Create endpoints** return a `Location` header; use `UUIDFromLocationHeader(resp)` to extract the new resource UUID. CircularityHub uses `IntFromLocationIDHeader(resp)` for integer IDs instead.
- **All API errors** are returned as `*models.APIError` (check with `errors.As`).
- **Tests** use `httptest.NewServer` with a `newTestClient(t, server)` helper defined in `client/client_test.go`. No mocking frameworks.
- **Integration tests** are gated behind `//go:build integration` tag and require four `SEVENTHINGS_*` environment variables.
- **Objects, locations, rooms** use `map[string]any` for fields (schema is dynamic). Tasks, rentals, field definitions use typed structs from `models/`.
