package rexos

import (
	"context"
	"encoding/json"

	"github.com/roboticeyes/gococo/event"
	"github.com/roboticeyes/gococo/status"
	"github.com/tidwall/gjson"
)

// Project is the top level description for a new project
type Project struct {

	// Name of the project
	Name string `json:"name" example:"Machine 1"`

	// Owner of the project
	Owner string `json:"owner" example:"test-user"`

	// Urn will be generated by the rexos system in order to identify this resource [out]
	Urn string `json:"urn" example:"robotic-eyes:project:5191 [out]"`
}

// TransferProject updates the owner of a project.
func (s *Service) TransferProject(ctx context.Context, projectResourceURL, userResourceURL string, projectUrn string, newOwner string) (Project, *status.Status) {
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
		return Project{}, ret
	}
	var project Project
	json.Unmarshal(projectResult, &project)

	// find user for given email or userID
	query = userResourceURL + "/search/findUserIdByEmail?email=" + newOwner
	userIDResult, ret := s.GetHalResource(ctx, "User", query)
	if ret != nil {
		if ret.Code == 404 {
			// find user for given userID
			query = userResourceURL + "/search/findUserIdByUsername?username=" + newOwner
			userIDResult, ret = s.GetHalResource(ctx, "User", query)
			if ret != nil {
				log.WithFields(event.Fields{
					"projectUrn": projectUrn,
					"query":      query,
					"status":     ret,
				}).Error("Failed to get user by username or email")

				ret.Message = "Could not get user by username or email. User does not exist."
				return project, ret
			}
		} else {
			log.WithFields(event.Fields{
				"projectUrn": projectUrn,
				"query":      query,
				"status":     ret,
			}).Error("Failed to get user by email")

			ret.Message = "Could not get user by email."
			return project, ret
		}
	}

	owner := gjson.Get(string(userIDResult), "userId").String()

	if newOwner == project.Owner {
		// nothing to do
		log.WithFields(event.Fields{
			"projectUrn": projectUrn,
		}).Info("Nothing to update.")

		return project, nil
	}

	project.Owner = owner

	// update project
	_, ret = s.PatchHalResource(ctx, projectResourceURL, GetSelfLinkFromHal(projectResult), project)
	if ret != nil {
		log.WithFields(event.Fields{
			"projectUrn": projectUrn,
		}).Error("Failed to update owner for project")
		ret.Message = "Could not update owner for project. Please make sure that you have the correct access rights."
		return project, ret
	}

	log.WithFields(event.Fields{
		"projectUrn": projectUrn,
	}).Info("Project owner updated.")
	return project, nil
}
