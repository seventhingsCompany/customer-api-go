package models

// RentalCaseStatus represents the status of a rental case.
type RentalCaseStatus string

const (
	// RentalCaseStatusRequested indicates the rental has been requested.
	RentalCaseStatusRequested RentalCaseStatus = "requested"
	// RentalCaseStatusConfirmed indicates the rental has been confirmed.
	RentalCaseStatusConfirmed RentalCaseStatus = "confirmed"
	// RentalCaseStatusBorrowed indicates the item is currently borrowed.
	RentalCaseStatusBorrowed RentalCaseStatus = "borrowed"
	// RentalCaseStatusRejected indicates the rental was rejected.
	RentalCaseStatusRejected RentalCaseStatus = "rejected"
	// RentalCaseStatusCompleted indicates the rental is completed.
	RentalCaseStatusCompleted RentalCaseStatus = "completed"
	// RentalCaseStatusReturnOverdue indicates the return is overdue.
	RentalCaseStatusReturnOverdue RentalCaseStatus = "return_overdue"
	// RentalCaseStatusPickupOverdue indicates the pickup is overdue.
	RentalCaseStatusPickupOverdue RentalCaseStatus = "pickup_overdue"
)

// RentalCaseReferenceType is the type of entity a rental case references.
type RentalCaseReferenceType string

const (
	// RentalCaseReferenceTypeAsset references an asset.
	RentalCaseReferenceTypeAsset RentalCaseReferenceType = "asset"
)

// RenterType specifies how the renter is identified.
type RenterType string

const (
	// RenterTypePlain is a plain-text renter name.
	RenterTypePlain RenterType = "plain"
	// RenterTypeUser is a renter identified by user reference.
	RenterTypeUser RenterType = "user"
)

// RentalCaseReference is a reference to an entity in a rental case response.
type RentalCaseReference struct {
	Type RentalCaseReferenceType `json:"type"`
	UUID string                  `json:"uuid"`
	Name string                  `json:"name"`
	ID   int                     `json:"id"`
}

// RentalCaseReferenceInput is a reference to an entity in a rental case request.
type RentalCaseReferenceInput struct {
	Type RentalCaseReferenceType `json:"type"`
	UUID string                  `json:"uuid"`
}

// RentalCaseRenter specifies the renter in a rental case request.
type RentalCaseRenter struct {
	Type  RenterType `json:"type"`
	Value string     `json:"value"`
}

// RentalCaseRenterResponse represents the renter in a rental case response.
// The API returns renter as an object with type and value (same shape as input),
// or null when no renter is set.
type RentalCaseRenterResponse struct {
	Type  RenterType `json:"type"`
	Value string     `json:"value"`
}

// RentalCase represents a rental case in the seventhings API.
type RentalCase struct {
	UUID                string                    `json:"uuid"`
	Status              RentalCaseStatus          `json:"status"`
	Title               string                    `json:"title"`
	Renter              *RentalCaseRenterResponse `json:"renter"`
	References          []RentalCaseReference     `json:"references"`
	IsFullDay           bool                      `json:"is_full_day"`
	IssueAt             *string                   `json:"issue_at"`
	DueAt               *string                   `json:"due_at"`
	IssueDate           *string                   `json:"issue_date"`
	IssueDateReminder   *TimeInterval             `json:"issue_date_reminder"`
	DueDate             *string                   `json:"due_date"`
	DueDateReminder     *TimeInterval             `json:"due_date_reminder"`
	Comment             *string                   `json:"comment"`
	ResponsibleUserUUID *string                   `json:"responsible_user_uuid"`
	Author              *string                   `json:"author"`
	Attachments         []AttachmentFile          `json:"attachments"`
	CreatedAt           string                    `json:"created_at"`
	UpdatedAt           string                    `json:"updated_at"`
}

// CreateRentalCase is the request body for creating a rental case.
// SDK create calls require a non-nil Renter to avoid a server-side failure.
type CreateRentalCase struct {
	// Omit IsFullDay to let the API infer the kind from IssueAt and DueAt.
	IsFullDay           *bool                      `json:"is_full_day,omitempty"`
	IssueAt             *string                    `json:"issue_at,omitempty"`
	DueAt               *string                    `json:"due_at,omitempty"`
	Title               string                     `json:"title"`
	Renter              *RentalCaseRenter          `json:"renter,omitempty"`
	References          []RentalCaseReferenceInput `json:"references"`
	IssueDate           string                     `json:"issue_date"`
	IssueDateReminder   *TimeInterval              `json:"issue_date_reminder,omitempty"`
	DueDate             string                     `json:"due_date"`
	DueDateReminder     *TimeInterval              `json:"due_date_reminder,omitempty"`
	Comment             string                     `json:"comment"`
	ResponsibleUserUUID string                     `json:"responsible_user_uuid"`
	Attachments         []string                   `json:"attachments"`
}

// UpdateRentalCase is the complete replacement body for updating a rental case.
// SDK update calls require a non-nil Renter to avoid a server-side failure.
type UpdateRentalCase struct {
	// Omit the kind and times to retain the stored kind and move times to the dates.
	IsFullDay           *bool                      `json:"is_full_day,omitempty"`
	IssueAt             *string                    `json:"issue_at,omitempty"`
	DueAt               *string                    `json:"due_at,omitempty"`
	Title               string                     `json:"title"`
	Renter              *RentalCaseRenter          `json:"renter,omitempty"`
	References          []RentalCaseReferenceInput `json:"references"`
	IssueDate           string                     `json:"issue_date"`
	IssueDateReminder   *TimeInterval              `json:"issue_date_reminder,omitempty"`
	DueDate             string                     `json:"due_date"`
	DueDateReminder     *TimeInterval              `json:"due_date_reminder,omitempty"`
	Comment             string                     `json:"comment"`
	ResponsibleUserUUID string                     `json:"responsible_user_uuid"`
	Attachments         []string                   `json:"attachments"`
}
