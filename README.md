# Usernaut

Usernaut, a Kubernetes Operator based Application that manages user accounts and teams across multiple SaaS platforms, providing a unified, declarative approach to user lifecycle management.

## Why Usernaut?

Managing user access across multiple SaaS platforms is challenging:

- **Manual Provisioning is Slow**: Onboarding users to Fivetran, GitLab, Snowflake, and other platforms requires repetitive manual work
- **Offboarding is Error-Prone**: When employees leave, their access often lingers across systems, creating security risks
- **Inconsistent Access Control**: Without a single source of truth, team memberships drift out of sync across platforms
- **Compliance Burden**: Auditing who has access to what requires checking multiple systems

**Usernaut solves these problems** by using Kubernetes Custom Resources to declaratively define user groups and automatically synchronizing them to all your SaaS backends.

## Key Features

- **Declarative User Management**: Define groups as Kubernetes CRs; Usernaut handles the rest
- **Multi-Backend Sync**: Simultaneously manage users across Fivetran, GitLab, Snowflake, and Rover
- **LDAP Integration**: Automatically fetches user details from your corporate directory
- **Nested Groups**: Groups can include other groups for flexible team structures
- **Automatic Offboarding**: Daily job removes users no longer in LDAP from all backends
- **REST API**: Query user group memberships programmatically

## Supported Backends

| Backend           | Type        | Description                                     |
| ----------------- | ----------- | ----------------------------------------------- |
| **Fivetran**      | `fivetran`  | Data pipeline platform                          |
| **GitLab**        | `gitlab`    | Git repository hosting (with LDAP sync support) |
| **Snowflake**     | `snowflake` | Cloud data warehouse                            |
| **Red Hat Rover** | `rover`     | Red Hat internal user directory                 |

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Kubernetes Cluster                          │
│                                                                     │
│  ┌──────────────┐      ┌─────────────────┐      ┌───────────────┐  │
│  │  Group CRD   │─────▶│  Usernaut       │─────▶│   Backends    │  │
│  │              │      │  Controller     │      │               │  │
│  │ - group_name │      │                 │      │ • Fivetran    │  │
│  │ - members    │      │ • Sync users    │      │ • GitLab      │  │
│  │ - backends   │      │ • Manage teams  │      │ • Snowflake   │  │
│  └──────────────┘      │ • Auto-offboard │      │ • Rover       │  │
│                        └────────┬────────┘      └───────────────┘  │
│                                 │                                   │
│                                 ▼                                   │
│                        ┌─────────────────┐      ┌───────────────┐  │
│                        │  Cache Layer    │      │  LDAP Server  │  │
│                        │  (Redis)        │      │  (User Data)  │  │
│                        └─────────────────┘      └───────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

## Documentation

| Document                             | Description                                                  |
| ------------------------------------ | ------------------------------------------------------------ |
| [DEVELOPMENT.md](./DEVELOPMENT.md)   | Architecture deep-dive, core components, and developer setup |
| [CONTRIBUTING.md](./CONTRIBUTING.md) | How to contribute to Usernaut                                |
| [Deployment.md](./Deployment.md)     | Detailed deployment instructions                             |

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](./LICENSE) file for details.
