// Package rexos is the connection layer to store the data in the REXos.
package rexos

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/roboticeyes/gococo/event"
	"github.com/roboticeyes/gococo/status"
	"github.com/tidwall/gjson"
)

var log = event.Log

// Service is the connection to REXos
type Service struct {
	client *Client // this is the client which is used to perform the REXos calls
}

// NewService returns a new rexos service which is implementing the RexOSAccessor interface
func NewService(config Config) *Service {
	return &Service{
		client: NewClient(config),
	}
}

// StripTemplateParameter removes the trailing template parameters of an HATEOAS URL
// For example: "https://rex.robotic-eyes.com/rex-gateway/api/v2/rexReferences/1000/project{?projection}"
func StripTemplateParameter(templateURL string) string {
	return strings.Split(templateURL, "{")[0]
}

// GetSelfLinkFromHal returns the stripped self link of a HAL resource. The input is the JSON
// response as string
func GetSelfLinkFromHal(json []byte) string {
	return StripTemplateParameter(gjson.Get(string(json), "_links.self.href").String())
}

// GetUrnFromHal returns the urn of a resource which was returned as a HAL response
func GetUrnFromHal(json []byte) string {
	return StripTemplateParameter(gjson.Get(string(json), "urn").String())
}

// GetPublicShareLinkFromHal returns the stripped public share link of a HAL resource. The input is the JSON
// response as string
func GetPublicShareLinkFromHal(json []byte) string {
	return StripTemplateParameter(gjson.Get(string(json), "_links.publicShare.href").String())
}

// GetProjectLinkFromHal returns the stripped project link of a HAL resource. The input is the JSON
// response as string
func GetProjectLinkFromHal(json []byte) string {
	return StripTemplateParameter(gjson.Get(string(json), "_links.project.href").String())
}

// GetHalResourceWithServiceUser returns the requested resource which got fetched with the service user
func (s *Service) GetHalResourceWithServiceUser(ctx context.Context, resourceName, url string) ([]byte, *status.Status) {
	return s.getHalResource(ctx, resourceName, url, true)
}

// GetHalResource returns the requested resource
func (s *Service) GetHalResource(ctx context.Context, resourceName, url string) ([]byte, *status.Status) {
	return s.getHalResource(ctx, resourceName, url, false)
}

// getHalResource performs the GET request for getting a rexos resource
// Returns the body in case of success. If an error occurred, then the according
// status is returned. The resourceName is used for the error message
func (s *Service) getHalResource(ctx context.Context, resourceName, url string, useServiceUser bool) ([]byte, *status.Status) {

	var body []byte
	var code int
	var err error

	if useServiceUser {
		_, body, code, err = s.client.GetWithServiceUser(ctx, url, true)
	} else {
		_, body, code, err = s.client.Get(ctx, url, true)
	}

	if err != nil {
		log.WithFields(event.Fields{
			"resourceName": resourceName,
			"code":         code,
			"url":          url,
		}).Debug("Can not get HAL resource: " + err.Error())
		return []byte{}, status.NewStatus(body, code, "Can not get resource "+resourceName)
	}

	// A GET request should return 200 StatusOK
	if code != http.StatusOK {
		log.WithFields(event.Fields{
			"resourceName": resourceName,
			"code":         code,
			"url":          url,
		}).Debug("Can not get HAL resource")
		return []byte{}, status.NewStatus(body, code, "Can not get resource "+resourceName)
	}
	return body, nil
}

// CreateHalResourceWithServiceUser uses the service user credential to create a new resource
func (s *Service) CreateHalResourceWithServiceUser(ctx context.Context, resourceName, url string, r interface{}) ([]byte, *status.Status) {
	return s.createHalResource(ctx, resourceName, url, r, true)
}

// CreateHalResource creates a new resource with the caller's credentials
func (s *Service) CreateHalResource(ctx context.Context, resourceName, url string, r interface{}) ([]byte, *status.Status) {
	return s.createHalResource(ctx, resourceName, url, r, false)
}

