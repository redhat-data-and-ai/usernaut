package fivetran

import (
	"context"
	"fmt"
	"strings"

	"github.com/fivetran/go-fivetran/users"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
)

// Fetches all the users onboarded over Fivetran
// returns 2 maps where:
// 1st map will have ID as key in order to map with team membership response
// and 2nd will have email as key
func (fc *FivetranClient) FetchAllUsers(ctx context.Context) (
	map[string]*structs.User, map[string]*structs.User, error) {
	log := logger.Logger(ctx).WithField("service", "fivetran")

	usersEmailMap := make(map[string]*structs.User, 0)
	userIDMap := make(map[string]*structs.User, 0)

	log.Info("fetching all the users")
	resp, err := fc.fivetranClient.NewUsersList().Do(ctx)
	if err != nil {
		log.WithField("response", resp.CommonResponse).WithError(err).Error("error fetching list of users")
		return nil, nil, err
	}
	for _, item := range resp.Data.Items {
		usersEmailMap[item.Email] = userDetailsFromResponse(item)
		userIDMap[item.ID] = userDetailsFromResponse(item)
	}
	cursor := resp.Data.NextCursor

	// paginate over the cursor until last page
	for len(cursor) != 0 {
		resp, err := fc.fivetranClient.NewUsersList().Cursor(cursor).Do(ctx)
		if err != nil {
			log.WithField("response", resp.CommonResponse).WithError(err).Error("error fetching list of users")
			return nil, nil, err
		}
		for _, item := range resp.Data.Items {
			usersEmailMap[item.Email] = userDetailsFromResponse(item)
			userIDMap[item.ID] = userDetailsFromResponse(item)
		}
		cursor = resp.Data.NextCursor
	}

	log.WithFields(logrus.Fields{
		"total_user_count": len(usersEmailMap),
		"response":         resp.CommonResponse,
	}).Info("found users")

	return usersEmailMap, userIDMap, nil
}

// Onboards the user on fivetran
func (fc *FivetranClient) CreateUser(ctx context.Context, u *structs.User) (*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "fivetran",
		"user":    u,
	})

	log.Info("inviting user")
	resp, err := fc.fivetranClient.NewUserInvite().
		Email(u.Email).
		FamilyName(u.LastName).
		GivenName(u.FirstName).
		Do(ctx)
	if err != nil {
		log.WithField("response", resp.CommonResponse).WithError(err).Error("error inviting the user")

		// 409 status code conflict
		if strings.Contains(err.Error(), "status code: 409") ||
			(resp.CommonResponse.Code == "UserExists") {
			log.Info("user already exists, fetching existing user details")

			usersByEmail, _, fetchErr := fc.FetchAllUsers(ctx)
			if fetchErr != nil {
				log.WithError(fetchErr).Error("failed to fetch users to find existing user")
				return &structs.User{}, err
			}

			allUserDetails := make([]map[string]interface{}, 0, len(usersByEmail))
			allEmails := make([]string, 0, len(usersByEmail))
			for email, user := range usersByEmail {
				allEmails = append(allEmails, email)
				allUserDetails = append(allUserDetails, map[string]interface{}{
					"key":         email,
					"id":          user.ID,
					"email":       user.Email,
					"username":    user.UserName,
					"displayName": user.DisplayName,
				})
			}
			log.WithFields(logrus.Fields{
				"searchingFor": u.Email,
				"allEmails":    allEmails,
				"userDetails":  allUserDetails,
			}).Info("debugging email lookup with full user details")

			if existingUser, found := usersByEmail[u.Email]; found {
				log.WithField("existingUser", existingUser).Info("found existing user (exact match)")
				return existingUser, nil
			}

			lowerEmail := strings.ToLower(u.Email)
			for email, user := range usersByEmail {
				if strings.ToLower(email) == lowerEmail {
					log.WithFields(logrus.Fields{
						"searchedFor":  u.Email,
						"foundEmail":   email,
						"existingUser": user,
					}).Info("found existing user (case-insensitive match)")
					return user, nil
				}
			}

			for mapKey, user := range usersByEmail {
				if strings.ToLower(user.Email) == lowerEmail {
					log.WithFields(logrus.Fields{
						"searchedFor":      u.Email,
						"foundInUserField": user.Email,
						"mapKey":           mapKey,
						"existingUser":     user,
					}).Info("found existing user (by user.Email field)")
					return user, nil
				}
			}

			if idx := strings.Index(u.Email, "@"); idx > 0 {
				username := u.Email[:idx]
				lowerUsername := strings.ToLower(username)

				for mapKey, user := range usersByEmail {
					if strings.ToLower(mapKey) == lowerUsername {
						log.WithFields(logrus.Fields{
							"searchedFor":       u.Email,
							"extractedUsername": username,
							"foundMapKey":       mapKey,
							"existingUser":      user,
						}).Info("found existing user (by extracted username)")
						return user, nil
					}
				}

				for mapKey, user := range usersByEmail {
					if strings.ToLower(user.UserName) == lowerUsername {
						log.WithFields(logrus.Fields{
							"searchedFor":       u.Email,
							"extractedUsername": username,
							"foundUserName":     user.UserName,
							"mapKey":            mapKey,
							"existingUser":      user,
						}).Info("found existing user (by user.UserName field)")
						return user, nil
					}
				}
			}

			log.WithFields(logrus.Fields{
				"searchedFor": u.Email,
				"totalUsers":  len(usersByEmail),
				"allEmails":   allEmails,
				"userDetails": allUserDetails,
			}).Error("user should exist but not found in user list")
		}

		return &structs.User{}, err
	}
	log.WithField("response", resp).Info("invite sent to the user")

	return userDetailsFromResponse(resp.Data), nil
}

