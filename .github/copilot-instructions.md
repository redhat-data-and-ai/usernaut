# GitHub Copilot Instructions for Usernaut

## Project Overview

Usernaut is a Kubernetes operator built using Go 1.24.2 and controller-runtime. It manages user and team synchronization across multiple backends (LDAP, Fivetran, Red Hat Rover) with caching capabilities.

## Go Development Best Practices

### Code Style and Formatting

- **Always run `gofmt`** on all Go files before committing
- Use `goimports` to automatically manage imports
- Follow the existing code structure and naming conventions
- Use camelCase for variable and function names, PascalCase for exported types
- Keep line length under reasonable limits (current project uses `lll` linter)

### Error Handling

- Always handle errors explicitly - never ignore them
- Use descriptive error messages that include context
- Wrap errors with additional context using `fmt.Errorf` or error wrapping
- Return errors as the last return value in functions
- Use early returns to reduce nesting

```go
// Good
func processUser(userID string) (*User, error) {
    if userID == "" {
        return nil, errors.New("userID cannot be empty")
    }

    user, err := fetchUser(userID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch user %s: %w", userID, err)
    }

    return user, nil
}
```

### Struct and Interface Design

- Use composition over inheritance
- Keep interfaces small and focused (interface segregation principle)
- Define interfaces where they are used, not where they are implemented
- Use struct embedding for extending functionality
- Add JSON/YAML tags for serialization when needed

```go
// Good interface design
type UserFetcher interface {
    FetchUser(ctx context.Context, id string) (*User, error)
}

type TeamManager interface {
    CreateTeam(ctx context.Context, team *Team) error
    DeleteTeam(ctx context.Context, teamID string) error
}
```

### Context Usage

- Always pass `context.Context` as the first parameter to functions that may block
- Use `context.Background()` for main functions and tests
- Use `context.WithTimeout()` or `context.WithCancel()` for operations with timeouts
- Never store context in structs

### Testing

- Write unit tests for all public functions
- Use table-driven tests for multiple test cases
- Mock external dependencies using interfaces
- Use meaningful test names that describe the scenario
- Follow the AAA pattern: Arrange, Act, Assert

