package client

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// FilesList returns all files.
func (c *Client) FilesList(ctx context.Context) ([]models.File, error) {
	resp, err := c.Get(ctx, "files")
	if err != nil {
		return nil, err
	}
	var result []models.File
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// FileGet returns a single file's metadata by UUID.
func (c *Client) FileGet(ctx context.Context, uuid string) (*models.File, error) {
	resp, err := c.Get(ctx, "file/"+uuid)
	if err != nil {
		return nil, err
	}
	var result models.File
	if err := DecodeJSON(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FileUpload uploads a file and returns its UUID. The file is sent as
// multipart/form-data with the field name "data".
func (c *Client) FileUpload(ctx context.Context, filename string, data io.Reader) (string, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	part, err := mw.CreateFormFile("data", filename)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, data); err != nil {
		return "", err
	}
	if err := mw.Close(); err != nil {
		return "", err
	}

	resp, err := c.PostMultipart(ctx, "file", &buf, mw.FormDataContentType())
	if err != nil {
		return "", err
	}
	// The file endpoint returns a Location-UUID header with the file UUID.
	// The Location header points to the /data sub-resource, so we prefer
	// Location-UUID when available.
	if uuid := resp.Header.Get("Location-UUID"); uuid != "" {
		return uuid, nil
	}
	return UUIDFromLocationHeader(resp)
}

// FileGetData downloads the binary data of a file by UUID.
func (c *Client) FileGetData(ctx context.Context, uuid string) ([]byte, error) {
	resp, err := c.GetRaw(ctx, "file/"+uuid+"/data")
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// FileGetThumbnail downloads the thumbnail image of a file by UUID.
func (c *Client) FileGetThumbnail(ctx context.Context, uuid string) ([]byte, error) {
	resp, err := c.GetRaw(ctx, "file/"+uuid+"/thumbnail")
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
