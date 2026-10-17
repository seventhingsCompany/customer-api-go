# seventhings Go SDK

Go client library for the [seventhings](https://seventhings.com) customer API. See the [API documentation](https://api.seventhings.com/customer-api/v1/) for details on available endpoints.

Targets API description `v0.20143_c5f8b54b3_20261005`.

## Installation

```sh
go get github.com/SeventhingsCompany/customer-api-go
```

## Quick Start

### Password Authentication

```go
c, err := client.NewWithCredentials(ctx, "https://example.seventhings.com", "user@example.com", "password", "client-id")
if err != nil {
    log.Fatal(err)
}
```

### Pre-existing Token

```go
c := client.NewWithToken("https://example.seventhings.com", "my-jwt-token")
```

### Manual Login

```go
c := client.New("https://example.seventhings.com")
tok, err := c.Login(ctx, "user@example.com", "password", "client-id")
if err != nil {
    log.Fatal(err)
}
// tok.RefreshToken can be stored for later use with c.Refresh()
```

### SSO Authentication

```go
c := client.New("https://example.seventhings.com")
tok, err := c.LoginSSO(ctx, models.SSOProviderAzure, "auth-code", "client-id", nil)
```

## Usage

### Ping

```go
ping, err := c.Ping(ctx)
// ping.Status == "ok"
```

### Objects

```go
// Create
uuid, err := c.ObjectCreate(ctx, map[string]any{"name": "Laptop", "serial": "SN-123"})

// Get
obj, err := c.ObjectGet(ctx, uuid)

// Get by barcode (pass the raw scancode, without URL encoding)
obj, err = c.ObjectGetByBarcode(ctx, "INV/10001")

// List with filtering
objects, err := c.ObjectsList(ctx, &models.ListOptions{
    Page:    1,
    PerPage: 25,
    Sort:    map[string]models.SortDirection{"name": models.SortASC},
    Filters: []models.FilterEntry{
        {Field: "name", Operator: models.FilterLike, Values: []string{"Laptop"}},
    },
})

// Update
updated, err := c.ObjectPatch(ctx, uuid, map[string]any{"name": "Updated Laptop"})

// Delete
err = c.ObjectDelete(ctx, uuid)

// Archive / Unarchive
err = c.ObjectArchive(ctx, uuid)
err = c.ObjectUnarchive(ctx, uuid)

// Attach files to an object
resp, err := c.ObjectAddFiles(ctx, uuid, []models.FileAttachment{
    {FieldKey: "photo", FileUUID: "file-uuid"},
})

// Remove files from an object
resp, err = c.ObjectRemoveFiles(ctx, uuid, []models.FileAttachment{
    {FieldKey: "photo", FileUUID: "file-uuid"},
})
```

### History

History endpoints return recorded changes newest first, together with pagination metadata:

```go
history, err := c.ObjectHistory(ctx, "object-uuid", &models.HistoryListOptions{
    Page:    1,
    PerPage: 50,
})
// history.Items, history.Page, history.PerPage, history.Total
```

Also available: `RoomHistory`, `LocationHistory`, `PersonHistory`, `TaskHistory`, and
`RentalCaseHistory`, with the same UUID and options arguments. Passing `nil` options
uses the API defaults (page 1, 50 entries per page; maximum 200).

Object history entries are dynamic maps that preserve asset, task, rental-case, and
object-merge payloads. Other history entries are typed structs; their `Details`
field contains a JSON string or an empty string, and `OccurredAt` is a UTC timestamp.

### PDF Reports

```go
templates, err := c.ReportTemplatesList(ctx)
pdf, err := c.ReportCreate(ctx, models.CreateReport{
    ReportTemplateUUID: "template-uuid",
    ObjectUUIDs:        []string{"object-uuid-1", "object-uuid-2"},
})
// pdf contains the generated PDF bytes; the API does not store the document.
```

### Files

```go
// Upload
uuid, err := c.FileUpload(ctx, "photo.jpg", fileReader)

// Get metadata
file, err := c.FileGet(ctx, uuid)

// Download data
data, err := c.FileGetData(ctx, uuid)

// Download thumbnail
thumb, err := c.FileGetThumbnail(ctx, uuid)

// Custom size (1–2000 pixels per dimension, preserving aspect ratio)
thumb, err = c.FileGetThumbnailWithSize(ctx, uuid, 320, 240)

// List all files
files, err := c.FilesList(ctx)
```

### Tasks

For create/update requests, nil `Reminders` is omitted; an explicit empty
`[]models.TimeInterval{}` disables reminders. A zero interval reminds on the deadline.

```go
uuid, err := c.TaskCreate(ctx, models.CreateTask{Title: "Inspect server room"})
task, err := c.TaskGet(ctx, uuid)
tasks, err := c.TasksList(ctx, &models.TaskListOptions{Status: &open})
err = c.TaskUpdate(ctx, uuid, models.UpdateTask{Title: "Updated title"})
err = c.TaskUpdateStatus(ctx, uuid, models.TaskStatusClosed)
err = c.TaskDelete(ctx, uuid)
```

### Rental Cases

`CreateRentalCase` and `UpdateRentalCase` support `IsFullDay` (`*bool`),
`IssueAt` and `DueAt` (`*string`, RFC 3339 timestamps). Explicit `false` selects
a timed rental; the return time is exclusive. Leave these fields nil to use
the API's inference/preservation rules. Responses include all three fields.

**`RentalCaseUpdate` is a full replacement (PUT), not a partial patch.** Send all
required fields and any optional values you want to retain. Both create and update
require a non-nil `Renter` in the SDK: the current server returns HTTP 500 when it
is omitted, despite the API description marking it optional. Missing renters return
`client.ErrRentalCaseRenterRequired` locally without an HTTP request (detectable with
`errors.Is`).

```go
references := []models.RentalCaseReferenceInput{
    {Type: models.RentalCaseReferenceTypeAsset, UUID: "object-uuid"},
}
renter := &models.RentalCaseRenter{Type: models.RenterTypePlain, Value: "John"}
uuid, err := c.RentalCaseCreate(ctx, models.CreateRentalCase{
    Title:               "Equipment loan",
    Renter:              renter,
    References:          references,
    IssueDate:           "2099-01-01",
    DueDate:             "2099-01-02",
    ResponsibleUserUUID: "responsible-user-uuid",
    Attachments:         []string{},
})
if err != nil {
    log.Fatal(err)
}

// Send the complete desired state, even when changing only the comment.
err = c.RentalCaseUpdate(ctx, uuid, models.UpdateRentalCase{
    Title:               "Equipment loan",
    Renter:              renter,
    References:          references,
    IssueDate:           "2099-01-01",
    DueDate:             "2099-01-02",
    ResponsibleUserUUID: "responsible-user-uuid",
    Attachments:         []string{},
    Comment:             "Updated pickup instructions",
})
if err != nil {
    log.Fatal(err)
}
err = c.RentalCaseDelete(ctx, uuid)
```

### Locations

```go
locations, err := c.LocationsList(ctx, &models.ListOptions{Page: 1, PerPage: 25})
uuid, err := c.LocationCreate(ctx, map[string]any{"name": "Building A"})
loc, err := c.LocationGet(ctx, uuid)
updated, err := c.LocationPatch(ctx, uuid, map[string]any{"name": "Building B"})
err = c.LocationDelete(ctx, uuid)
count, err := c.LocationsCount(ctx, nil)
```

### Rooms

```go
rooms, err := c.RoomsList(ctx, &models.ListOptions{Page: 1, PerPage: 25})
uuid, err := c.RoomCreate(ctx, map[string]any{"name": "Server Room"})
room, err := c.RoomGet(ctx, uuid)
updated, err := c.RoomPatch(ctx, uuid, map[string]any{"name": "Lab"})
err = c.RoomDelete(ctx, uuid)
count, err := c.RoomsCount(ctx, nil)
```

### Persons and Flat Field Maps

Persons, rooms, and locations are decoded as flat template-defined field maps,
without `{uuid, fields}` envelope unwrapping. Person typed convenience fields
use their field keys (including `person_uuid`); no `uuid` alias is applied.
`Person.Fields` preserves the entire flat map, including custom fields and attachments.

### Product Catalog

Requires the instance's product catalog feature and appropriate permissions.

```go
page, err := c.CatalogItemsList(ctx, &models.ListOptions{Page: 1, PerPage: 25})
item, err := c.CatalogItemGet(ctx, "catalog-item-uuid")
uuid, err := c.CatalogItemCreate(ctx, map[string]any{
    "catalog_object_inventory_name": "Laptop",
})
err = c.CatalogItemUpdate(ctx, uuid, item.Fields)
err = c.CatalogItemDelete(ctx, uuid)
// page.Items, page.Page, page.PerPage, page.Total, page.Sort
```

Updates use PUT and replace the complete field state: omitted fields are emptied.
Object field keys use `catalog_object_`; catalog metadata keys have no prefix.
Updating or deleting a catalog item does not modify objects created from it.

### Users

```go
resp, err := c.UsersList(ctx, &models.UserListOptions{Page: intPtr(1)})
user, err := c.UserGet(ctx, "user-uuid")
user, err := c.UserGetByID(ctx, 42)
```

### Field Definitions

```go
defs, err := c.FieldDefinitionsList(ctx, models.AssetTrackingTemplateAsset)
uuid, err := c.FieldDefinitionCreate(ctx, models.AssetTrackingTemplateAsset, models.CreateFieldDefinition{
    Label:     "Color",
    FieldType: models.FieldDefinitionFieldType{Name: models.FieldTypeText},
})
def, err := c.FieldDefinitionGet(ctx, models.AssetTrackingTemplateAsset, uuid)
err = c.FieldDefinitionUpdate(ctx, models.AssetTrackingTemplateAsset, uuid, input)
```

### CircularityHub

```go
// Suggest category
suggestions, err := c.CircularityHubSuggestCategory(ctx, models.FilterObject{})

// Suggest rest price
prices, err := c.CircularityHubSuggestRestPrice(ctx, map[string]string{"category": "Electronics"})

// Add objects in bulk
err = c.CircularityHubAddObjects(ctx, map[string]models.AddObjectEntry{
    "obj-uuid": {Category: "Chair", Price: "50.00"},
})

// Items (integer IDs)
items, err := c.CircularityHubItemsList(ctx, &models.ListOptions{PerPage: 25})
item, err := c.CircularityHubItemGet(ctx, 42)
err = c.CircularityHubItemUpdate(ctx, 42, map[string]any{"price": "99.00"})
err = c.CircularityHubItemDelete(ctx, 42)

// Orders (integer IDs)
orders, err := c.CircularityHubOrdersList(ctx, nil)
orderID, err := c.CircularityHubOrderCreate(ctx, []int{1, 2, 3})
order, err := c.CircularityHubOrderGet(ctx, orderID)
err = c.CircularityHubOrderUpdate(ctx, orderID, map[string]any{"completed": true})
```

## Filtering & Sorting

List endpoints that accept `*models.ListOptions` support deep-object query parameter encoding:

```go
opts := &models.ListOptions{
    Page:    1,
    PerPage: 50,
    Sort: map[string]models.SortDirection{
        "name":       models.SortASC,
        "created_at": models.SortDESC,
    },
    Filters: []models.FilterEntry{
        {Field: "name", Operator: models.FilterLike, Values: []string{"Laptop"}},
        {Field: "status", Operator: models.FilterIn, Values: []string{"active", "pending"}},
        {Field: "price", Operator: models.FilterGte, Values: []string{"100"}},
    },
}
```

This encodes to:
```
page=1&per_page=50&sort[name]=ASC&sort[created_at]=DESC&filter[name][like][]=Laptop&filter[status][in][]=active&filter[status][in][]=pending&filter[price][gte]=100
```

### Available Filter Operators

| Operator | Description |
|----------|-------------|
| `eq` | Equal |
| `neq` | Not equal |
| `gt`, `gte` | Greater than (or equal) |
| `gt_or_null`, `gte_or_null` | Greater than (or equal), including null |
| `lt`, `lte` | Less than (or equal) |
| `lt_or_null`, `lte_or_null` | Less than (or equal), including null |
| `like` | Contains substring (multi-value) |
| `not_like` | Does not contain substring (multi-value) |
| `in` | Value in set (multi-value) |
| `nin` | Value not in set (multi-value) |

## Error Handling

All API errors are returned as `*models.APIError`:

```go
obj, err := c.ObjectGet(ctx, "nonexistent")
if err != nil {
    var apiErr *models.APIError
    if errors.As(err, &apiErr) {
        fmt.Println(apiErr.StatusCode) // 404
        fmt.Println(apiErr.Body)       // response body
        // Response headers are preserved on errors as well as successful responses.
        fmt.Println(apiErr.Header.Get("X-Request-ID")) // optional response header
        if apiErr.IsStatusCode(404) {
            // handle not found
        }
    }
}
```

### Retry-After on HTTP 429

`APIError.Header` retains a copy of the response headers. `RetryAfter(now)` parses
the standard `Retry-After` header as either seconds or an HTTP date:

```go
var apiErr *models.APIError
if errors.As(err, &apiErr) && apiErr.IsRateLimited() {
    now := time.Now()
    if delay, ok := apiErr.RetryAfter(now); ok {
        retryAt := now.Add(delay)
        fmt.Printf("Requeue this operation no earlier than %s\n", retryAt)
        // Schedule it in your application's retry queue; do not discard the item.
    } else {
        // Missing or invalid header: apply your application's bounded backoff policy.
    }
}
```

Past dates produce a zero delay. Missing, repeated, invalid or overflowing values
return `ok=false`. The helper also works on other API errors with this header
(for example HTTP 503). It does not sleep or retry. The SDK sends each operation
once; the application decides whether it is safe to retry and must respect its
context, cancellation and retry limits. Headers are not included in `Error()` text.

## Pagination

Most list endpoints support pagination via `ListOptions.Page` and `ListOptions.PerPage`. Users use a dedicated `UserListOptions` struct.

## Scope & Limitations

### Rate Limiting

Clients default to **200 requests per minute**, evenly spaced (one every 300 ms)
without bursts. The first request is immediate. Authentication, uploads, and
downloads share the same per-client limit. Waiting respects context cancellation
and deadlines. Concurrent requests are admitted in FIFO enqueue order, so new
arrivals cannot overtake waiting requests. Canceled waiters are removed without
consuming a request slot. Separate client instances have independent limits.

```go
// Override the limit (requests per minute).
c := client.New("https://example.seventhings.com", client.WithRateLimit(100))

// Disable throttling.
c = client.New("https://example.seventhings.com", client.WithRateLimit(0))
```

The same option works with `NewWithToken` and `NewWithCredentials`. Throttling
does not automatically retry API rate-limit errors (HTTP 429).

- **No automatic token refresh.** Call `c.Refresh(ctx, refreshToken)` manually when the access token expires.
- **No automatic retry.** Implement your own retry logic for transient failures.
- **CircularityHub uses integer IDs** while all other modules use UUID strings.

## Testing

For an executable live smoke test, see [`examples/demo`](examples/demo/README.md).
It covers catalog CRUD, timed rentals, custom thumbnails, optional reminders and
the configurable rate limit, in addition to the existing SDK modules.

```sh
go build -o bin/seventhings-demo ./examples/demo
./bin/seventhings-demo -help
```

### Unit Tests

Unit tests use `httptest.NewServer` to mock API responses — no network access required:

```sh
go test ./...
```

Run a single test:

```sh
go test ./client -run TestObjectCreate
```

### Integration Tests

Integration tests run against a live seventhings instance and are gated behind the `integration` build tag. They are skipped automatically when the required environment variables are not set.

Use a test instance: some existing tests create persistent users, uploaded files,
or field definitions, and CircularityHub order tests check out existing items.
The upgrade CRUD tests create their own catalog items, objects, tasks, and rentals
and clean them up. The thumbnail test leaves an uploaded image because the API
has no file-delete endpoint.

```sh
SEVENTHINGS_BASE_URL=https://example.seventhings.com \
SEVENTHINGS_USERNAME=user@example.com \
SEVENTHINGS_PASSWORD=password \
SEVENTHINGS_CLIENT_ID=client-id \
go test -tags=integration -v ./...
```

Run a specific integration test:

```sh
go test -tags=integration -v -run TestIntegrationLocationsCRUD ./...
```

### Integration Test Coverage

| Area | Test | What it covers |
|------|------|----------------|
| **Auth** | `TestIntegrationAuth` | Login, token presence |
| | `TestIntegrationAuthRefreshAndRevoke` | Refresh token flow, revoke tokens |
| **Ping** | `TestIntegrationPing` | Unauthenticated health check |
| **Rate limiting** | `TestIntegrationRateLimit` | Concurrent live requests spaced at the default 200/min and overridden 400/min |
| **Error headers** | `TestIntegrationAPIErrorHeaders` | Real error headers retained on 404; no deliberate live rate-limit overload |
| **Objects** | `TestIntegrationObjects` | Full CRUD cycle (create, get, list, patch, delete) |
| | `TestIntegrationObjectsListWithFilters` | `ListOptions` with `FilterEq` and `Sort` |
| | `TestIntegrationObjectsCount` | Count with and without filters |
| | `TestIntegrationObjectArchiveUnarchive` | Archive and unarchive lifecycle |
| | `TestIntegrationObjectFiles` | Upload file, attach to object, remove from object |
| **Locations** | `TestIntegrationLocationsCRUD` | Create, get, patch, list, count, delete |
| **Rooms** | `TestIntegrationRoomsCRUD` | Create (with dynamic mandatory fields), get, patch, list, count, delete |
| **Tasks** | `TestIntegrationTasks` | Full CRUD cycle |
| | `TestIntegrationTaskOptionalReminders` | Omitted, empty, and zero-interval reminders against the live API |
| | `TestIntegrationTaskUpdate` | Update title and comment via `TaskUpdate` |
| | `TestIntegrationTaskStatusTransition` | Close and re-open via `TaskUpdateStatus` |
| | `TestIntegrationTasksListWithFilters` | `TaskListOptions` with status, deadline range, reference type |
| **Rentals** | `TestIntegrationRentals` | List rental cases |
| | `TestIntegrationTimedRentalCRUD` | Timed create/update, offset timestamps, conversion to full-day, delete |
| | `TestIntegrationRentalsCRUD` | Create, get, update, list, delete (skips if instance schema doesn't support it) |
| **Catalog** | `TestIntegrationCatalogItemsCRUD` | Create, get, full-state update, filtered list, delete and verify 404 |
| **Users** | `TestIntegrationUsers` | List users |
| | `TestIntegrationUserGetAndOptions` | Get by UUID, get by ID, list with pagination and sort options |
| **Files** | `TestIntegrationFiles` | List files |
| | `TestIntegrationThumbnailCustomSize` | Upload image, verify custom dimensions and cached thumbnail |
| | `TestIntegrationFileUploadAndDownload` | Upload, get metadata, download data and thumbnail |
| **Field Definitions** | `TestIntegrationFieldDefinitions` | List field definitions |
| | `TestIntegrationFieldDefinitionsCRUD` | Create, get, update, list for both asset and room templates |
| **CircularityHub** | `TestIntegrationCircularityHub` | List items, list orders, suggest category |
| | `TestIntegrationCircularityHubItemsCRUD` | Get and update an existing item |
| | `TestIntegrationCircularityHubOrdersCRUD` | Create order from item IDs, get, update |
| | `TestIntegrationCircularityHubAddObjects` | Add objects to CircularityHub (skips on server config issues) |
| | `TestIntegrationCircularityHubSuggestRestPrice` | Suggest rest price endpoint |

Some integration tests gracefully skip when the instance doesn't support a feature or lacks required configuration. All tests clean up after themselves using `t.Cleanup`.

## License

See [LICENSE](LICENSE) for details.
