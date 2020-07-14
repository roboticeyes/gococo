package rexos

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/roboticeyes/gococo/event"
	"github.com/roboticeyes/gococo/status"
	"github.com/tidwall/gjson"
)

// User is a container for user details used for project sharing
type User struct {
	UserName  string `json:"userName"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	UserID    string `json:"userID"`
}

// UserShare describes the type of sharing of a project with a user
type UserShare struct {
	User  User `json:"user"`
	Write bool `json:"write"`
	Read  bool `json:"read"`
}

// Share contains all sharing information for a project
type Share struct {
	PublicShare bool        `json:"publicShare"`
	UserShares  []UserShare `json:"userShares,omitempty"`
}

// GetShare returns the sharing information for a project
func (s *Service) GetShare(ctx context.Context, projectResourceURL, userResourceURL, projectUrn string) (Share, *status.Status) {
	var share Share

	// get number from project urn robotic-eyes:project:12345 -> 12345
	parts := strings.Split(projectUrn, ":")
	if len(parts) < 3 {
		log.WithFields(event.Fields{
			"projectUrn": projectUrn,
		}).Error("Failed to get number from urn")

		return Share{}, status.NewStatus([]byte{}, http.StatusInternalServerError, "Cannot get number from projectUrn ")
	}
	projectNumber := parts[2]

	// get public sharing information
	query := projectResourceURL + "/" + projectNumber + "/publicShare"
	publicShareResult, ret := s.GetHalResource(ctx, "Projects", query)
	if ret != nil {
		log.WithFields(event.Fields{
			"status":     ret,
			"projectUrn": projectUrn,
			"query":      query,
		}).Error("Failed to get public share information")

		ret.Message = "Cannot not get public share information for the project. Please make sure you have the correct access rights."
		return Share{}, ret
	}
	share.PublicShare = gjson.Get(string(publicShareResult), "shared").Bool()

	// get user sharing information
	query = projectResourceURL + "/" + projectNumber + "/userShares"
	userShareResult, ret := s.GetHalResource(ctx, "Projects", query)
	if ret != nil {
		log.WithFields(event.Fields{
			"status":     ret,
			"projectUrn": projectUrn,
			"query":      query,
		}).Error("Failed to get user share information")

		ret.Message = "Cannot not get user share information for the project. Please make sure you have the correct access rights."
		return Share{}, ret
	}
	userShares := gjson.Get(string(userShareResult), "_embedded.userShares").Array()
	for _, u := range userShares {
		var userShare UserShare

		// find user
		userID := gjson.Get(u.String(), "user").String()
		query = userResourceURL + "/search/findByUserId?userId=" + userID
		userResult, ret := s.GetHalResource(ctx, "Users", query)
		if ret != nil {
			log.WithFields(event.Fields{
				"status": ret,
				"userID": userID,
				"query":  query,
			}).Error("Failed to get user information")

			ret.Message = "Cannot not get user information. Please make sure you have the correct access rights."
			return Share{}, ret
		}
		var user User
		json.Unmarshal(userResult, &user)
		userShare.User = user

		if gjson.Get(u.String(), "action").String() == "READ" {
			userShare.Read = true
			userShare.Write = false
		} else {
			userShare.Read = false
			userShare.Write = true
		}
		share.UserShares = append(share.UserShares, userShare)
	}

	return share, nil
}

// UpdateShare updates the project sharing
func (s *Service) UpdateShare(ctx context.Context, projectResourceURL, userResourceURL, projectUrn string, share Share) (Share, *status.Status) {
	return share, nil
}

// DeleteUserShare deletes a user sharing of a project
func (s *Service) DeleteUserShare(ctx context.Context, resourceURL, projectUrn, userID string) *status.Status {
	return nil
}
