package rexos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

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

// ProjectInvitation is a container for user project and sharing information
type ProjectInvitation struct {
	InviteUser UserAndProjectData `json:"inviteUser"`
	// ProjectAPIUrl string     `json:"projectAPIUrl"`
	Sharing string `json:"sharing"`
}

// CreateProjectInvitation shares a project with a new user
func (s *Service) CreateProjectInvitation(ctx context.Context, projectUrn string, invitation ProjectInvitation, projectResourceURL, invitationURL, rexCodesResourceURL string) (ProjectInvitation, *status.Status) {
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
	fmt.Println("ProjectResult")
	fmt.Println(string(projectResult))

	// find key of portal reference
	key := gjson.Get(string(projectResult), "_embedded.rexReferences.#(type==\"portal\").key")

	query = invitationURL
	// var invitation ProjectInvitation
	// var inviteUser UserAndProjectData
	// inviteUser.Email = inv.InviteUser.Email
	// inviteUser.FirstName = inv.InviteUser.FirstName
	// inviteUser.LastName = inv.InviteUser.LastName
	// invitation.InviteUser = inviteUser
	invitation.InviteUser.ProjectName = project.Name
	invitation.InviteUser.ProjectURL = rexCodesResourceURL + "/" + key.String()
	// invitation.ProjectAPIUrl = invitation.InviteUser.ProjectUrl
	// invitation.Sharing = inv.Sharing

	fmt.Println(query)
	body, _ := PrettyJson(invitation)
	fmt.Println(body)
	_, ret = s.CreateHalResource(ctx, "Auth", query, invitation)
	if ret != nil {
		log.WithFields(event.Fields{
			"status": ret,
			"query":  query,
		}).Error("Failed to create invitation")

		ret.Message = "Could not create invitation. Please make sure you have the correct access rights."
		return ProjectInvitation{}, ret
	}
	return invitation, nil
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
