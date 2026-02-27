# Demo

An end-to-end example that exercises the core seventhings Go SDK modules against a live instance. It is designed to verify that the SDK works correctly and to show new users how each API area is used.

## Prerequisites

- Go 1.25+
- A running seventhings instance with valid credentials and the REST API Integration active

## Usage

Set the required environment variables and run:

```sh
export SEVENTHINGS_BASE_URL=https://example.seventhings.com
export SEVENTHINGS_USERNAME=user@example.com
export SEVENTHINGS_PASSWORD=secret
export SEVENTHINGS_CLIENT_ID=your-client-id

go run ./examples/demo/
```

## What it does

The demo runs the following steps in order. Every resource it creates is cleaned up before the program exits.

### 1. Auth

Logs in with the provided credentials and prints the user ID and a truncated access token.

### 2. Objects — CRUD

- Lists the first page of objects (max 5)
- Creates a new object, then patches its name
- Archives and unarchives the object
- Deletes the object and confirms a 404 response

### 3. Objects — Sorting & Filtering

- Fetches the **5 most recently changed** assets by sorting on `updated_at DESC`
- Filters assets whose `inventory_name` contains "SDK" using the `like` operator

These two calls demonstrate how to use `ListOptions` with `Sort` and `Filters`.

### 4. Files

- Uploads a small text file
- Reads back its metadata (name, type, size)
- Creates a temporary object, attaches the file, then detaches it
- Cleans up the temporary object

### 5. Tasks

- Looks up the current user's UUID
- Creates a temporary object to use as a task reference
- Creates a task with a deadline, assignee, reference, and reminder
- Closes the task, deletes it, and confirms a 404
- Cleans up the reference object

### 6. Auth cleanup

Revokes all tokens for the session.

## Expected output

```
── Auth ──────────────────────────────────────────
[Auth   ] Logging in…
[Auth   ] Logged in — user_id=1, token=eyJ0eXAiOiJKV1QiLCJh…

── Objects ──────────────────────────────────────────
[Objects] Listing objects…
[Objects] Listed 5 object(s) (first page, max 5)
[Objects] Created object <uuid>
[Objects] Patched object — inventory_name=SDK Demo Object (updated)
[Objects] Archived object <uuid>
[Objects] Unarchived object <uuid>
[Objects] Deleted object <uuid>
[Objects] Confirmed deletion (404)

── Objects ──────────────────────────────────────────
[Objects] Fetching last 5 changed assets (sorted + filtered)…
[Objects] Got 5 recently changed asset(s):
[Objects]   1. Some Asset (updated_at=2026-02-27 00:10:33)
[Objects]   ...

── Objects ──────────────────────────────────────────
[Objects] Filtering assets by name containing "SDK"…
[Objects] Got 2 asset(s) matching filter:
[Objects]   1. SDK Test
[Objects]   2. SDK Test 2

── Files ──────────────────────────────────────────
[Files  ] Uploading file…
[Files  ] Uploaded file <uuid> (demo.txt, 40 bytes)
...

── Tasks ──────────────────────────────────────────
[Tasks  ] Creating task…
...

── Auth ──────────────────────────────────────────
[Auth   ] Revoking tokens…
[Auth   ] Tokens revoked

Done — all steps completed successfully.
```
