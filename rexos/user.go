package rexos

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/roboticeyes/gococo/event"
	"github.com/roboticeyes/gococo/status"
	"github.com/tidwall/gjson"
)

// Address user address
type Address struct {
	Zip         *string `json:"zip,omitempty"  example:"8010"`
	City        *string `json:"city,omitempty" example:"Graz"`
	Address     *string `json:"address,omitempty" example:"Josef Huber Gasse"`
	State       *string `json:"state,omitempty"`
	CountryName *string `json:"countryname,omitempty" example:"Austria"`
	Country     *string `json:"country,omitempty"`
}

// UserInformation is a container for detailed user information
type UserInformation struct {
	FirstName *string  `json:"firstName,omitempty" example:"Josef"`
	LastName  *string  `json:"lastName,omitempty" example:"Huber"`
	Address   *Address `json:"address,omitempty"`
	Email     *string  `json:"email,omitempty" example:"josef.huber@gasse.at"`
	Company   *string  `json:"company,omitempty" example:"Robotic Eyes"`
	UID       *string  `json:"uid,omitempty"`
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
	Name           string `json:"name"`
	ActivationDate string `json:"activationDate"`
	ExpirationDate string `json:"expirationDate"`
}

// UserLicenses contains a list of all licenses assigned to the user
type UserLicenses struct {
	UserLicenses []License `json:"userLicenses"`
}

// UserDescription struct to update /userDescriptions
type UserDescription struct {
	Address Address `json:"address,omitempty"`
	Company string  `json:"company,omitempty"`
	UID     string  `json:"uid,omitempty"`
}

// UserInfo struct to update /users/current
type UserInfo struct {
	Email     string `json:"email,omitempty"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
}

// GetUserInformation returns detailed user information
func (s *Service) GetUserInformation(ctx context.Context, resourceURL string) (UserInformation, *status.Status) {
	query := resourceURL + "/current"

	currentUserResult, ret := s.GetHalResource(ctx, "User", query)
	if ret != nil {
		log.WithFields(event.Fields{
			"status": ret,
			"query":  query,
		}).Error("Failed to get current user")

		ret.Message = "Could not get current user. Please make sure you have the correct access rights."
		return UserInformation{}, ret
	}

	// get user properties name, email
	userResultLink := StripTemplateParameter(gjson.Get(string(currentUserResult), "_links.user.href").String())
	userResult, ret := s.GetHalResource(ctx, "User", userResultLink)
	if ret != nil {
		log.WithFields(event.Fields{
			"status": ret,
			"query":  userResultLink,
		}).Error("Failed to get user information")

		ret.Message = "Could not get user information. Please make sure you have the correct access rights."
		return UserInformation{}, ret
	}

	var userInformation UserInformation
	json.Unmarshal(userResult, &userInformation)

	// get user properties address, company
	userDescriptionLink := gjson.Get(string(currentUserResult), "_links.userDescription.href").String()
	userDescriptionResult, ret := s.GetHalResource(ctx, "User", userDescriptionLink)
	if ret != nil {
		log.WithFields(event.Fields{
			"status": ret,
			"query":  userDescriptionLink,
		}).Error("Failed to get user description")

		ret.Message = "Could not get user description. Please make sure you have the correct access rights."
		return UserInformation{}, ret
	}
	json.Unmarshal(userDescriptionResult, &userInformation)

	return userInformation, nil
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
		userLicense.Name = gjson.Get(string(licenseResult), "name").String()
		list = append(list, userLicense)
	}

	return UserLicenses{UserLicenses: list}, nil
}

// UpdateUserInformation updates user data like address, ...
func (s *Service) UpdateUserInformation(ctx context.Context, resourceURL string, info UserInformation) (UserInformation, *status.Status) {
	// get current user to get the links for update
	query := resourceURL + "/current"

	currentUserResult, ret := s.GetHalResource(ctx, "User", query)
	if ret != nil {
		log.WithFields(event.Fields{
			"status": ret,
			"query":  query,
		}).Error("Failed to get current user")

		ret.Message = "Could not get current user. Please make sure you have the correct access rights."
		return info, ret
	}

	// update user description (properties address, company, newsletter, language, uid)
	userDescriptionLink := gjson.Get(string(currentUserResult), "_links.userDescription.href").String()
	userDescriptionResult, ret := s.GetHalResource(ctx, "User", userDescriptionLink)
	if ret != nil {
		log.WithFields(event.Fields{
			"status": ret,
			"query":  query,
		}).Error("Failed to get current user description")

		ret.Message = "Could not get current user description. Please make sure you have the correct access rights."
		return info, ret
	}

	var userDescription UserDescription
	json.Unmarshal(userDescriptionResult, &userDescription)

	if info.Company != nil {
		userDescription.Company = *info.Company
	}

	if info.Address != nil {
		if info.Address.Zip != nil {
			userDescription.Address.Zip = info.Address.Zip
		} else {
			info.Address.Zip = userDescription.Address.Zip
		}
		if info.Address.Address != nil {
			userDescription.Address.Address = info.Address.Address
		} else {
			info.Address.Address = userDescription.Address.Address
		}
		if info.Address.State != nil {
			userDescription.Address.State = info.Address.State
		} else {
			info.Address.State = userDescription.Address.State
		}
		if info.Address.City != nil {
			userDescription.Address.City = info.Address.City
		} else {
			info.Address.City = userDescription.Address.City
		}
		if info.Address.Country != nil {
			userDescription.Address.Country = info.Address.Country
		} else {
			info.Address.Country = userDescription.Address.Country
		}
		if info.Address.CountryName != nil {
			userDescription.Address.CountryName = info.Address.CountryName
		} else {
			info.Address.CountryName = userDescription.Address.CountryName
		}
	} else {
		info.Address = &userDescription.Address
	}

	if info.UID != nil {
		userDescription.UID = *info.UID
	}

	_, ret = s.PatchHalResource(ctx, "User", userDescriptionLink, userDescription)
	if ret != nil {
		log.WithFields(event.Fields{
			"status": ret,
			"query":  userDescriptionLink,
		}).Error("Failed to update user description")

		ret.Message = "Could not update user description. Please make sure you have the correct access rights."
		return info, ret
	}

	// update user information  (properties userId, username, firstName, lastName, lastLogin)
	userInfoLink := StripTemplateParameter(gjson.Get(string(currentUserResult), "_links.self.href").String())

	var userInfo UserInfo

	if info.Email != nil {
		userInfo.Email = *info.Email
	} else {
		info.Email = &userInfo.Email
	}

	if info.FirstName != nil {
		userInfo.FirstName = *info.FirstName
	} else {
		info.FirstName = &userInfo.FirstName
	}

	if info.LastName != nil {
		userInfo.LastName = *info.LastName
	} else {
		info.LastName = &userInfo.LastName
	}

	_, ret = s.PatchHalResource(ctx, "User", userInfoLink, userInfo)
	if ret != nil {
		log.WithFields(event.Fields{
			"status": ret,
			"query":  userInfoLink,
		}).Error("Failed to update user information")

		ret.Message = "Could not update user information. Please make sure you have the correct access rights."
		return info, ret
	}
	return info, nil
}
