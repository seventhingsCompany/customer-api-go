package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"

	"github.com/SeventhingsCompany/customer-api-go/models"
)

// FilesList returns the most recent files. The API does not support pagination
// and returns at most 20 files.
func (c *Client) FilesList(ctx context.Context) ([]models.File, error) {
	resp, err := c.Get(ctx, "files")
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Items []models.File `json:"items"`
	}
	if err := DecodeJSON(resp, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Items, nil
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

// FileGetThumbnailWithSize downloads a thumbnail fitted within width x height,
// preserving aspect ratio. Both dimensions must be between 1 and 2000 pixels.
// The API creates and caches new sizes and may reject excess sizes with 422.
func (c *Client) FileGetThumbnailWithSize(ctx context.Context, uuid string, width, height int) ([]byte, error) {
	if width < 1 || width > 2000 || height < 1 || height > 2000 {
		return nil, fmt.Errorf("thumbnail width and height must be between 1 and 2000")
	}
	resp, err := c.GetRaw(ctx, "file/"+uuid+"/thumbnail?width="+strconv.Itoa(width)+"&height="+strconv.Itoa(height))
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
