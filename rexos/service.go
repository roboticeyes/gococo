// Package rexos is the connection layer to store the data in the REXos.
package rexos

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/roboticeyes/gococo/event"
	"github.com/roboticeyes/gococo/status"
	"github.com/tidwall/gjson"
)

var log = event.Log

// Service is the connection to REXos
type Service struct {
	client *Client // this is the client which is used to perform the REXos calls
}

type postFunction func(ctx context.Context, query string, payload io.Reader, contentType string) ([]byte, int, error)
type getFunction func(ctx context.Context, query string, authenticate bool) (string, []byte, int, error)
type getFunctionNoContext func(query string, authenticate bool) (string, []byte, int, error)
type patchFunction func(ctx context.Context, query string, payload io.Reader, contentType string) ([]byte, int, error)
type patchFunctionNoContext func(query string, payload io.Reader, contentType string) ([]byte, int, error)

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

// GetHalResourceWithServiceUser returns the requested resource which got fetched with the service user - x-forwarded header fields added
func (s *Service) GetHalResourceWithServiceUser(ctx context.Context, resourceName, url string) ([]byte, *status.Status) {
	return s.getHalResource(ctx, resourceName, url, s.client.GetWithServiceUser)
}

// GetHalResource returns the requested resource - x-forwarded header fields added
func (s *Service) GetHalResource(ctx context.Context, resourceName, url string) ([]byte, *status.Status) {
	return s.getHalResource(ctx, resourceName, url, s.client.Get)
}

// GetHalResourceWithServiceUserNoXF returns the requested resource which got fetched with the service user - x-forwarded header fields not added
func (s *Service) GetHalResourceWithServiceUserNoXF(ctx context.Context, resourceName, url string) ([]byte, *status.Status) {
	return s.getHalResource(ctx, resourceName, url, s.client.GetWithServiceUserNoXF)
}

// GetHalResourceNoXF returns the requested resource - x-forwarded header fields not added
func (s *Service) GetHalResourceNoXF(ctx context.Context, resourceName, url string) ([]byte, *status.Status) {
	return s.getHalResource(ctx, resourceName, url, s.client.GetNoXF)
}

// GetHalResourceWithServiceUserNoXFNoContext retruns the requested resource with service user - x-forwarded header fields not added
func (s *Service) GetHalResourceWithServiceUserNoXFNoContext(resourceName, url string) ([]byte, *status.Status) {
	return s.getHalResourceNoContext(resourceName, url, s.client.GetWithServiceUserNoXFNoContext)
}

// getHalResource performs the GET request for getting a rexos resource
// Returns the body in case of success. If an error occurred, then the according
// status is returned. The resourceName is used for the error message
func (s *Service) getHalResource(ctx context.Context, resourceName, url string, gf getFunction) ([]byte, *status.Status) {

	var body []byte
	var code int
	var err error

	_, body, code, err = gf(ctx, url, true)

	if err != nil {
		log.WithFields(event.Fields{
			"resourceName": resourceName,
			"code":         code,
			"url":          url,
		}).Debug("Can not get HAL resource: " + err.Error())
		return []byte{}, status.NewStatus(body, code, "Can not get resource "+resourceName)
	}

	// A GET request should return a value in range of [200,300[
	if code < http.StatusOK || code >= http.StatusMultipleChoices {
		log.WithFields(event.Fields{
			"resourceName": resourceName,
			"code":         code,
			"url":          url,
		}).Debug("Can not get HAL resource")
		return []byte{}, status.NewStatus(body, code, "Can not get resource "+resourceName)
	}
	return body, nil
}

