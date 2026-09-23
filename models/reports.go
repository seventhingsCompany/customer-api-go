package models

// ReportTemplate is a PDF template available on the instance.
type ReportTemplate struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// CreateReport selects a template and at least one object to render, in order.
type CreateReport struct {
	ReportTemplateUUID string   `json:"report_template_uuid"`
	ObjectUUIDs        []string `json:"object_uuids"`
}