// Fetches user details based on userID (fivetran ID)
func (fc *FivetranClient) FetchUserDetails(ctx context.Context, userID string) (*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "fivetran",
		"userID":  userID,
	})
	log.Info("fetching user details by ID")
	resp, err := fc.fivetranClient.NewUserDetails().UserID(userID).Do(ctx)
	if err != nil {
		log.WithField("response", resp.CommonResponse).WithError(err).Error("error fetching user details")
		return &structs.User{}, err
	}

	log.Info("found user details")
	return userDetailsFromResponse(resp.Data), nil
}

// Updates user details based on userID (fivetran ID)
func (fc *FivetranClient) UpdateUser(ctx context.Context, u *structs.User) (*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "fivetran",
		"user":    u,
	})

	log.Info("updating user details")

	resp, err := fc.fivetranClient.NewUserModify().
		UserID(u.ID).
		FamilyName(u.LastName).
		GivenName(u.FirstName).
		Do(ctx)
	if err != nil {
		log.WithField("response", resp.CommonResponse).WithError(err).Error("error updating the user")
		return &structs.User{}, err
	}

	return userDetailsFromResponse(resp.Data), nil

}

// Offboards the user from fivetran based on userID (fivetran ID)
func (fc *FivetranClient) DeleteUser(ctx context.Context, userID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "fivetran",
		"userID":  userID,
	})

	log.Info("dropping the user")

	resp, err := fc.fivetranClient.NewUserDelete().UserID(userID).Do(ctx)
	if err != nil {
		log.WithFields(logrus.Fields{
			"code":    resp.Code,
			"message": resp.Message,
		}).WithError(err).Error("error deleting the user")
		return err
	}
	log.Info("user deleted successfully")
	return nil
}

// converts users.UserDetailsData to structs.User
func userDetailsFromResponse(u users.UserDetailsData) *structs.User {
	return &structs.User{
		ID:          u.ID,
		Email:       u.Email,
		FirstName:   u.GivenName,
		LastName:    u.FamilyName,
		DisplayName: fmt.Sprintf("%s %s", u.GivenName, u.FamilyName),
		Role:        u.Role,
	}
}