// getHalResource performs the GET request for getting a rexos resource
// Returns the body in case of success. If an error occurred, then the according
// status is returned. The resourceName is used for the error message
func (s *Service) getHalResourceNoContext(resourceName, url string, gf getFunctionNoContext) ([]byte, *status.Status) {

	var body []byte
	var code int
	var err error

	_, body, code, err = gf(url, true)

	if err != nil {
		log.WithFields(event.Fields{
			"resourceName": resourceName,
			"code":         code,
			"url":          url,
		}).Debug("Can not get HAL resource: " + err.Error())
		return []byte{}, status.NewStatus(body, code, "Can not get resource "+resourceName)
	}

	// A GET request should return a value in range of [200,300[
	if code < http.StatusOK || code >= http.StatusMultipleChoices {
		log.WithFields(event.Fields{
			"resourceName": resourceName,
			"code":         code,
			"url":          url,
		}).Debug("Can not get HAL resource")
		return []byte{}, status.NewStatus(body, code, "Can not get resource "+resourceName)
	}
	return body, nil
}

// CreateHalResourceWithServiceUser uses the service user credential to create a new resource - no x-forwarded header fields added
func (s *Service) CreateHalResourceWithServiceUser(ctx context.Context, resourceName, url string, r interface{}) ([]byte, *status.Status) {
	return s.createHalResource(ctx, resourceName, url, r, s.client.PostWithServiceUser)
}

// CreateHalResource creates a new resource with the caller's credentialsa - no x-forwarded header fields added
func (s *Service) CreateHalResource(ctx context.Context, resourceName, url string, r interface{}) ([]byte, *status.Status) {
	return s.createHalResource(ctx, resourceName, url, r, s.client.Post)
}

// CreateHalResourceWithServiceUserWithXF uses the service user credential to create a new resource - x-forwarded header fields added
func (s *Service) CreateHalResourceWithServiceUserWithXF(ctx context.Context, resourceName, url string, r interface{}) ([]byte, *status.Status) {
	return s.createHalResource(ctx, resourceName, url, r, s.client.PostWithServiceUserWithXF)
}

// CreateHalResourceWithXF creates a new resource with the caller's credentials - x-forwarded header fields added
func (s *Service) CreateHalResourceWithXF(ctx context.Context, resourceName, url string, r interface{}) ([]byte, *status.Status) {
	return s.createHalResource(ctx, resourceName, url, r, s.client.PostWithXF)
}