// CreateHalResource creates a new resource which needs to be able to write itself to the proper
// JSON string. In case of success, the body is returned. The resourceName is just for error handling.
// The url is required to identify the endpoint
func (s *Service) createHalResource(ctx context.Context, resourceName, url string, r interface{}, useServiceUser bool) ([]byte, *status.Status) {

	var body []byte
	var code int
	var err error

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(r)

	if useServiceUser {
		body, code, err = s.client.PostWithServiceUser(ctx, url, b, "application/json")
	} else {
		body, code, err = s.client.Post(ctx, url, b, "application/json")
	}
	if err != nil {
		log.WithFields(event.Fields{
			"resourceName": resourceName,
			"code":         code,
			"url":          url,
		}).Debug("Can not create HAL resource: " + err.Error())
		return []byte{}, status.NewStatus(body, code, "Can not create resource "+resourceName)
	}

	// A POST request should return 201 StatusCreated
	if code != http.StatusCreated {
		log.WithFields(event.Fields{
			"resourceName": resourceName,
			"code":         code,
			"url":          url,
		}).Debug("Can not create HAL resource")
		return []byte{}, status.NewStatus(body, code, "Can not create resource "+resourceName)
	}
	return body, nil
}

// PatchHalResourceWithServiceUser patches the resource with the service user
func (s *Service) PatchHalResourceWithServiceUser(ctx context.Context, resourceName, url string, r interface{}) ([]byte, *status.Status) {
	return s.patchHalResource(ctx, resourceName, url, r, true)
}

// PatchHalResource patches the resource with the caller's credentials
func (s *Service) PatchHalResource(ctx context.Context, resourceName, url string, r interface{}) ([]byte, *status.Status) {
	return s.patchHalResource(ctx, resourceName, url, r, false)
}

// PatchHalResource sends a partial update to the requested resource
func (s *Service) patchHalResource(ctx context.Context, resourceName, url string, r interface{}, useServiceUser bool) ([]byte, *status.Status) {

	var body []byte
	var code int
	var err error

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(r)

	if useServiceUser {
		body, code, err = s.client.PatchWithServiceUser(ctx, url, b, "application/json")
	} else {
		body, code, err = s.client.Patch(ctx, url, b, "application/json")
	}
	if err != nil {
		log.WithFields(event.Fields{
			"resourceName": resourceName,
			"code":         code,
			"url":          url,
		}).Debug("Can not patch HAL resource: " + err.Error())
		return []byte{}, status.NewStatus(body, code, "Can not modify resource "+resourceName)
	}

	// A PATCH request should return 200 StatusOK
	if code != http.StatusOK {
		log.WithFields(event.Fields{
			"resourceName": resourceName,
			"code":         code,
			"url":          url,
		}).Debug("Can not patch HAL resource")
		return []byte{}, status.NewStatus(body, code, "Can not modify resource "+resourceName)
	}
	return body, nil
}

// DeleteHalResource patches the resource with the caller's credentials
func (s *Service) DeleteHalResource(ctx context.Context, resourceName, url string) *status.Status {
	_, code, err := s.client.Delete(ctx, url)
	if err != nil {
		log.WithFields(event.Fields{
			"resoucreName": resourceName,
			"code":         code,
			"url":          url,
		}).Info("Failed to delete HAL resource" + err.Error())
		return status.NewStatus(nil, code, "Can not delete resource "+resourceName)
	}

	// A DELETE request should return 200 StatusOK or 204 StatusNoContent
	if code != http.StatusOK && code != http.StatusNoContent {
		log.WithFields(event.Fields{
			"resourceName": resourceName,
			"code":         code,
			"url":          url,
		}).Debug("Can not delete HAL resource")
		return status.NewStatus(nil, code, "Can not delete resource "+resourceName)
	}
	return nil
}

// DownloadFileContent uploads the actual binary file for a project file
func (s *Service) DownloadFileContent(ctx context.Context, downloadURL string, authenticate bool) ([]byte, *status.Status) {
	// download file content
	fileName, blob, code, err := s.client.Get(ctx, downloadURL, authenticate)
	if fileName == "" {
		fileName = "file.rex"
	}
	if err != nil {
		log.WithFields(event.Fields{
			"downloadUrl": downloadURL,
			"fileName":    fileName,
		}).Debug("Can not download file content: " + err.Error())
		return []byte{}, status.NewStatus(nil, code, "Can not access file "+fileName)
	}

	if code != http.StatusOK {
		log.WithFields(event.Fields{
			"downloadUrl": downloadURL,
			"fileName":    fileName,
		}).Debug("Can not download file content")
		return []byte{}, status.NewStatus(nil, code, "Can not access file "+fileName)
	}

	return blob, nil
}