```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   *User
        want    *User
        wantErr bool
    }{
        {
            name:  "valid user creation",
            input: &User{Email: "test@example.com"},
            want:  &User{ID: "123", Email: "test@example.com"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Logging

- Use structured logging with logrus (as used in the project)
- Include relevant context fields in log entries
- Use appropriate log levels (Debug, Info, Warn, Error)
- Log errors with sufficient context but avoid logging the same error multiple times

```go
log := logger.Logger(ctx).WithFields(logrus.Fields{
    "userID": userID,
    "component": "user-service",
})
log.Info("processing user")
```

### Package Organization

- Keep packages focused and cohesive
- Use internal packages for implementation details
- Group related functionality together
- Avoid circular dependencies
- Use clear, descriptive package names

### Performance Considerations

- Use context for cancellation and timeouts
- Implement proper caching strategies (as done with Redis in this project)
- Avoid unnecessary allocations in hot paths
- Use buffered channels appropriately
- Consider using sync.Pool for frequently allocated objects

### Kubernetes Operator Specific Guidelines

- Follow controller-runtime patterns for reconciliation
- Use proper status updates with conditions
- Implement proper finalizers for cleanup
- Handle resource creation idempotently
- Use client-go best practices for API interactions

```go
// Good reconciler pattern
func (r *GroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := r.Log.WithValues("group", req.NamespacedName)

    var group usernautdevv1alpha1.Group
    if err := r.Get(ctx, req.NamespacedName, &group); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Reconciliation logic here

    return ctrl.Result{}, nil
}
```

### Security Best Practices

- Never log sensitive information (passwords, API keys, tokens)
- Validate all inputs
- Use proper authentication and authorization
- Handle secrets securely
- Implement proper RBAC for Kubernetes resources

### Code Quality Tools

This project uses golangci-lint with the following enabled linters:

- `dupl` - Check for code duplication
- `errcheck` - Check for unchecked errors
- `copyloopvar` - Check for loop variable copying issues
- `ginkgolinter` - Ginkgo test framework linting
- `goconst` - Check for repeated strings that could be constants
- `gocyclo` - Check cyclomatic complexity
- `gofmt` - Check formatting
- `goimports` - Check import formatting
- `gosimple` - Suggest simplifications
- `govet` - Go vet checks
- `ineffassign` - Check for ineffectual assignments
- `lll` - Line length limit
- `misspell` - Check for misspellings
- `nakedret` - Check for naked returns
- `prealloc` - Check for slice preallocation opportunities
- `revive` - Fast, configurable, extensible, flexible linter
- `staticcheck` - Static analysis checks
- `typecheck` - Type checking
- `unconvert` - Check for unnecessary type conversions
- `unparam` - Check for unused parameters
- `unused` - Check for unused code

### Common Patterns in This Project

- Use dependency injection for external dependencies (LDAP, cache, clients)
- Implement client interfaces for different backends
- Use structured configuration with YAML
- Implement proper health checks and readiness probes
- Use controller-runtime for Kubernetes operators

### Documentation

- Write clear package-level documentation
- Document exported functions and types
- Use examples in documentation when helpful
- Keep README and documentation up to date
- Document configuration options and environment variables

### Git and Commit Practices

- Make atomic commits with clear messages
- Run `make lint test` before committing (enforced by pre-commit hook)
- Follow conventional commit format
- Keep commits focused on single changes
- Write descriptive commit messages

### Make Targets

Use the provided Makefile targets:

- `make test` - Run tests with coverage
- `make lint` - Run linter
- `make build` - Build the binary
- `make fmt` - Format code
- `make vet` - Run go vet
- `make mockgen` - Generate mocks

## Adding New Client Support (GitLab, Atlassian, etc.)

When adding support for a new backend client, follow these comprehensive guidelines:

### 1. Library Selection and Usage

**ALWAYS prefer well-established, official or widely-used libraries** over creating custom HTTP clients:

#### Recommended Libraries by Platform

- **GitLab**: Use `gitlab.com/gitlab-org/api/client-go` (official GitLab Go client)
- **GitHub**: Use `github.com/google/go-github` (official GitHub Go library)
- **Atlassian/Jira**: Use `github.com/andygrunwald/go-jira`
- **Slack**: Use `github.com/slack-go/slack`
- **Microsoft Graph**: Use `github.com/microsoftgraph/msgraph-sdk-go`

#### Library Evaluation Criteria

- Official or community-endorsed libraries
- Active maintenance (recent commits, issue responses)
- Good documentation and examples
- Support for the required API features
- Proper error handling and type safety
- Context support for cancellation/timeouts

```go
// Good: Using official GitLab Go client
import "gitlab.com/gitlab-org/api/client-go"

type GitLabClient struct {
    client   *gitlab.Client
    config   Backend
}

func NewGitLabClient(config Backend) (*GitLabClient, error) {
    client, err := gitlab.NewClient(config.APIToken, gitlab.WithBaseURL(config.BaseURL))
    if err != nil {
        return nil, fmt.Errorf("failed to create GitLab client: %w", err)
    }

    return &GitLabClient{
        client: client,
        config: config,
    }, nil
}
```

#### When to Create Custom HTTP Clients

Only create custom HTTP clients when:

- No suitable library exists for the platform
- Existing libraries don't support required API features
- Performance requirements necessitate custom implementation
- Security requirements conflict with existing libraries

If creating custom clients, use the project's existing `pkg/request/httpclient` package.

### 2. Directory Structure and Package Organization

Create the new client package following the existing pattern:

```bash
pkg/clients/
├── client.go              # Main client interface
├── fivetran/             # Existing client example
├── redhat_rover/         # Existing client example
└── gitlab/               # New client package
    ├── client.go         # Main client implementation
    ├── types.go          # API response types & converters
    ├── users.go          # User management methods
    ├── teams.go          # Team management methods
    ├── team_membership.go # Team membership methods
    └── client_test.go    # Comprehensive tests
