# GitHub Copilot Instructions for Usernaut

## Project Overview

Usernaut is a Kubernetes operator built using Go 1.24.2 and controller-runtime. It manages user and team synchronization across multiple backends (LDAP, Fivetran, Red Hat Rover, GitLab, Snowflake) with Redis-based caching capabilities. The operator follows a namespace-scoped pattern, watching resources in the `usernaut` namespace (configurable via `WATCHED_NAMESPACE`).

### Key Technologies

- **Go**: 1.24.2
- **Framework**: controller-runtime (Kubernetes operator SDK)
- **Cache**: Redis with in-memory fallback
- **Logging**: logrus with JSON formatting
- **HTTP Client**: Custom client with Hystrix circuit breaker
- **Testing**: Ginkgo/Gomega for BDD-style tests

## Go Development Best Practices

### Code Style and Formatting

- **Always run `gofmt`** on all Go files before committing
- Use `goimports` to automatically manage imports
- Follow the existing code structure and naming conventions
- Use camelCase for variable and function names, PascalCase for exported types
- Keep line length under reasonable limits (current project uses `lll` linter with configurable exceptions for `api/*` paths)
- **Copyright Headers**: All new files must include the Apache 2.0 license header (see `hack/boilerplate.go.txt`)

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

### Configuration Management

This project uses YAML-based configuration with environment-specific overrides.

#### Configuration Structure

```yaml
app:
  name: usernaut
  version: 0.0.1
  environment: local
  debug: true

ldap:
  url: "ldaps://ldap.example.com:636"
  bindDN: "uid=service,ou=users,dc=example,dc=com"
  bindPassword: "${LDAP_PASSWORD}" # Environment variable substitution
  baseDN: "ou=users,dc=example,dc=com"

cache:
  type: redis # or "inmemory"
  redis:
    address: "localhost:6379"
    password: "${REDIS_PASSWORD}"
    db: 0

backends:
  - name: fivetran
    type: fivetran
    enabled: true
    connection:
      apikey: "${FIVETRAN_API_KEY}"
      apisecret: "${FIVETRAN_API_SECRET}"

  - name: gitlab
    type: gitlab
    enabled: true
    depends_on:
      name: rover
      type: rover
    connection:
      url: "https://gitlab.example.com"
      token: "${GITLAB_TOKEN}"

httpClient:
  connectionPoolConfig:
    maxIdleConns: 100
    maxConnsPerHost: 100
    idleConnTimeout: 90
  hystrixResiliencyConfig:
    timeout: 30000
    maxConcurrentRequests: 100
    errorPercentThreshold: 50

apiServer:
  address: ":8080"
  auth:
    enabled: true
    basic_users:
      - username: admin
        password: "${API_ADMIN_PASSWORD}"
  cors:
    allowed_origins:
      - "http://localhost:3000"
```

#### Configuration Best Practices

**Loading Configuration**:

```go
// Load config with environment variable
appConfig, err := config.LoadConfig(os.Getenv("APP_ENV"))
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}

// Or use default loading (respects APP_ENV)
appConfig, err := config.GetConfig()
```

**Accessing Configuration**:

```go
// Access nested configuration
ldapURL := appConfig.LDAP.URL
cacheType := appConfig.Cache.Type

// Access backends by type and name
backend := appConfig.BackendMap["gitlab"]["gitlab"]
apiToken := backend.GetStringConnection("token", "")

// Helper method for safe string access with defaults
timeout := backend.GetStringConnection("timeout", "30")
```

**Configuration Guidelines**:

- **Environment-specific files**: `appconfig/default.yaml`, `appconfig/local.yaml`, `appconfig/<env>.yaml`
- **Environment variable substitution**: Use `${VAR_NAME}` syntax in YAML files
- **Sensitive data**: Always use environment variables for secrets
- **Defaults**: Provide sensible defaults for optional configuration
- **Validation**: Validate configuration on startup, fail fast on missing required values
- **Backend dependencies**: Use `depends_on` to specify dependencies between backends (e.g., GitLab depends on Rover for LDAP sync)

#### Backend Configuration Pattern

When adding new backend support:

```go
case "newbackend":
    // Extract required connection parameters
    apiKey := backend.GetStringConnection("apikey", "")
    baseURL := backend.GetStringConnection("url", "")

    // Validate required parameters
    if apiKey == "" {
        return nil, errors.New("missing required connection parameter: apikey")
    }

    // Get shared configuration (circuit breaker, connection pool)
    appConfig, err := config.GetConfig()
    if err != nil {
        return nil, err
    }

    // Create client with configuration
    return newbackend.NewClient(
        backend.Connection,
        appConfig.HttpClient.ConnectionPoolConfig,
        appConfig.HttpClient.HystrixResiliencyConfig,
    )
```

### Testing