// CreateHalResource creates a new resource which needs to be able to write itself to the proper
// JSON string. In case of success, the body is returned. The resourceName is just for error handling.
// The url is required to identify the endpoint
func (s *Service) createHalResource(ctx context.Context, resourceName, url string, r interface{}, pf postFunction) ([]byte, *status.Status) {

	var body []byte
	var code int
	var err error

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(r)

	body, code, err = pf(ctx, url, b, "application/json")
	if err != nil {
		log.WithFields(event.Fields{
			"resourceName": resourceName,
			"code":         code,
			"url":          url,
		}).Debug("Can not create HAL resource: " + err.Error())
		return []byte{}, status.NewStatus(body, code, "Can not create resource "+resourceName)
	}

	// A POST request should return a value in range of [200,300[
	if code < http.StatusOK || code >= http.StatusMultipleChoices {
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
	return s.patchHalResource(ctx, resourceName, url, r, s.client.PatchWithServiceUser)
}

// PatchHalResourceWithServiceUserNoContext patches the resource with the service user no context available
func (s *Service) PatchHalResourceWithServiceUserNoContext(resourceName, url string, r interface{}) ([]byte, *status.Status) {
	return s.patchHalResourceNoContext(resourceName, url, r, s.client.PatchWithServiceUserNoContext)
}

// PatchHalResource patches the resource with the caller's credentials
func (s *Service) PatchHalResource(ctx context.Context, resourceName, url string, r interface{}) ([]byte, *status.Status) {
	return s.patchHalResource(ctx, resourceName, url, r, s.client.Patch)
}

// PatchHalResourceWithServiceUserWithXF patches the resource with the service user
func (s *Service) PatchHalResourceWithServiceUserWithXF(ctx context.Context, resourceName, url string, r interface{}) ([]byte, *status.Status) {
	return s.patchHalResource(ctx, resourceName, url, r, s.client.PatchWithServiceUserWithXF)
}

// PatchHalResourceWithXF patches the resource with the caller's credentials
func (s *Service) PatchHalResourceWithXF(ctx context.Context, resourceName, url string, r interface{}) ([]byte, *status.Status) {
	return s.patchHalResource(ctx, resourceName, url, r, s.client.PatchWithXF)
}

// PatchHalResource sends a partial update to the requested resource
func (s *Service) patchHalResource(ctx context.Context, resourceName, url string, r interface{}, pf patchFunction) ([]byte, *status.Status) {

	var body []byte
	var code int
	var err error

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(r)

	body, code, err = pf(ctx, url, b, "application/json")
	if err != nil {
		log.WithFields(event.Fields{
			"resourceName": resourceName,
			"code":         code,
			"url":          url,
		}).Debug("Can not patch HAL resource: " + err.Error())
		return []byte{}, status.NewStatus(body, code, "Can not modify resource "+resourceName)
	}

	// A PATCH request should return a value in range of [200,300[
	if code < http.StatusOK || code >= http.StatusMultipleChoices {
		log.WithFields(event.Fields{
			"resourceName": resourceName,
			"code":         code,
			"url":          url,
		}).Debug("Can not patch HAL resource")
		return []byte{}, status.NewStatus(body, code, "Can not modify resource "+resourceName)
	}
	return body, nil
}

// patchHalResourceNoContext sends a partial update to the requested resource
func (s *Service) patchHalResourceNoContext(resourceName, url string, r interface{}, pf patchFunctionNoContext) ([]byte, *status.Status) {

	var body []byte
	var code int
	var err error

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(r)

	body, code, err = pf(url, b, "application/json")
	if err != nil {
		log.WithFields(event.Fields{
			"resourceName": resourceName,
			"code":         code,
			"url":          url,
		}).Debug("Can not patch HAL resource: " + err.Error())
		return []byte{}, status.NewStatus(body, code, "Can not modify resource "+resourceName)
	}

	// A PATCH request should return a value in range of [200,300[
	if code < http.StatusOK || code >= http.StatusMultipleChoices {
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

	// A DELETE request should return a value in range of [200,300[
	if code < http.StatusOK || code >= http.StatusMultipleChoices {
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

	u, err := url.Parse(link)
	if err != nil {
		log.Fatal(err)
		return ""
	}

	if u.Scheme == "" {
		log.Println("Scheme is empty for link:", link)
		return ""
	}

	res := strings.Split(u.Path, "/")

	if len(res) < 2 {
		return ""
	}
	return res[len(res)-1]
}

// GetNumberFromUrn extracts the number of an urn e.g. robotic-eyes:project:12345 -> 12345
func GetNumberFromUrn(urn string) (string, *status.Status) {
	parts := strings.Split(urn, ":")
	if len(parts) < 3 {
		log.WithFields(event.Fields{
			"urn": urn,
		}).Error("Failed to get number from urn")

		return "", status.NewStatus([]byte{}, http.StatusInternalServerError, "Cannot get number from urn ")
	}
	return parts[2], nil
}

// GetFileWithServiceUser returns the file from the requested url which got fetched with the service user
func (s *Service) GetFileWithServiceUser(ctx context.Context, c *gin.Context, url string) *status.Status {
	code, err := s.client.GetFileWithServiceUser(ctx, c, url, true)

	if err != nil {
		log.WithFields(event.Fields{
			"url": url,
		}).Debug("Can not get file: " + err.Error())
		return status.NewStatus([]byte{}, code, "Can not get file from "+url)
	}
	if code != http.StatusOK {
		log.WithFields(event.Fields{
			"url": url,
		}).Debug("Can not get file ")
		return status.NewStatus([]byte{}, code, "Can not get file from "+url)
	}

	return nil
}
