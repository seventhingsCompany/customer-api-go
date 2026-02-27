package models

// FilterObject represents a filter/sort payload for CircularityHub POST-body
// filter endpoints such as suggest-category.
type FilterObject struct {
	Filter map[string]map[FilterOperator]any `json:"filter,omitempty"`
	Sort   map[string]SortDirection          `json:"sort,omitempty"`
}

// AddObjectEntry describes an object to be added to the CircularityHub.
type AddObjectEntry struct {
	Category string `json:"category"`
	Price    string `json:"price"`
}

// CircularityHubBillingData holds billing address details for a CircularityHub order.
type CircularityHubBillingData struct {
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	Street      *string `json:"street"`
	HouseNumber *string `json:"house_number"`
	ZipCode     *string `json:"zip_code"`
	City        *string `json:"city"`
}

// CircularityHubOrder represents an order in the CircularityHub module.
type CircularityHubOrder struct {
	ID                 int                        `json:"id"`
	OrderNumber        string                     `json:"order_number"`
	CreatedAt          string                     `json:"created_at"`
	UserID             *int                       `json:"user_id"`
	TotalPrice         *float64                   `json:"total_price"`
	Completed          bool                       `json:"completed"`
	Cancelled          bool                       `json:"cancelled"`
	CancellationReason *string                    `json:"cancellation_reason"`
	BillingData        *CircularityHubBillingData `json:"billing_data"`
	Articles           []map[string]any           `json:"articles"`
}
