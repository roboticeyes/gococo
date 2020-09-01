package rexos

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/roboticeyes/gococo/event"
	"github.com/roboticeyes/gococo/status"
	"github.com/tidwall/gjson"
)

// UserData is a container for email and name of a new user
type UserData struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// UserAndProjectData describes user and project information
type UserAndProjectData struct {
	UserData
	ProjectName string `json:"projectName"`
	ProjectURL  string `json:"projectUrl"`
}

// ProjectInvitation container for the user to invite and the sharing information
type ProjectInvitation struct {
	User  UserData `json:"user"`
	Write bool     `json:"write"`
	Read  bool     `json:"read"`
}

// CreateProjectInvitation shares a project with a new user
func (s *Service) CreateProjectInvitation(ctx context.Context, projectUrn string, projectInvitation ProjectInvitation, projectResourceURL, userResourceURL, invitationURL, rexCodesResourceURL string) (ProjectInvitation, *status.Status) {
	// find project
	query := QueryFindByUrn(projectResourceURL, projectUrn)
	projectResult, ret := s.GetHalResource(ctx, "Project", query)
	if ret != nil {
		log.WithFields(event.Fields{
			"projectUrn": projectUrn,
			"query":      query,
			"status":     ret,
		}).Error("Failed to get project")

		ret.Message = "Could not get project. Please make sure you have the correct access rights."
		return ProjectInvitation{}, ret
	}
	var project Project
	json.Unmarshal(projectResult, &project)

	// find key of portal reference
	key := gjson.Get(string(projectResult), "_embedded.rexReferences.#(type==\"portal\").key")

	query = invitationURL
	var invitation UserAndProjectData
	invitation.Email = projectInvitation.User.Email
	invitation.FirstName = projectInvitation.User.FirstName
	invitation.LastName = projectInvitation.User.LastName
	invitation.ProjectName = project.Name
	invitation.ProjectURL = rexCodesResourceURL + "/" + key.String()

	invResult, ret := s.CreateHalResourceWithXF(ctx, "Auth", query, invitation)
	if ret != nil {
		log.WithFields(event.Fields{
			"status": ret,
			"query":  query,
		}).Error("Failed to create invitation")

		ret.Message = "Could not create invitation. Please make sure you have the correct access rights."
		return ProjectInvitation{}, ret
	}
	// find userId
	userID := gjson.Get(string(invResult), "userId").String()

	// share project with user
	action := writeAction
	if projectInvitation.Read {
		action = readAction
	}

	projectNumber, ret := GetNumberFromUrn(projectUrn)
	if ret != nil {
		return projectInvitation, ret
	}

	share := UserShareReduced{UserID: userID, Action: action}

	// update user sharing
	query = projectResourceURL + "/" + projectNumber + "/userShares"
	_, ret = s.CreateHalResource(ctx, "Projects", query, share)
	if ret != nil {
		log.WithFields(event.Fields{
			"status":     ret,
			"projectUrn": projectUrn,
			"query":      query,
		}).Error("Failed to create user share information")

		ret.Message = "Cannot not create user share information for the project. Please make sure you have the correct access rights."
		return projectInvitation, ret
	}

	return projectInvitation, nil
}

// PrettyJson for development use
func PrettyJson(data interface{}) (string, error) {
	const (
		empty = ""
		tab   = "\t"
	)

	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent(empty, tab)

	err := encoder.Encode(data)
	if err != nil {
		return empty, err
	}
	return buffer.String(), nil
}