- Write unit tests for all public functions
- Use table-driven tests for multiple test cases
- Mock external dependencies using interfaces
- Use meaningful test names that describe the scenario
- Follow the AAA pattern: Arrange, Act, Assert
- **Use Ginkgo/Gomega** for controller tests (BDD-style)
- **Generate mocks with mockgen**: Run `make mockgen` to regenerate mocks

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
        {
            name:    "empty email",
            input:   &User{Email: ""},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            service := NewUserService(mockClient)

            // Act
            got, err := service.CreateUser(context.Background(), tt.input)

            // Assert
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("CreateUser() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

#### Ginkgo/Gomega Tests for Controllers

```go
var _ = Describe("Group Controller", func() {
    Context("When reconciling a Group resource", func() {
        It("Should create users in all backends", func() {
            // Arrange
            groupCR := &usernautdevv1alpha1.Group{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "test-group",
                    Namespace: "usernaut",
                },
                Spec: usernautdevv1alpha1.GroupSpec{
                    GroupName: "data-science-team",
                    Members: usernautdevv1alpha1.GroupMembers{
                        Users: []string{"user1", "user2"},
                    },
                    Backends: []usernautdevv1alpha1.Backend{
                        {Name: "fivetran", Type: "fivetran"},
                    },
                },
            }
            Expect(k8sClient.Create(ctx, groupCR)).To(Succeed())

            // Act
            _, err := reconciler.Reconcile(ctx, reconcile.Request{
                NamespacedName: types.NamespacedName{
                    Name:      "test-group",
                    Namespace: "usernaut",
                },
            })

            // Assert
            Expect(err).NotTo(HaveOccurred())

            // Verify status was updated
            Expect(k8sClient.Get(ctx, types.NamespacedName{
                Name: "test-group", Namespace: "usernaut",
            }, groupCR)).To(Succeed())
            Expect(groupCR.Status.BackendsStatus).To(HaveLen(1))
            Expect(groupCR.Status.BackendsStatus[0].Status).To(BeTrue())
        })
    })
})
```

#### Mock Generation

```bash
# Generate all mocks
make mockgen

# Mocks are generated for:
# - LDAP client: pkg/clients/ldap/mocks/ldap_mock.go
# - LDAP connection: internal/controller/mocks/ldap_mock.go
# - Backend clients: internal/controller/periodicjobs/mocks/client_mock.go
```

#### Testing Best Practices

- **Test isolation**: Each test should be independent
- **Use test fixtures**: Create helper functions for common test data
- **Test error paths**: Don't just test happy paths
- **Test edge cases**: Empty inputs, nil values, boundary conditions
- **Mock external dependencies**: Never make real API calls in unit tests
- **Use envtest for controller tests**: Tests run against real Kubernetes API
- **Run tests before committing**: Pre-commit hook enforces `make lint test`

### Logging

- **Use the project's custom logger package** (`pkg/logger`) which provides context-aware logging
- Always use structured logging with logrus
- Include relevant context fields in log entries
- Use appropriate log levels (Debug, Info, Warn, Error)
- Log errors with sufficient context but avoid logging the same error multiple times
- **Request ID Tracking**: Use `logger.WithRequestId()` to add request IDs to context for tracing
- **Debug Mode**: Controlled via `DEBUG_MODE` environment variable (default: Info level)

```go
// Correct: Use context-aware logger from the project's logger package
log := logger.Logger(ctx).WithFields(logrus.Fields{
    "userID":    userID,
    "component": "user-service",
    "backend":   backendName,
})
log.Info("processing user")

// Add request ID to context for tracing
ctx = logger.WithRequestId(ctx, requestID)

// Add fields to existing context logger
ctx = logger.AddValueToContextLogger(ctx, "operation", "sync")
```

#### Logging Best Practices

- **Use consistent field names** across the codebase:
  - `request_id`: Unique identifier for request tracking
  - `backend`: Backend name (e.g., "fivetran", "gitlab")
  - `backend_type`: Backend type
  - `user`: User email or identifier
  - `group`: Group name
  - `team_id`: Team identifier in backend
  - `component`: Component name (e.g., "preloadCache", "reconciliation")
- **Log at appropriate levels**:
  - `Debug`: Detailed information for development/debugging
  - `Info`: General operational messages (user created, team synced)
  - `Warn`: Recoverable issues (cache miss, retrying operation)
  - `Error`: Unrecoverable errors requiring attention
- **Include error context**: Always use `WithError(err)` when logging errors
- **Avoid sensitive data**: Never log passwords, API keys, tokens, or PII

### Package Organization

- Keep packages focused and cohesive
- Use `internal/` packages for implementation details not meant to be imported by external projects
- Use `pkg/` packages for reusable components
- Group related functionality together
- Avoid circular dependencies
- Use clear, descriptive package names

#### Project Package Structure

```
pkg/                    # Reusable packages
├── cache/             # Cache interface and implementations (Redis, in-memory)
├── clients/           # Backend client implementations
│   ├── client.go      # Client interface and factory
│   ├── fivetran/      # Fivetran client
│   ├── gitlab/        # GitLab client
│   ├── ldap/          # LDAP client
│   ├── redhat_rover/  # Red Hat Rover client
│   └── snowflake/     # Snowflake client
├── common/            # Common types and constants
├── config/            # Configuration management
├── logger/            # Custom logger with context support
├── metrics/           # Metrics and observability
├── request/           # HTTP client with circuit breaker
├── store/             # Cache store layer with prefixed keys
└── utils/             # Utility functions

internal/              # Internal packages
├── controller/        # Kubernetes controllers
│   ├── controllerutils/  # Controller helper functions
│   ├── periodicjobs/     # Periodic task implementations
│   └── mocks/            # Generated mocks for testing
└── httpapi/           # HTTP API server
    ├── handlers/      # API handlers
    ├── middleware/    # HTTP middleware
    └── server/        # API server setup

api/v1alpha1/          # Kubernetes API types
cmd/                   # Main application entry point
```

### Performance Considerations

- Use context for cancellation and timeouts
- Implement proper caching strategies (as done with Redis in this project)
- Avoid unnecessary allocations in hot paths
- Use buffered channels appropriately
- Consider using sync.Pool for frequently allocated objects
- **Use errgroup for concurrent operations** with proper error handling (see `preloadCache` in main.go)
- **Capture loop variables** when launching goroutines (Go 1.22+ handles this automatically, but be explicit for clarity)

### Concurrency and Thread Safety

This project uses a **shared cache mutex** to prevent race conditions during cache operations. Understanding and correctly using this mutex is critical.

#### Shared Cache Mutex Pattern

```go
// In main.go - create shared mutex
sharedCacheMutex := &sync.RWMutex{}

// Pass to controllers and jobs
groupReconciler := &controller.GroupReconciler{
    CacheMutex: sharedCacheMutex,
}

periodicTasksReconciler := controller.NewPeriodicTasksReconciler(
    client, sharedCacheMutex, cache, store, ldapClient, backendClients,
)
```

#### When to Lock

**ALWAYS lock** when performing these operations:

- Reading and modifying cache entries (read-modify-write operations)
- Updating user:groups reverse index
- Updating group members
- Updating user_list metadata
- Any operation that depends on consistent cache state across multiple keys

**Lock at the highest level necessary**:

```go
// Good: Lock for entire reconciliation to ensure consistency
r.CacheMutex.Lock()
defer r.CacheMutex.Unlock()

ldapResult := r.fetchLDAPData(ctx, uniqueMembers)
backendErrors := r.processAllBackends(ctx, groupCR, uniqueMembers)

if !hasErrors {
    r.updateCacheIndexes(ctx, groupCR.Spec.GroupName, ldapResult)
}
```

**Document lock assumptions in functions**:

```go
// fetchLDAPData fetches LDAP data for all unique members
// NOTE: This function assumes CacheMutex is already held by the caller
func (r *GroupReconciler) fetchLDAPData(
    ctx context.Context,
    uniqueMembers []string,
) *LDAPFetchResult {
    // Implementation...
}
```

#### Lock Granularity

For **preload operations** with concurrent backend processing:

```go
// Lock per-operation for fine-grained concurrency
g.Go(func() error {
    users, _, err := backendClient.FetchAllUsers(ctx)
    if err != nil {
        return err
    }

    for _, user := range users {
        // Lock for each individual write
        cacheMutex.Lock()
        err := dataStore.User.SetBackend(ctx, user.GetEmail(), backendKey, user.ID)
        cacheMutex.Unlock()

        if err != nil {
            return err
        }
    }
    return nil
})
```

For **reconciliation operations**:

```go
// Lock for entire reconciliation to maintain consistency
r.CacheMutex.Lock()
defer r.CacheMutex.Unlock()

// All cache operations protected by single lock
// This ensures all-or-nothing semantics for cache updates
```

#### RWMutex vs Mutex

- This project uses `sync.RWMutex` but currently **only uses exclusive locks** (`.Lock()`)
- **Future enhancement**: Use `.RLock()` for read-only operations (e.g., checking if user exists)
- **Be careful**: Read-modify-write operations MUST use exclusive locks, not read locks

#### Common Concurrency Pitfalls

❌ **Don't**: Lock inside loops when processing multiple items

```go
// BAD: Excessive lock contention
for _, user := range users {
    r.CacheMutex.Lock()
    // Do lots of work...
    r.CacheMutex.Unlock()
}
```

✅ **Do**: Lock once for batch operations when consistency is required

```go
// GOOD: Single lock for consistent state
r.CacheMutex.Lock()
defer r.CacheMutex.Unlock()
for _, user := range users {
    // Process all users...
}
```

✅ **Do**: Fine-grained locks for independent operations (e.g., preload)

```go
// GOOD: Fine-grained locks for concurrent processing
for _, user := range users {
    r.CacheMutex.Lock()
    err := dataStore.User.SetBackend(ctx, email, key, id)
    r.CacheMutex.Unlock()
}
```

❌ **Don't**: Forget to document lock assumptions

```go
// BAD: No indication that caller must hold lock
func (s *GroupStore) Get(ctx context.Context, groupName string) (*GroupData, error) {
    // ...
}
```

✅ **Do**: Document lock requirements clearly

```go
// GOOD: Clear documentation
// Get retrieves the full group data from cache
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *GroupStore) Get(ctx context.Context, groupName string) (*GroupData, error) {
    // ...
}
```

### Kubernetes Operator Specific Guidelines

- Follow controller-runtime patterns for reconciliation
- Use proper status updates with conditions
- Implement proper finalizers for cleanup
- Handle resource creation idempotently
- Use client-go best practices for API interactions
- **Use predicate filters** for efficient event handling (e.g., `GenerationChangedPredicate` to avoid unnecessary reconciles)
- **Implement custom predicates** when needed (see `ForceReconcilePredicate` in controllerutils)

```go
// Good reconciler pattern
func (r *GroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Always add request ID to context for tracing
    ctx = logger.WithRequestId(ctx, controller.ReconcileIDFromContext(ctx))
    log := logger.Logger(ctx).WithValues("group", req.NamespacedName)

    var group usernautdevv1alpha1.Group
    if err := r.Get(ctx, req.NamespacedName, &group); err != nil {
        // IgnoreNotFound returns nil on NotFound errors
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Handle deletion with finalizers
    if group.GetDeletionTimestamp() != nil {
        return ctrl.Result{}, r.handleDeletion(ctx, &group)
    }

    // Add finalizer if missing
    if !controllerutil.ContainsFinalizer(&group, groupFinalizer) {
        controllerutil.AddFinalizer(&group, groupFinalizer)
        if err := r.Update(ctx, &group); err != nil {
            return ctrl.Result{}, err
        }
    }

    // Reconciliation logic here

    return ctrl.Result{}, nil
}
```

#### Finalizer Best Practices

- **Always use constants for finalizer names** (e.g., `const groupFinalizer = "operator.dataverse.redhat.com/finalizer"`)
- **Check for deletion timestamp** before performing reconciliation
- **Cleanup must be idempotent** - handle cases where resources might already be deleted
- **Remove finalizer only after successful cleanup**
- **Log cleanup operations** for debugging

#### Status Updates

- **Update status separately from spec** using `r.Status().Update(ctx, obj)`
- **Use custom status conditions** to indicate different states
- **Set waiting state** at the start of reconciliation
- **Update status with backend-specific results** (see `BackendStatus` in Group CR)
- **Handle partial failures** - update status even when some backends fail

```go
// Set status before starting work
groupCR.SetWaiting()
if err := r.Status().Update(ctx, groupCR); err != nil {
    return ctrl.Result{}, err
}

// Update status with results
groupCR.Status.BackendsStatus = backendStatus
groupCR.UpdateStatus(hasErrors)
if err := r.Status().Update(ctx, groupCR); err != nil {
    return ctrl.Result{}, err
}
```

#### Controller Setup

- **Use field indexers** for efficient queries (see Group controller's setup for referenced groups)
- **Implement watches for related resources** using `EnqueueRequestsFromMapFunc`
- **Use event filters** to reduce unnecessary reconciliations
- **Set MaxConcurrentReconciles** appropriately based on workload

```go
func (r *GroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
    // Add index for efficient lookups
    indexField := "spec.members.groups"
    if err := mgr.GetFieldIndexer().IndexField(context.Background(),
        &usernautdevv1alpha1.Group{}, indexField,
        func(obj client.Object) []string {
            group := obj.(*usernautdevv1alpha1.Group)
            return group.Spec.Members.Groups
        }); err != nil {
        return err
    }

    // Setup controller with predicates
    labelPredicate := controllerutils.ForceReconcilePredicate()
    return ctrl.NewControllerManagedBy(mgr).
        For(&usernautdevv1alpha1.Group{}).
        WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, labelPredicate)).
        Watches(
            &usernautdevv1alpha1.Group{},
            handler.EnqueueRequestsFromMapFunc(mapFunc),
        ).
        Complete(r)
}
```

### Caching Patterns and Store Layer

This project uses a **two-tier caching architecture**:

1. **Cache Layer** (`pkg/cache`): Interface and implementations (Redis, in-memory)
2. **Store Layer** (`pkg/store`): Encapsulates cache operations with prefixed keys and business logic

#### Cache Key Naming Convention

**Always use prefixed keys** to organize cache data:

```
user:<email>                     # User backend mappings
team:<transformed_name>          # Team backend mappings (legacy, from preload)
group:<group_name>               # Consolidated group data (members + backends)
user:groups:<email>              # Reverse index: user -> groups
meta:user_list                   # List of all active users
```

#### Store Layer Architecture

The store layer provides **encapsulated cache operations** with consistent key prefixes:

```go
// Store provides access to all cache stores
type Store struct {
    User       UserStoreInterface
    Team       TeamStoreInterface      // Legacy, used for preload fallback
    Group      GroupStoreInterface     // Primary group store
    Meta       MetaStoreInterface
    UserGroups UserGroupsStoreInterface
}

// Create store instance
dataStore := store.New(cacheClient)

// Use store operations
err := dataStore.User.SetBackend(ctx, email, backendKey, userID)
backends, err := dataStore.Group.GetBackends(ctx, groupName)
groups, err := dataStore.UserGroups.GetGroups(ctx, email)
```

#### Cache Data Structures

**User Store** (`user:<email>`):

```json
{
  "fivetran_fivetran": "user_id_123",
  "gitlab_gitlab": "42",
  "rover_rover": "jdoe"
}
```

**Team Store** (`team:<transformed_name>`) - Legacy, populated by preload:

```json
{
  "fivetran_fivetran": "team_id_456",
  "gitlab_gitlab": "789"
}
```

**Group Store** (`group:<group_name>`) - Primary, populated by reconciliation:

```json
{
  "members": ["user1@example.com", "user2@example.com"],
  "backends": {
    "fivetran_fivetran": {
      "id": "team_id_456",
      "name": "fivetran",
      "type": "fivetran"
    },
    "gitlab_gitlab": {
      "id": "789",
      "name": "gitlab",
      "type": "gitlab"
    }
  }
}
```

**User Groups Index** (`user:groups:<email>`):

```json
["data-science-team", "ml-engineers", "platform-users"]
```

**Meta User List** (`meta:user_list`):

```json
["user1", "user2", "user3", "user4"]
```

#### Backend Key Format

Backend keys are **composite keys** combining name and type:

```
<backend_name>_<backend_type>

Examples:
- fivetran_fivetran
- gitlab_gitlab
- rhplatformtest_snowflake
- rover_rover
```

**Benefits**:

- Supports multiple instances of the same backend type
- Clear identification of both instance and type
- Easy parsing and validation

#### Cache Store Pattern

When implementing new store operations:

1. **Create interface** in `pkg/store/interfaces.go`
2. **Implement store** in `pkg/store/<name>_store.go`
3. **Document lock requirements** in all functions
4. **Handle empty results gracefully** - return empty slices/maps, not nil

```go
// Good store implementation
type UserStore struct {
    cache cache.Cache
}

// userKey returns the prefixed cache key
func (s *UserStore) userKey(email string) string {
    return "user:" + email
}

// GetBackends returns backend mappings for a user
// NOTE: Caller must hold appropriate lock if concurrent access is possible
func (s *UserStore) GetBackends(ctx context.Context, email string) (map[string]string, error) {
    key := s.userKey(email)
    val, err := s.cache.Get(ctx, key)
    if err != nil {
        // Return empty map on cache miss
        return make(map[string]string), nil
    }

    var backends map[string]string
    if err := json.Unmarshal([]byte(val.(string)), &backends); err != nil {
        return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
    }

    return backends, nil
}
```

#### Cache Consistency Patterns

**All-or-Nothing Updates**:

```go
// Fetch LDAP data and process backends
ldapResult := r.fetchLDAPData(ctx, uniqueMembers)
backendErrors := r.processAllBackends(ctx, groupCR, uniqueMembers)

// Only update cache indexes if ALL backends succeeded
if !hasErrors {
    r.updateCacheIndexes(ctx, groupCR.Spec.GroupName, ldapResult)
} else {
    r.log.Warn("Backend errors detected, skipping cache index updates")
}
```

**Two-Store Fallback Pattern** (Group/Team stores):

```go
// 1. Check primary store (GroupStore with original name)
teamID, err := r.Store.Group.GetBackendID(ctx, groupName, backendName, backendType)
if teamID != "" {
    return teamID, nil
}

// 2. Fallback to secondary store (TeamStore with transformed name)
teamBackends, err := r.Store.Team.GetBackends(ctx, transformedGroupName)
if id, exists := teamBackends[backendKey]; exists {
    // Migrate to primary store
    r.Store.Group.SetBackend(ctx, groupName, backendName, backendType, id)
    return id, nil
}

// 3. Create new resource
newTeam, err := backendClient.CreateTeam(ctx, team)
```

#### Cache Expiration Policy

**Current Strategy**: No expiration for all keys

**Rationale**:

- User and team data is explicitly managed by Kubernetes CRs
- Cache invalidation is event-driven (Group CR updated/deleted)
- Prevents stale data from timing out mid-operation
- Reduces unnecessary backend API calls

**Cache Invalidation**:

- **On Group CR deletion**: Remove group entry, update user:groups indexes
- **On Group CR update**: Update group members and user:groups indexes
- **On user removal**: Remove from group members, update user:groups index
- **Never**: Set TTL on cache keys (explicit management only)

### Security Best Practices

- Never log sensitive information (passwords, API keys, tokens)
- Validate all inputs
- Use proper authentication and authorization
- Handle secrets securely using Kubernetes Secrets
- Implement proper RBAC for Kubernetes resources
- **Configuration Security**:
  - Store sensitive data in Kubernetes Secrets (mounted as volumes or env vars)
  - Use ConfigMaps for non-sensitive configuration
  - Never commit secrets to git (enforced by `scripts/check-secrets.sh`)
- **API Authentication**: HTTP API supports basic auth (configurable in `apiServer.auth`)

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

### HTTP Client with Circuit Breaker

This project uses a **custom HTTP client with Hystrix circuit breaker** for resilient backend communication.

#### HTTP Client Configuration

```yaml
httpClient:
  connectionPoolConfig:
    maxIdleConns: 100
    maxConnsPerHost: 100
    idleConnTimeout: 90 # seconds
  hystrixResiliencyConfig:
    timeout: 30000 # milliseconds
    maxConcurrentRequests: 100
    errorPercentThreshold: 50
    sleepWindow: 5000 # milliseconds
    requestVolumeThreshold: 20
```

#### Using the HTTP Client

```go
import "github.com/redhat-data-and-ai/usernaut/pkg/request/httpclient"

// Create HTTP client with circuit breaker
client := httpclient.NewHTTPClient(
    connectionPoolConfig,
    hystrixResiliencyConfig,
)

// Make request with circuit breaker protection
resp, err := client.Do(ctx, "backend-service", func() (*http.Response, error) {
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    return httpClient.Do(req)
})
```

#### Circuit Breaker Benefits

- **Automatic failure detection**: Opens circuit after threshold errors
- **Fast failure**: Fails immediately when circuit is open
- **Self-healing**: Closes circuit after sleep window
- **Request limiting**: Prevents overwhelming downstream services
- **Timeout protection**: Enforces request timeouts

### HTTP API Server

The project includes an **HTTP API server** for external access to cache data.

#### API Endpoints

```
GET  /health                      # Health check
GET  /api/v1/users/:email         # Get user backends
GET  /api/v1/groups/:groupname    # Get group details
GET  /api/v1/users/:email/groups  # Get user's groups
```

#### API Server Configuration

```yaml
apiServer:
  address: ":8080"
  auth:
    enabled: true
    basic_users:
      - username: admin
        password: "${API_ADMIN_PASSWORD}"
  cors:
    allowed_origins:
      - "http://localhost:3000"
      - "https://app.example.com"
```

#### Implementing New API Endpoints

1. **Add handler** in `internal/httpapi/handlers/handlers.go`:

```go
func (h *Handler) GetUserGroups(c *gin.Context) {
    email := c.Param("email")

    groups, err := h.store.UserGroups.GetGroups(c.Request.Context(), email)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"email": email, "groups": groups})
}
```

2. **Register route** in `internal/httpapi/server/server.go`:

```go
v1 := r.Group("/api/v1")
{
    v1.GET("/users/:email/groups", handler.GetUserGroups)
}
```

3. **Apply middleware**:
   - Basic auth: `middleware.BasicAuth(authConfig)`
   - CORS: `middleware.CORS(corsConfig)`

### Periodic Jobs

The project supports **periodic background jobs** that run independently of Kubernetes reconciliation.

#### Periodic Task Architecture

```go
// Task manager coordinates periodic jobs
type PeriodicTaskManager struct {
    tasks []PeriodicTask
}

// Individual task interface
type PeriodicTask interface {
    Name() string
    Schedule() string  // Cron expression
    Run(ctx context.Context) error
}
```

#### Implementing a Periodic Job

1. **Create job struct** in `internal/controller/periodicjobs/`:

```go
type MyJob struct {
    cacheMutex     *sync.RWMutex
    store          *store.Store
    backendClients map[string]clients.Client
}

func (j *MyJob) Name() string {
    return "my-job"
}

func (j *MyJob) Schedule() string {
    return "0 2 * * *"  // Daily at 2 AM
}

func (j *MyJob) Run(ctx context.Context) error {
    log := logger.Logger(ctx).WithField("job", j.Name())
    log.Info("Starting job execution")

    // Acquire lock if accessing cache
    j.cacheMutex.Lock()
    defer j.cacheMutex.Unlock()

    // Job logic here

    return nil
}
```

2. **Register job** in `cmd/main.go`:

```go
periodicTaskManager := periodicjobs.NewPeriodicTaskManager()

myJob := periodicjobs.NewMyJob(
    sharedCacheMutex,
    dataStore,
    backendClients,
)
myJob.AddToPeriodicTaskManager(periodicTaskManager)
```

#### User Offboarding Job Example

The project includes a **user offboarding job** that runs daily:

- Fetches all users from cache
- Checks LDAP status for each user
- Removes inactive users from all backends
- Updates cache entries

**Key features**:

- Uses shared cache mutex for consistency
- Processes users in batches
- Handles partial failures gracefully
- Logs detailed progress for auditing

### Make Targets

Use the provided Makefile targets:

- `make test` - Run tests with coverage (requires `mockgen`, `manifests`, `generate`, `fmt`, `vet`, `envtest`)
- `make lint` - Run golangci-lint linter
- `make lint-fix` - Run linter and auto-fix issues
- `make build` - Build the binary (requires `manifests`, `generate`, `fmt`, `vet`)
- `make fmt` - Format code with `go fmt`
- `make vet` - Run `go vet`
- `make mockgen` - Generate mocks for testing
- `make run` - Run controller from host (requires `setup-pre-commit`, `manifests`, `generate`, `fmt`, `vet`)
- `make manifests` - Generate CRDs and RBAC
- `make generate` - Generate DeepCopy methods
- `make docker-build` - Build container image with buildah
- `make deploy` - Deploy to Kubernetes (requires `USERNAUT_ENV` env var)

#### Pre-commit Hook

The project uses a **pre-commit hook** that automatically runs linting and tests:

```bash
# Install pre-commit hook (done automatically on make install/run)
make setup-pre-commit

# The hook runs: make lint test
# Commit will fail if linting or tests fail
```

**Bypassing the hook** (not recommended):

```bash
git commit --no-verify
```

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

#### Code Quality

- [ ] All errors are properly handled (no ignored errors)
- [ ] Error messages include sufficient context
- [ ] Code is properly formatted (`gofmt`/`goimports` applied)
- [ ] No linter warnings (`make lint` passes)
- [ ] Line length is reasonable (exceptions allowed for `api/*` paths)
- [ ] Code follows existing patterns and project conventions
- [ ] Functions are focused and do one thing well
- [ ] Variable names are descriptive and follow camelCase/PascalCase conventions

#### Testing

- [ ] Unit tests are included for new functionality
- [ ] Tests use table-driven approach where appropriate
- [ ] Tests cover error paths and edge cases
- [ ] Mocks are generated/updated (`make mockgen` run)
- [ ] Tests pass locally (`make test` succeeds)
- [ ] Test names clearly describe the scenario being tested

#### Logging and Observability

- [ ] Structured logging is used with appropriate context fields
- [ ] Log levels are appropriate (Debug/Info/Warn/Error)
- [ ] Sensitive information is NOT logged (passwords, tokens, API keys)
- [ ] Request IDs are added to context for tracing
- [ ] Error logging includes `.WithError(err)` for stack traces
- [ ] Consistent field names used (`backend`, `user`, `group`, `component`, etc.)

#### Context and Interfaces

- [ ] Context is passed as first parameter to functions
- [ ] Context is properly propagated through call chains
- [ ] Interfaces are used for external dependencies
- [ ] Interfaces are defined where used, not where implemented
- [ ] Context cancellation is respected (select on `ctx.Done()` in loops)

#### Documentation

- [ ] Public functions have godoc comments
- [ ] Complex logic is explained with comments
- [ ] Lock requirements are documented (`NOTE: Caller must hold lock`)
- [ ] Package-level documentation exists for new packages
- [ ] README or other docs updated if behavior changes
- [ ] Apache 2.0 license header included in new files

### Kubernetes Operator Specific Review

#### Controller Implementation

- [ ] Reconciler follows standard pattern (Get, handle deletion, add finalizer, reconcile, update status)
- [ ] Finalizers are properly implemented for cleanup
- [ ] `client.IgnoreNotFound(err)` used for Get operations
- [ ] Status updates use `r.Status().Update()` not `r.Update()`
- [ ] Predicates are used to filter unnecessary reconciliations
- [ ] Field indexers are used for efficient queries
- [ ] Request ID is added to context in Reconcile function

#### Status and Conditions

- [ ] Status is updated separately from spec
- [ ] Custom status conditions are used appropriately
- [ ] Status reflects current state accurately
- [ ] Waiting state is set at start of reconciliation
- [ ] Backend-specific status is updated (BackendStatus slice)
- [ ] Partial failures are handled (some backends succeed, some fail)

#### Finalizers and Cleanup

- [ ] Finalizer name is a constant
- [ ] Deletion timestamp is checked before reconciliation
- [ ] Cleanup is idempotent (can be called multiple times)
- [ ] Finalizer is removed only after successful cleanup
- [ ] Cache entries are properly cleaned up (group, user:groups index)

#### Resource Management

- [ ] Resources are created idempotently
- [ ] Owner references are set correctly
- [ ] Watches are configured for related resources
- [ ] Controller options are appropriate (MaxConcurrentReconciles, namespace scope)

### Concurrency and Thread Safety Review

#### Mutex Usage

- [ ] Shared cache mutex is used for cache operations
- [ ] Lock is acquired at appropriate granularity (coarse for consistency, fine for concurrency)
- [ ] `defer unlock()` is used immediately after lock
- [ ] Lock assumptions are documented in function comments
- [ ] No locks are held during blocking operations (HTTP calls, LDAP queries) unless necessary
- [ ] RWMutex is used correctly (currently only exclusive locks, but consider RLock for read-only)

#### Concurrent Processing

- [ ] Loop variables are captured when launching goroutines (or use Go 1.22+ for-loop semantics)
- [ ] errgroup is used for concurrent operations with error handling
- [ ] Context cancellation is checked in loops
- [ ] Race conditions are prevented (verified with `go test -race`)

### Caching and Store Layer Review

#### Store Usage

- [ ] Store layer is used instead of direct cache access
- [ ] Appropriate store is used (User/Team/Group/Meta/UserGroups)
- [ ] Cache keys follow naming convention (`user:`, `team:`, `group:`, `user:groups:`, `meta:`)
- [ ] Backend keys use format `<name>_<type>`
- [ ] Empty results return empty slices/maps, not nil
- [ ] JSON marshaling/unmarshaling errors are handled

#### Cache Consistency

- [ ] All-or-nothing semantics maintained (cache indexes updated only on full success)
- [ ] User:groups reverse index is updated correctly
- [ ] Group members are updated atomically
- [ ] Two-store fallback pattern used correctly (Group -> Team migration)
- [ ] No TTL/expiration is set on cache keys (explicit management only)

#### Cache Operations

- [ ] Get operations handle cache misses gracefully
- [ ] Set operations use `cache.NoExpiration`
- [ ] Delete operations clean up all related entries (group, indexes)
- [ ] Pattern-based queries use correct prefix

### Configuration Review

#### Configuration Structure

- [ ] Configuration follows YAML structure in `appconfig/`
- [ ] Environment-specific overrides are used correctly
- [ ] Sensitive values use environment variable substitution (`${VAR_NAME}`)
- [ ] Required configuration is validated on startup
- [ ] Backend dependencies are specified correctly (`depends_on`)

#### Configuration Access

- [ ] Configuration is loaded via `config.GetConfig()` or `config.LoadConfig(env)`
- [ ] Backend configuration accessed via `BackendMap`
- [ ] Connection parameters use `GetStringConnection()` with defaults
- [ ] Missing required parameters cause early failure

### New Client Implementation Review

- [ ] Uses established, well-maintained library (e.g., `gitlab.com/gitlab-org/api/client-go`)
- [ ] Library choice is justified and documented
- [ ] Implements complete `clients.Client` interface
- [ ] Follows established directory structure (`pkg/clients/<name>/`)
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
- [ ] `go mod vendor` executed and `vendor/` directory updated

### HTTP API Review

- [ ] Endpoints follow RESTful conventions
- [ ] Handlers use proper HTTP status codes
- [ ] Request parameters are validated
- [ ] Context from request is propagated
- [ ] Store layer is used for data access
- [ ] Error responses are consistent
- [ ] Authentication middleware applied where needed
- [ ] CORS middleware configured correctly

### Periodic Jobs Review

- [ ] Job implements `PeriodicTask` interface
- [ ] Cron schedule is appropriate for job frequency
- [ ] Shared cache mutex is used when accessing cache
- [ ] Job is registered in `cmd/main.go`
- [ ] Job handles context cancellation
- [ ] Batch processing is used for large datasets
- [ ] Partial failures are handled gracefully
- [ ] Detailed logging for auditing

## Summary and Quick Reference

### Common Patterns

**Error Handling**:

```go
if err != nil {
    return fmt.Errorf("failed to fetch user: %w", err)
}
```

**Context-Aware Logging**:

```go
log := logger.Logger(ctx).WithFields(logrus.Fields{
    "user": email,
    "backend": backendName,
})
log.Info("processing user")
```

**Cache Lock Pattern**:

```go
r.CacheMutex.Lock()
defer r.CacheMutex.Unlock()
// All cache operations here
```

**Store Operations**:

```go
backends, err := dataStore.User.GetBackends(ctx, email)
err = dataStore.Group.SetBackend(ctx, groupName, backendName, backendType, teamID)
groups, err := dataStore.UserGroups.GetGroups(ctx, email)
```

**Table-Driven Tests**:

```go
tests := []struct {
    name    string
    input   interface{}
    want    interface{}
    wantErr bool
}{
    {name: "valid input", input: x, want: y},
    {name: "error case", input: z, wantErr: true},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { /* ... */ })
}
```

### Key Files and Their Purposes

- `cmd/main.go` - Application entry point, controller setup, cache preload
- `internal/controller/group_controller.go` - Main reconciliation logic
- `internal/controller/periodic_tasks_controller.go` - Periodic job orchestration
- `pkg/store/` - Cache store layer with prefixed keys
- `pkg/cache/` - Cache interface and implementations
- `pkg/clients/` - Backend client implementations
- `pkg/config/` - Configuration management
- `pkg/logger/` - Context-aware logging
- `api/v1alpha1/` - Kubernetes CRD types

### Anti-Patterns to Avoid

❌ **Don't** log sensitive information:

```go
log.Infof("Using token: %s", apiToken)  // BAD
```

❌ **Don't** ignore errors:

```go
client.CreateUser(ctx, user)  // BAD: error ignored
```

❌ **Don't** use direct cache access:

```go
cache.Set(ctx, "user:"+email, data, 0)  // BAD: use store layer
```

❌ **Don't** forget to lock when accessing cache:

```go
dataStore.User.GetBackends(ctx, email)  // BAD if others might be writing
```

❌ **Don't** set TTL on cache keys:

```go
cache.Set(ctx, key, value, 5*time.Minute)  // BAD: use NoExpiration
```

❌ **Don't** store context in structs:

```go
type Service struct {
    ctx context.Context  // BAD: pass context as parameter
}
```

### Development Workflow

1. **Start development**: `make run` (installs pre-commit hook)
2. **Make changes**: Edit code, follow patterns
3. **Format code**: `make fmt`
4. **Run linter**: `make lint` (or `make lint-fix` for auto-fixes)
5. **Generate mocks**: `make mockgen` (if interfaces changed)
6. **Run tests**: `make test`
7. **Commit**: Git pre-commit hook runs `make lint test` automatically
8. **Build**: `make build`
9. **Deploy**: `make deploy USERNAUT_ENV=<env>`

### Troubleshooting

**Linter errors**: Run `make lint-fix` to auto-fix, or check `.golangci.yml` for configuration
**Test failures**: Check mock generation with `make mockgen`, verify envtest setup
**Cache issues**: Check Redis connection, verify store layer usage, check lock acquisition
**Controller not reconciling**: Check predicates, verify namespace scope, check leader election
**Backend errors**: Check circuit breaker status, verify HTTP client configuration, check API credentials

### Resources

- **Project Documentation**: `README.md`, `DEVELOPMENT.md`, `Deployment.md`
- **Kubernetes Docs**: https://kubernetes.io/docs/
- **controller-runtime**: https://pkg.go.dev/sigs.k8s.io/controller-runtime
- **Operator Pattern**: https://kubernetes.io/docs/concepts/extend-kubernetes/operator/
- **Ginkgo Testing**: https://onsi.github.io/ginkgo/

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
