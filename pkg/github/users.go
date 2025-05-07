/*
 * WHO: GitHubUserService
 * WHAT: GitHub user operations implementation
 * WHEN: During user interactions
 * WHERE: System Layer 6 (Integration)
 * WHY: To manage GitHub user operations
 * HOW: Using REST API with 7D context awareness
 * EXTENT: All user operations
 */

package github

import (
	"fmt"
	"net/http"
)

// UserService provides methods for working with GitHub users
type UserService struct {
	// WHO: UserServiceManager
	// WHAT: User service structure
	// WHEN: During user operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To manage user access
	// HOW: Using client reference
	// EXTENT: All user operations

	client *Client
}

// NewUserService creates a new user service
func NewUserService(client *Client) *UserService {
	// WHO: UserServiceCreator
	// WHAT: Create user service instance
	// WHEN: During service initialization
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide user operations
	// HOW: Using dependency injection
	// EXTENT: Service lifecycle

	return &UserService{
		client: client,
	}
}

// GetAuthenticated retrieves the authenticated user
func (s *UserService) GetAuthenticated() (*User, error) {
	// WHO: AuthenticatedUserGetter
	// WHAT: Get authenticated user details
	// WHEN: During authentication verification
	// WHERE: System Layer 6 (Integration)
	// WHY: To verify current user
	// HOW: Using GitHub API
	// EXTENT: Current user data

	user := new(User)
	err := s.client.Request(http.MethodGet, UserEndpoint, nil, user)
	return user, err
}

// Get retrieves a user by username
func (s *UserService) Get(username string) (*User, error) {
	// WHO: UserGetter
	// WHAT: Get user details
	// WHEN: During user lookup
	// WHERE: System Layer 6 (Integration)
	// WHY: To access user information
	// HOW: Using GitHub API
	// EXTENT: Single user data

	path := fmt.Sprintf("%s/%s", UserEndpoint, username)
	user := new(User)
	err := s.client.Request(http.MethodGet, path, nil, user)
	return user, err
}

// ListFollowers retrieves followers for a user
func (s *UserService) ListFollowers(username string) ([]User, error) {
	// WHO: FollowerLister
	// WHAT: List user followers
	// WHEN: During follower enumeration
	// WHERE: System Layer 6 (Integration)
	// WHY: To list followers
	// HOW: Using GitHub API
	// EXTENT: Multiple user data

	path := fmt.Sprintf("%s/%s/followers", UserEndpoint, username)
	var followers []User
	err := s.client.Request(http.MethodGet, path, nil, &followers)
	return followers, err
}

// ListFollowing retrieves users that a user is following
func (s *UserService) ListFollowing(username string) ([]User, error) {
	// WHO: FollowingLister
	// WHAT: List users being followed
	// WHEN: During following enumeration
	// WHERE: System Layer 6 (Integration)
	// WHY: To list followed users
	// HOW: Using GitHub API
	// EXTENT: Multiple user data

	path := fmt.Sprintf("%s/%s/following", UserEndpoint, username)
	var following []User
	err := s.client.Request(http.MethodGet, path, nil, &following)
	return following, err
}

// GetContext retrieves the 7D context for user operations
func (s *UserService) GetContext() map[string]interface{} {
	// WHO: ContextProvider
	// WHAT: Get user service context
	// WHEN: During context operations
	// WHERE: System Layer 6 (Integration)
	// WHY: To provide 7D awareness
	// HOW: Using context mapping
	// EXTENT: Service context

	if s.client == nil {
		return nil
	}

	// Create a 7D context map for user operations
	return map[string]interface{}{
		"who":    "UserService",
		"what":   "UserOperation",
		"when":   "CurrentTime",
		"where":  "SystemLayer6",
		"why":    "UserManagement",
		"how":    "GitHubAPI",
		"extent": "UserContext",
	}
}
