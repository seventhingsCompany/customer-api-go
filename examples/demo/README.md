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

Or load the repository's local `.env` and build/run the executable:

```sh
go build -o bin/seventhings-demo ./examples/demo
set -a
source .env
set +a
./bin/seventhings-demo
```

The binary is written to `bin/seventhings-demo` (gitignored). Run `./bin/seventhings-demo -help`
for options. The default limit is **200 requests per minute** per client, evenly
spaced without bursts; all requests, including authentication and file downloads,
share it. Waiting respects context cancellation; concurrent waiters are served
in FIFO enqueue order, and canceled waiters do not consume a request slot.
If an operation returns HTTP 429 with a valid `Retry-After` header, the demo
prints the parsed delay before exiting with the error. It does not automatically
retry or intentionally trigger rate-limit responses on the live instance.

```sh
./bin/seventhings-demo -rate-limit 100 # Override requests per minute
./bin/seventhings-demo -rate-limit 0   # Disable throttling
```

## What it does

Run against a **test instance**: the demo creates and modifies temporary records.
On successful completion, temporary objects, persons, tasks, catalog items and rentals
are deleted. The new catalog/rental flows also attempt cleanup on errors. Uploaded
text and image files remain because the API has no file-delete endpoint. Existing
records are only read; no login users, schema fields or marketplace orders are created.

### 1. Auth

Logs in with password credentials and prints the user ID and
configured rate limit, without printing access tokens.

### 2. Objects — CRUD

- Lists the first page of objects (max 5)
- Creates a new object, then patches its name
- Looks it up by barcode and reads paginated history
- Archives and unarchives the object
- Lists report templates and generates a PDF for the temporary object, if a template exists
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
- Creates a task with omitted reminders and notifications disabled
- Verifies a zero-interval reminder (on the deadline), then an empty list (disabled)
- Closes the task, deletes it, and confirms a 404
- Cleans up the reference object

### 6. Persons

- Counts all persons and lists the first page (max 5)
- Looks up the first person by UUID and by numeric ID
- Creates a person, patches its department, deletes it, and confirms a 404
- Shows (commented out) how `PersonCreateUser` promotes a person to a login user

### 7. Product catalog — CRUD

Creates a temporary catalog item, reads it, updates its complete field state with
PUT, verifies a filtered list, deletes it and confirms 404. Requires the product
catalog feature and permissions; a 403 at creation is explicitly reported as skipped.

### 8. Timed rental — CRUD

Creates a temporary reference object and a timed rental with an offset timestamp.
Verifies equivalent timestamps after UTC normalization, extends the exclusive return
time, converts to a full-day rental and verifies that timestamps become null.
Deletes both rental and object. A renter is supplied because the current server
returns HTTP 500 when the spec's optional renter is omitted. SDK create/update calls
reject nil renters locally with `client.ErrRentalCaseRenterRequired` before sending HTTP.

### 9. Custom thumbnail

Uploads a generated 640×480 PNG, requests a 320×240 thumbnail, verifies dimensions
and checks that requesting the same size again returns identical cached bytes.
The image remains uploaded (no file-delete endpoint).

### 10. Resource history

Reads history for an existing room, location and rental when available. Persons,
rooms and locations use the flat API field-map contract; no envelope unwrapping
or person UUID-alias decoding is applied.

### 11. Auth cleanup

Revokes all tokens for the session.

## Expected output

```
── Auth ──────────────────────────────────────────
[Auth   ] Logging in…
[Auth   ] Logged in — user_id=1
[Rate   ] Limit: 200 requests/minute, evenly spaced; auth, uploads and downloads share it

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

── Persons ──────────────────────────────────────────
[Persons] Counting and listing persons…
[Persons] PersonsCount() → 12 person(s)
[Persons] Got 5 person(s) (page 1, max 5):
[Persons]   1. id=1 uuid=<uuid> Alice Smith <alice@example.com>
...
[Persons] Created person <uuid> <sdk.demo+…@example.com>
[Persons] Patched person <uuid> (department=IT)
[Persons] Deleted person <uuid>
[Persons] Confirmed deletion (404)

── Catalog ──────────────────────────────────────────
[Catalog] Verified create, get, PUT and filtered list — total=1
[Catalog] Deleted temporary item; confirmed 404

── Rentals ──────────────────────────────────────────
[Rentals] Verified timed rental — issue_at=… due_at=… (exclusive)
[Rentals] Verified extended return time and conversion to full-day; cleaning up rental and object

── Files ──────────────────────────────────────────
[Files  ] Verified 320x240 thumbnail and cached response (… bytes)

── Auth ──────────────────────────────────────────
[Auth   ] Revoking tokens…
[Auth   ] Tokens revoked

Done — all steps completed successfully.
```

The output above is abbreviated; record counts and timestamps vary by instance.
