package rexos

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/roboticeyes/gococo/event"
	"github.com/roboticeyes/gococo/status"
	"github.com/tidwall/gjson"
)

// UserInformation is a container for detailed user information
type UserInformation struct {
	FirstName *string `json:"firstName" example:"Josef"`
	LastName  *string `json:"lastName" example:"Huber"`
	LastLogin string  `json:"lastLogin"`
	Email     *string `json:"email" example:"josef.huber@gasse.at"`
	UserID    *string `json:"userId" example:"userId"`
	UserName  *string `json:"userName" example:"userName"`
}

// UserStatistics is a container for global project information for the user
type UserStatistics struct {
	NumberOfProjects              int    `json:"numberOfProjects"`
	TotalUsedDiskSpace            uint64 `json:"totalUsedDiskSpace"`
	MaxTotalUsedDiskSpace         uint64 `json:"maxTotalUsedDiskSpace"`
	NumberOfPubicShareActions     int    `json:"numberOfPublicShareActions"`
	MaxNumberOfPublicShareActions int    `json:"maxNumberOfPublicShareActions"`
}

// License is a container for license object
type License struct {
	LicenseName    string `json:"licenseName"`
	ActivationDate string `json:"activationDate"`
	ExpirationDate string `json:"expirationDate"`
	Urn            string `json:"urn"`
	SubscriptionID string `json:"subscriptionId"`
	LicenseKey     string `json:"licenseKey"`
}

// UserLicenses contains a list of all licenses assigned to the user
type UserLicenses struct {
	UserLicenses []License `json:"userLicenses"`
}

// GetUserInformation returns current user information
func (s *Service) GetUserInformation(ctx context.Context, resourceURL string) (UserInformation, *status.Status) {
	currentUser, _, ret := s.GetCurrentUser(ctx, resourceURL)
	return currentUser, ret
}

//GetCurrentUser returns current user information and a string representing the user
func (s *Service) GetCurrentUser(ctx context.Context, resourceURL string) (UserInformation, string, *status.Status) {

	query := resourceURL + "/current"

	currentUserResult, ret := s.GetHalResourceNoXF(ctx, "User", query)
	if ret != nil {
		log.WithFields(event.Fields{
			"status": ret,
			"query":  query,
		}).Error("Failed to get current user")

		ret.Message = "Could not get current user. Please make sure you have the correct access rights."
		return UserInformation{}, "", ret
	}

	// get user properties name, email
	userResultLink := StripTemplateParameter(gjson.Get(string(currentUserResult), "_links.user.href").String())
	userResult, ret := s.GetHalResourceNoXF(ctx, "User", userResultLink)
	if ret != nil {
		log.WithFields(event.Fields{
			"status": ret,
			"query":  userResultLink,
		}).Error("Failed to get user information")

		ret.Message = "Could not get user information. Please make sure you have the correct access rights."
		return UserInformation{}, "", ret
	}

	var userInformation UserInformation
	json.Unmarshal(userResult, &userInformation)

	return userInformation, string(currentUserResult), nil
}

// GetUserStatistics returns statitisc information for the current user
func (s *Service) GetUserStatistics(ctx context.Context, resourceURL string) (UserStatistics, *status.Status) {
	// get userID
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		log.WithFields(event.Fields{
			"error": err.Error(),
		}).Error("Failed to get userID")

		return UserStatistics{}, status.NewStatus([]byte{}, http.StatusInternalServerError, "Cannot get userID ")
	}

	query := resourceURL + "/statisticsByUser?userId=" + userID
	userStatisticsResult, ret := s.GetHalResource(ctx, "Project", query)
	if ret != nil {
		log.WithFields(event.Fields{
			"status": ret,
			"query":  query,
		}).Error("Failed to get current user statistics")

		ret.Message = "Could not get current user statistics. Please make sure you have the correct access rights."
		return UserStatistics{}, ret
	}
	var stat UserStatistics
	json.Unmarshal(userStatisticsResult, &stat)

	return stat, nil
}

// GetUserLicenses returns the licenses for the current user
func (s *Service) GetUserLicenses(ctx context.Context, resourceURL string) (UserLicenses, *status.Status) {
	query := resourceURL + "/current"

	currentUserResult, ret := s.GetHalResource(ctx, "User", query)
	if ret != nil {
		log.WithFields(event.Fields{
			"status": ret,
			"query":  query,
		}).Error("Failed to get current user")

		ret.Message = "Could not get current user. Please make sure you have the correct access rights."
		return UserLicenses{}, ret
	}

	userLicensesLink := StripTemplateParameter(gjson.Get(string(currentUserResult), "_links.userLicenses.href").String())
	userLicensesResult, ret := s.GetHalResource(ctx, "User", userLicensesLink)
	if ret != nil {
		log.WithFields(event.Fields{
			"status": ret,
			"query":  userLicensesLink,
		}).Error("Failed to get user licenses")

		ret.Message = "Could not get user licenses. Please make sure you have the correct access rights."
		return UserLicenses{}, ret
	}
	userLicenseList := gjson.Get(string(userLicensesResult), "_embedded.userLicenses").Array()

	list := make([]License, 0)
	for _, l := range userLicenseList {
		var userLicense License
		json.Unmarshal([]byte(l.Raw), &userLicense)

		licenseLink := gjson.Get(l.String(), "_links.license.href").String()
		licenseResult, ret := s.GetHalResource(ctx, "User", licenseLink)
		if ret != nil {
			log.WithFields(event.Fields{
				"status": ret,
				"query":  licenseLink,
			}).Error("Failed to get license")

			ret.Message = "Could not get license. Please make sure you have the correct access rights."
			return UserLicenses{}, ret
		}
		userLicense.LicenseName = gjson.Get(string(licenseResult), "name").String()
		userLicense.LicenseKey = gjson.Get(string(licenseResult), "key").String()
		list = append(list, userLicense)
	}

	return UserLicenses{UserLicenses: list}, nil
}
