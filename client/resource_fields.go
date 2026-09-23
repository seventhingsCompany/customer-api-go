package client

// unwrapResourceFields keeps the flat field-map contract for room/location
// responses that use a {uuid, fields} envelope. Legacy flat maps pass through.
func unwrapResourceFields(resource map[string]any) map[string]any {
	fields, wrapped := resource["fields"].(map[string]any)
	uuid, hasUUID := resource["uuid"].(string)
	if !wrapped || !hasUUID {
		return resource
	}
	// Keep the envelope identity available without overwriting custom fields.
	if _, exists := fields["uuid"]; !exists {
		fields["uuid"] = uuid
	}
	return fields
}
