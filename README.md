# seventhings Go SDK

Go client library for the [seventhings](https://seventhings.com) customer API.

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

// List all files
files, err := c.FilesList(ctx)
```

### Tasks

```go
uuid, err := c.TaskCreate(ctx, models.CreateTask{Title: "Inspect server room"})
task, err := c.TaskGet(ctx, uuid)
tasks, err := c.TasksList(ctx, &models.TaskListOptions{Status: &open})
err = c.TaskUpdate(ctx, uuid, models.UpdateTask{Title: "Updated title"})
err = c.TaskUpdateStatus(ctx, uuid, models.TaskStatusClosed)
err = c.TaskDelete(ctx, uuid)
```

### Rental Cases

```go
uuid, err := c.RentalCaseCreate(ctx, models.CreateRentalCase{
    Renter: &models.RentalCaseRenter{Type: models.RenterTypePlain, Value: "John"},
})
rc, err := c.RentalCaseGet(ctx, uuid)
cases, err := c.RentalCasesList(ctx, nil)
err = c.RentalCaseUpdate(ctx, uuid, models.UpdateRentalCase{Comment: &comment})
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
        if apiErr.IsStatusCode(404) {
            // handle not found
        }
    }
}
```

## Pagination

Most list endpoints support pagination via `ListOptions.Page` and `ListOptions.PerPage`. Users use a dedicated `UserListOptions` struct.

## Scope & Limitations

- **No automatic token refresh.** Call `c.Refresh(ctx, refreshToken)` manually when the access token expires.
- **No automatic retry.** Implement your own retry logic for transient failures.
- **No rate limiting.** The client does not throttle requests.
- **CircularityHub uses integer IDs** while all other modules use UUID strings.

## Testing

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
| **Objects** | `TestIntegrationObjects` | Full CRUD cycle (create, get, list, patch, delete) |
| | `TestIntegrationObjectsListWithFilters` | `ListOptions` with `FilterEq` and `Sort` |
| | `TestIntegrationObjectsCount` | Count with and without filters |
| | `TestIntegrationObjectArchiveUnarchive` | Archive and unarchive lifecycle |
| | `TestIntegrationObjectFiles` | Upload file, attach to object, remove from object |
| **Locations** | `TestIntegrationLocationsCRUD` | Create, get, patch, list, count, delete |
| **Rooms** | `TestIntegrationRoomsCRUD` | Create (with dynamic mandatory fields), get, patch, list, count, delete |
| **Tasks** | `TestIntegrationTasks` | Full CRUD cycle |
| | `TestIntegrationTaskUpdate` | Update title and comment via `TaskUpdate` |
| | `TestIntegrationTaskStatusTransition` | Close and re-open via `TaskUpdateStatus` |
| | `TestIntegrationTasksListWithFilters` | `TaskListOptions` with status, deadline range, reference type |
| **Rentals** | `TestIntegrationRentals` | List rental cases |
| | `TestIntegrationRentalsCRUD` | Create, get, update, list, delete (skips if instance schema doesn't support it) |
| **Users** | `TestIntegrationUsers` | List users |
| | `TestIntegrationUserGetAndOptions` | Get by UUID, get by ID, list with pagination and sort options |
| **Files** | `TestIntegrationFiles` | List files |
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