```

### 3. Interface Implementation Requirements

The new client MUST implement the complete `clients.Client` interface:

```go
type Client interface {
    // User operations
    FetchAllUsers(ctx context.Context) (map[string]*structs.User, map[string]*structs.User, error)
    FetchUserDetails(ctx context.Context, userID string) (*structs.User, error)
    CreateUser(ctx context.Context, u *structs.User) (*structs.User, error)
    DeleteUser(ctx context.Context, userID string) error

    // Team operations
    FetchAllTeams(ctx context.Context) (map[string]structs.Team, error)
    FetchTeamDetails(ctx context.Context, teamID string) (*structs.Team, error)
    CreateTeam(ctx context.Context, t *structs.Team) (*structs.Team, error)
    DeleteTeam(ctx context.Context, teamID string) error

    // Team membership operations
    AddUserToTeam(ctx context.Context, userID, teamID string) error
    RemoveUserFromTeam(ctx context.Context, userID, teamID string) error
    FetchTeamMembership(ctx context.Context, teamID string) ([]string, error)
}
```

### 4. Configuration Integration

Add new backend configuration to the existing config structure:

```go
// In pkg/config/config.go
type Backend struct {
    Name     string            `yaml:"name"`
    Type     string            `yaml:"type"`     // "gitlab", "atlassian", etc.
    Enabled  bool              `yaml:"enabled"`
    BaseURL  string            `yaml:"base_url"`
    APIToken string            `yaml:"api_token"`
    Config   map[string]string `yaml:"config"`   // Backend-specific config
}
```

Update the client factory in `pkg/clients/client.go`:

```go
func New(name, backendType string, backendMap map[string]map[string]Backend) (Client, error) {
    switch backendType {
    case "fivetran":
        return fivetran.NewFivetranClient(/* ... */)
    case "redhat_rover":
        return redhatrover.NewRedHatRoverClient(/* ... */)
    case "gitlab":
        return gitlab.NewGitLabClient(/* ... */)
    case "atlassian":
        return atlassian.NewAtlassianClient(/* ... */)
    default:
        return nil, fmt.Errorf("unsupported backend type: %s", backendType)
    }
}
```

### 5. Data Type Mapping and Conversion

Create proper mapping between library types and internal structs:

```go
// In types.go - Conversion methods for library types
func GitLabUserToUser(gu *gitlab.User) *structs.User {
    return &structs.User{
        ID:       strconv.Itoa(gu.ID),
        Username: gu.Username,
        Email:    gu.Email,
        Name:     gu.Name,
        Active:   gu.State == "active",
    }
}

