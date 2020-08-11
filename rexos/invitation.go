package rexos

import (
	"context"
	"encoding/json"

	"github.com/roboticeyes/gococo/event"
	"github.com/roboticeyes/gococo/status"
)

// User is a container for email and name of a new user
type User struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// InviteUser describes user and project information
type InviteUser struct {
	User
	ProjectName string `json:"projectName"`
	ProjectUrl  string `json:"projectUrl"`
}

// ProjectInvitation is a container for user project and sharing information
type ProjectInvitation struct {
	InviteUser    InviteUser `json:"inviteUser"`
	ProjectAPIUrl string     `json:"projectAPIUrl"`
	Sharing       string     `json:"sharing"`
}

// CreateProjectInvitation shares a project with a new user
func (s *Service) CreateProjectInvitation(ctx context.Context, projectUrn string, user User, projectResourceURL, authResourceURL string) (ProjectInvitation, *status.Status) {
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

	query = authURL + "invitations/sharingInvitation"
	var invitation ProjectInvitation
	var inviteUser InviteUser
	inviteUser.Email = user.Email
	inviteUser.FirstName = user.FirstName
	inviteUser.LastName = user.LastName
	invitation.InviteUser = inviteUser

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
