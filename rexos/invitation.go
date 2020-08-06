package rexos

import (
	"context"

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
	query := 

	query := authURL + "invitations/sharingInvitation"

	_, ret := s.CreateHalResource(ctx, "Auth", query, invitation)
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