func GitLabGroupToTeam(gg *gitlab.Group) *structs.Team {
    return &structs.Team{
        ID:          strconv.Itoa(gg.ID),
        Name:        gg.Name,
        Description: gg.Description,
        Path:        gg.Path,
    }
}
```

### 6. Implementation Example with GitLab Library

```go
func (c *GitLabClient) FetchAllUsers(ctx context.Context) (map[string]*structs.User, map[string]*structs.User, error) {
    allUsersById := make(map[string]*structs.User)
    allUsersByEmail := make(map[string]*structs.User)

    opt := &gitlab.ListUsersOptions{
        ListOptions: gitlab.ListOptions{
            PerPage: 100,
            Page:    1,
        },
        Active: gitlab.Bool(true),
    }

    for {
        users, resp, err := c.client.Users.ListUsers(opt, gitlab.WithContext(ctx))
        if err != nil {
            return nil, nil, fmt.Errorf("failed to fetch users: %w", err)
        }

        for _, user := range users {
            internalUser := GitLabUserToUser(user)
            allUsersById[internalUser.ID] = internalUser
            allUsersByEmail[internalUser.Email] = internalUser
        }

        if resp.NextPage == 0 {
            break
        }
        opt.Page = resp.NextPage
    }

    return allUsersById, allUsersByEmail, nil
}
```

### 7. Comprehensive Testing Requirements

#### Unit Tests with Library Mocking

- Use library-provided test utilities when available
- Mock library methods, not HTTP endpoints
- Test all interface methods with table-driven tests
- Test error scenarios from the library

```go
func TestGitLabClient_FetchAllUsers(t *testing.T) {
    tests := []struct {
        name          string
        mockUsers     []*gitlab.User
        mockError     error
        expectedCount int
        expectError   bool
    }{
        {
            name: "successful user fetch",
            mockUsers: []*gitlab.User{
                {ID: 1, Username: "user1", Email: "user1@example.com"},
                {ID: 2, Username: "user2", Email: "user2@example.com"},
            },
            expectedCount: 2,
            expectError:   false,
        },
        {
            name:          "API error",
            mockError:     errors.New("API error"),
            expectedCount: 0,
            expectError:   true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create mock client or use library test utilities
            // Test implementation
        })
    }
}
```

### 8. Edge Cases and Error Scenarios

#### Handle These Scenarios

- **Library-specific Errors**: Understand and handle library-specific error types
- **Pagination**: Use library's pagination features correctly
- **Rate Limiting**: Leverage library's rate limiting if available
- **Authentication**: Handle token refresh using library features
- **API Versioning**: Use library's API version support
- **Context Cancellation**: Ensure library calls respect context cancellation
- **Partial Failures**: Handle scenarios where library operations partially succeed
- **Resource Limits**: Handle large datasets efficiently using library features

```go
// Example: Using GitLab library's built-in pagination and error handling
func (c *GitLabClient) FetchAllTeams(ctx context.Context) (map[string]structs.Team, error) {
    allTeams := make(map[string]structs.Team)

    opt := &gitlab.ListGroupsOptions{
        ListOptions: gitlab.ListOptions{PerPage: 100},
    }

    for {
        groups, resp, err := c.client.Groups.ListGroups(opt, gitlab.WithContext(ctx))
        if err != nil {
            // Handle GitLab-specific errors
            if gitlab.IsRateLimited(err) {
                return nil, fmt.Errorf("rate limited by GitLab API: %w", err)
            }
            return nil, fmt.Errorf("failed to fetch groups: %w", err)
        }

        for _, group := range groups {
            team := GitLabGroupToTeam(group)
            allTeams[team.ID] = *team
        }

        if resp.NextPage == 0 {
            break
        }
        opt.Page = resp.NextPage

        // Check for context cancellation
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }
    }

    return allTeams, nil
}
```

### 9. Security and Best Practices

- Use library's built-in security features
- Leverage library's authentication mechanisms
- Follow library's recommended practices for credential handling
- Use library's TLS/SSL configuration options
- Enable library's built-in request/response validation

### 10. Dependency Management

- Add new library dependencies to `go.mod`
- Ensure version compatibility with existing dependencies
- Document minimum required version of the library
- Consider library's own dependencies for conflicts
- Pin to specific versions for stability

## Review Checklist

### General Code Review

- [ ] All errors are properly handled
- [ ] Code is properly formatted (gofmt/goimports)
- [ ] Tests are included for new functionality
- [ ] Logging includes appropriate context
- [ ] No sensitive information is logged
- [ ] Interfaces are used appropriately
- [ ] Context is passed correctly
- [ ] Code follows existing patterns
- [ ] Documentation is updated if needed
- [ ] Linter checks pass

### New Client Implementation Review

- [ ] Uses established, well-maintained library (e.g., `gitlab.com/gitlab-org/api/client-go`)
- [ ] Library choice is justified and documented
- [ ] Implements complete `clients.Client` interface
- [ ] Follows established directory structure
- [ ] Includes comprehensive unit tests with proper mocking
- [ ] Handles all documented edge cases
- [ ] Implements proper error handling using library error types
- [ ] Includes integration tests (where applicable)
- [ ] Follows security best practices
- [ ] Includes proper configuration integration
- [ ] Documents API requirements and limitations
- [ ] Handles pagination using library features
- [ ] Leverages library's rate limiting capabilities
- [ ] Includes proper logging with context
- [ ] Uses library's validation features
- [ ] Efficient for large datasets using library optimizations
- [ ] Uses library's context support for cancellation
- [ ] Updates client factory in `pkg/clients/client.go`
- [ ] Dependencies added to `go.mod` with appropriate versions
- [ ] No unnecessary custom HTTP clients created
- [ ] Ensure `go mod vendor` is executed and `vendor/` directory is upto date
