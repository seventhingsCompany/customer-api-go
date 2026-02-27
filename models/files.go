package models

// File represents a file resource in the seventhings API.
type File struct {
	UUID         string `json:"uuid"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Size         int    `json:"size"`
	CreatorID    int    `json:"creator_id"`
	CreatedAt    string `json:"created_at"`
	DataURI      string `json:"data_uri"`
	ThumbnailURI string `json:"thumbnail_uri"`
}

// FileAttachment associates a file with a field on a resource.
type FileAttachment struct {
	FieldKey string `json:"field-key"`
	FileUUID string `json:"file-uuid"`
}

// CountResponse is the response body for count endpoints.
type CountResponse struct {
	Count int `json:"count"`
}
