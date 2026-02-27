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
// Input is an object with type and value; output is a plain string.
type RentalCaseRenter struct {
	Type  RenterType `json:"type"`
	Value string     `json:"value"`
}

// RentalCase represents a rental case in the seventhings API.
type RentalCase struct {
	UUID              string                `json:"uuid"`
	Status            RentalCaseStatus      `json:"status"`
	Renter            *string               `json:"renter"`
	References        []RentalCaseReference `json:"references"`
	PickupDate        *string               `json:"pickup_date"`
	ReturnDate        *string               `json:"return_date"`
	Comment           *string               `json:"comment"`
	RecurringSchedule *TimeInterval         `json:"recurring_schedule"`
	Attachments       []AttachmentFile      `json:"attachments"`
	CreatedAt         string                `json:"created_at"`
	UpdatedAt         string                `json:"updated_at"`
}

// CreateRentalCase is the request body for creating a rental case.
type CreateRentalCase struct {
	Renter            *RentalCaseRenter          `json:"renter,omitempty"`
	References        []RentalCaseReferenceInput `json:"references,omitempty"`
	PickupDate        *string                    `json:"pickup_date,omitempty"`
	ReturnDate        *string                    `json:"return_date,omitempty"`
	Comment           *string                    `json:"comment,omitempty"`
	RecurringSchedule *TimeInterval              `json:"recurring_schedule,omitempty"`
	Attachments       []string                   `json:"attachments,omitempty"`
}

// UpdateRentalCase is the request body for updating a rental case.
type UpdateRentalCase struct {
	Renter            *RentalCaseRenter          `json:"renter,omitempty"`
	References        []RentalCaseReferenceInput `json:"references,omitempty"`
	PickupDate        *string                    `json:"pickup_date,omitempty"`
	ReturnDate        *string                    `json:"return_date,omitempty"`
	Comment           *string                    `json:"comment,omitempty"`
	RecurringSchedule *TimeInterval              `json:"recurring_schedule,omitempty"`
	Attachments       []string                   `json:"attachments,omitempty"`
}