// UploadFileContent uploads the actual binary file for a project file
func (s *Service) UploadFileContent(ctx context.Context, uploadURL string, downloadURL string, authenticate bool) *status.Status {

	// download file content
	fileName, blob, code, err := s.client.Get(ctx, downloadURL, authenticate)
	if fileName == "" {
		fileName = "file.rex"
	}
	if err != nil {
		log.WithFields(event.Fields{
			"downloadUrl": downloadURL,
			"uploadUrl":   uploadURL,
			"fileName":    fileName,
		}).Debug("Can not download file content: " + err.Error())
		return status.NewStatus(nil, code, "Can not access file "+fileName)
	}

	if code != http.StatusOK {
		log.WithFields(event.Fields{
			"downloadUrl": downloadURL,
			"uploadUrl":   uploadURL,
			"fileName":    fileName,
		}).Debug("Can not download file content")
		return status.NewStatus(nil, code, "Can not access file "+fileName)
	}

	// get filename from header information

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", fileName)

	io.Copy(part, bytes.NewReader(blob))
	writer.Close()

	responseBody, code, err := s.client.Post(ctx, uploadURL, body, writer.FormDataContentType())
	if err != nil {
		log.WithFields(event.Fields{
			"downloadUrl": downloadURL,
			"uploadUrl":   uploadURL,
			"fileName":    fileName,
		}).Debug("Can not upload file content: " + err.Error())
		return status.NewStatus(responseBody, code, "Can not upload file "+fileName)
	}
	if code != http.StatusOK {
		log.WithFields(event.Fields{
			"downloadUrl": downloadURL,
			"uploadUrl":   uploadURL,
			"fileName":    fileName,
		}).Debug("Can not upload file content: " + err.Error())
		return status.NewStatus(responseBody, code, "Can not upload file "+fileName)
	}
	return nil
}

// UploadMultipartFile uploads the content of a multipart file
func (s *Service) UploadMultipartFile(ctx context.Context, fileName string, uploadURL string, data io.Reader) *status.Status {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", fileName)

	io.Copy(part, data)
	writer.Close()

	responseBody, code, err := s.client.Post(ctx, uploadURL, body, writer.FormDataContentType())

	// responseBody, code, err := s.client.Post(ctx, uploadURL, data, contentType)
	if err != nil {
		log.WithFields(event.Fields{
			"uploadUrl": uploadURL,
			"fileName":  fileName,
		}).Debug("Can not upload file content: " + err.Error())
		return status.NewStatus(responseBody, code, "Can not upload file "+fileName)
	}
	if code != http.StatusOK {
		log.WithFields(event.Fields{
			"uploadUrl": uploadURL,
			"fileName":  fileName,
		}).Debug("Can not upload file content: " + err.Error())
		return status.NewStatus(responseBody, code, "Can not upload file "+fileName)
	}
	return nil
}

// UploadFile uploads the byte array as file
func (s *Service) UploadFile(ctx context.Context, fileName string, uploadURL string, data []byte) *status.Status {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", fileName)

	io.Copy(part, bytes.NewReader(data))
	writer.Close()

	responseBody, code, err := s.client.Post(ctx, uploadURL, body, writer.FormDataContentType())
	if err != nil {
		log.WithFields(event.Fields{
			"uploadUrl": uploadURL,
			"fileName":  fileName,
		}).Debug("Can not upload file content: " + err.Error())
		return status.NewStatus(responseBody, code, "Can not upload file "+fileName)
	}
	if code != http.StatusOK {
		log.WithFields(event.Fields{
			"uploadUrl": uploadURL,
			"fileName":  fileName,
		}).Debug("Can not upload file content: " + err.Error())
		return status.NewStatus(responseBody, code, "Can not upload file "+fileName)
	}
	return nil
}

// GetHashFromDownloadLink extracts the contentHash from the rex project file download link
// e.g. https://api-dev-01.rexos.cloud/rex-gateway/api/v2/projectFiles/1747/file?contentHash=2dd1aee5e71621ea56042c92886be464
func GetHashFromDownloadLink(link string) string {

	res := strings.Split(link, "=")

	if len(res) < 2 {
		return ""
	}

	return res[1]
}

// GetGUIDFromRexTagURL extracts the GUID of the portal reference based on the actual
// RexTag link
func GetGUIDFromRexTagURL(link string) string {

	// check if link starts with http
	if !strings.HasPrefix(link, "http") {
		return ""
	}

	res := strings.Split(link, "/")

	if len(res) < 2 {
		return ""
	}
	return res[len(res)-1]
}
