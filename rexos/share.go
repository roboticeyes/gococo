package rexos

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/roboticeyes/gococo/event"
	"github.com/roboticeyes/gococo/status"
	"github.com/tidwall/gjson"
)

const (
	readAction  = "READ"
	writeAction = "WRITE"
)

// User is a container for user details used for project sharing
type User struct {
	UserName  string `json:"userName,omitempty"`
	Email     string `json:"email,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	UserID    string `json:"userID,omitempty"`
}

// UserShare describes the type of sharing of a project with a user
// data structure for frontend
type UserShare struct {
	User  User `json:"user"`
	Write bool `json:"write"`
	Read  bool `json:"read"`
}

// UserShareReduced describes the sharing action ("READ"| "WRITE") for the given user
// data structure for backend
type UserShareReduced struct {
	UserID string `json:"user"`
	Action string `json:"action" example:"READ | WRITE"`
}

// PublicShare contains the information about public sharing
// data structure for backend
type PublicShare struct {
	Shared bool `json:"shared"`
}

// Share contains all sharing information for a project
type Share struct {
	PublicShare *bool       `json:"publicShare,omitempty"`
	UserShares  []UserShare `json:"userShares,omitempty"`
}

// GetShare returns the sharing information for a project
func (s *Service) GetShare(ctx context.Context, projectResourceURL, userResourceURL, projectUrn string) (Share, *status.Status) {
	var share Share

	projectNumber, ret := GetNumberFromUrn(projectUrn)
	if ret != nil {
		return share, ret
	}

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
	val := gjson.Get(string(publicShareResult), "shared").Bool()
	share.PublicShare = &val

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
		userResult, ret := s.GetHalResourceWithServiceUser(ctx, "Users", query)
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

		if gjson.Get(u.String(), "action").String() == readAction {
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

// UpdateShare updates the project sharing (public sharing)
func (s *Service) UpdateShare(ctx context.Context, projectResourceURL, userResourceURL, projectUrn string, share Share) (Share, *status.Status) {
	projectNumber, ret := GetNumberFromUrn(projectUrn)
	if ret != nil {
		return share, ret
	}

	// update public sharing information
	query := projectResourceURL + "/" + projectNumber + "/publicShare"
	val := share.PublicShare
	_, ret = s.PatchHalResource(ctx, "Projects", query, PublicShare{Shared: *val})
	if ret != nil {
		log.WithFields(event.Fields{
			"status":     ret,
			"projectUrn": projectUrn,
			"query":      query,
		}).Error("Failed to update public share information")

		ret.Message = "Cannot not update public share information for the project. Please make sure you have the correct access rights."
		return Share{}, ret
	}
	return share, nil
}

// CreateOrUpdateUserShare shares a project with a given user
func (s *Service) CreateOrUpdateUserShare(ctx context.Context, projectResourceURL, userResourceURL, projectUrn string, userShare UserShare) (UserShare, *status.Status) {
	projectNumber, ret := GetNumberFromUrn(projectUrn)
	if ret != nil {
		return userShare, ret
	}

	var query string
	if userShare.User.Email != "" {
		query = userResourceURL + "/search/findUserIdByEmail?email=" + userShare.User.Email
	} else {
		if userShare.User.UserName != "" {
			query = userResourceURL + "/search/findUserIdByUsername?username=" + userShare.User.UserName
		} else {
			log.WithFields(event.Fields{
				"projectUrn": projectUrn,
			}).Error("No email address or username for user sharing.")
			return userShare, status.NewStatus([]byte{}, http.StatusBadRequest, "No email address or username found.")
		}
	}

	userResult, ret := s.GetHalResource(ctx, "Users", query)
	if ret != nil {
		log.WithFields(event.Fields{
			"status":     ret,
			"projectUrn": projectUrn,
			"email":      userShare.User.Email,
			"query":      query,
		}).Error("Failed to get userId by email")

		ret.Message = "Cannot find userId. Please make sure you have the correct access rights."
		return UserShare{}, ret
	}
	json.Unmarshal(userResult, &userShare.User)

	action := writeAction
	if userShare.Read {
		action = readAction
	}

	share := UserShareReduced{UserID: userShare.User.UserID, Action: action}

	// update user sharing
	query = projectResourceURL + "/" + projectNumber + "/userShares"
	_, ret = s.CreateHalResource(ctx, "Projects", query, share)
	if ret != nil {
		if ret.Code == 409 {
			// user share already exist, update it
			query = projectResourceURL + "/" + projectNumber + "/userShares/" + userShare.User.UserID
			_, ret = s.PatchHalResource(ctx, "Projects", query, share)
			if ret != nil {
				log.WithFields(event.Fields{
					"status":     ret,
					"projectUrn": projectUrn,
					"query":      query,
				}).Error("Failed to update user share information")

				ret.Message = "Cannot not update user share information for the project. Please make sure you have the correct access rights."
				return UserShare{}, ret
			}
		} else {
			log.WithFields(event.Fields{
				"status":     ret,
				"projectUrn": projectUrn,
				"query":      query,
			}).Error("Failed to create user share information")

			ret.Message = "Cannot not create user share information for the project. Please make sure you have the correct access rights."
			return UserShare{}, ret
		}
	}

	return userShare, nil
}

// DeleteUserShare deletes a user share of a project
func (s *Service) DeleteUserShare(ctx context.Context, resourceURL, projectUrn, userID string) *status.Status {
	projectNumber, ret := GetNumberFromUrn(projectUrn)
	if ret != nil {
		return ret
	}

	query := resourceURL + "/" + projectNumber + "/userShares/" + userID
	ret = s.DeleteHalResource(ctx, "Projects", query)
	if ret != nil {
		log.WithFields(event.Fields{
			"status":     ret,
			"projectUrn": projectUrn,
			"query":      query,
		}).Error("Failed to delete user share")

		ret.Message = "Cannot not delete user share for the project. Please make sure you have the correct access rights."
		return ret
	}
	return nil
}
