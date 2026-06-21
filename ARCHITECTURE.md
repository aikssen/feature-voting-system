ARCHITECTURE.md

Architecture Overview

The application follows a pragmatic architecture focused on:

* Simplicity
* Maintainability
* Fast delivery
* AI-assisted development
* Production-minded design

The goal is to demonstrate engineering decisions without introducing unnecessary complexity.

⸻
```
High-Level Architecture

Frontend (SPA)
        │
        ▼
REST API
        │
        ▼
Application Services
        │
        ▼
Repositories
        │
        ▼
PostgreSQL
```

The system is composed of:

* Responsive Frontend
* Go REST API
* PostgreSQL Database

⸻
```
Monorepo Structure

/
├── frontend/
├── backend/
├── docker-compose.yml
├── PROJECT.md
├── ARCHITECTURE.md
├── DESIGN.md
├── TASKS.md
├── CLAUDE.md
├── prompts.txt
└── README.md
```
⸻

Backend Architecture

Architectural Style

The backend follows a Vertical Slice Architecture approach.

Business capabilities are organized by feature rather than by technical layer, allowing each feature to own its handlers, services, repositories, and domain models.



Project Structure:

```
backend/

internal/

  auth/
  feature/
  vote/

  infrastructure/
  shared/
```

Feature Structure

Each feature contains the components required to implement a complete business capability.

Example:
```
feature/

  handler.go
  service.go
  repository.go
  model.go
```

Responsibilities:

* handler.go
    * HTTP request handling
    * Request validation
    * Response mapping
* service.go
    * Business logic
    * Business rule enforcement
    * Application workflows
* repository.go
    * Data access abstraction
    * Database interactions
* model.go
    * Domain models
    * Request and response DTOs when appropriate

Persistence Strategy

Repositories are used to abstract persistence concerns from business logic.

Services depend on repository interfaces rather than concrete database implementations.

Example:
```
HTTP Handler
      │
      ▼
Application Service
      │
      ▼
Repository Interface
      │
      ▼
PostgreSQL Repository
```

This separation allows infrastructure concerns to evolve independently from business rules while keeping the architecture simple and pragmatic.


Architectural Principles

The backend prioritizes:

* Simplicity over abstraction
* Clear ownership of business capabilities
* Production-minded design
* Fast onboarding and navigation
* Testability

The following patterns are intentionally out of scope:

* CQRS
* Event Sourcing
* Domain Events
* Message Brokers
* Microservices
* Complex DDD patterns

These approaches would introduce unnecessary complexity for the scope of this project.


⸻

Technology Stack

Dependencies should be version pinned to ensure reproducible builds.


Language:

* Go
* Use port :3000 for API

Router:

* Chi

Database:

* PostgreSQL

Authentication:

* JWT

Password Hashing:

* bcrypt

Backend decisions

* Use dependency injection to decople components

⸻

Frontend Architecture

The frontend should be:

* Mobile-first
* Responsive
* SPA (Single Page Application)

Suggested Stack:

* React
* TypeScript
* Vite
* TailwindCSS

* pnpm
* Vitest
* Use por :5173 for vite


React Router is not required.
The application may operate as a single-route SPA.

Why:

* Fast development
* Excellent AI support
* Modern developer experience
* Responsive UI implementation

The frontend communicates exclusively through the REST API.

⸻

Domain Model

User

Responsibilities:

* Authentication
* Ownership of Feature Requests
* Voting

Fields:

id
name
email
password_hash
created_at

⸻

Feature Request

Responsibilities:

* Submission
* Discovery
* Prioritization

Fields:

id
title
description
author_id
created_at

⸻

Vote

Responsibilities:

* Ranking
* Popularity
* Trending

Fields:

id
feature_request_id
user_id
created_at

⸻

Business Rules

The backend must enforce:

* Authentication required to create requests
* Authentication required to vote
* One vote per user per feature
* No self-voting
* Required title
* Required description

⸻

REST API

Healtcheck

A basic healthcheck endpoint must exists.

GET /api/v1/health


Authentication

Sign Up

POST /api/v1/auth/signup

Request:

{
  "name": "Ever",
  "email": "ever@example.com",
  "password": "password"
}

⸻

Login

POST /api/v1/auth/login

Request:

{
  "email": "ever@example.com",
  "password": "password"
}

Response:

{
  "token": "jwt"
}

⸻

Logout

Frontend-only operation.

JWT removal from local storage. 

⸻

Feature Requests

Create Feature Request

POST /api/v1/features

Authentication required.

⸻

List Feature Requests

List endpoints should return aggregated voting information and author information in a single query.

GET /api/v1/features

Supports:

search=
sort=
page=
limit= (20 by default)

Sort values:

newest
most_voted
trending (default)


⸻

Get Feature Request

GET /api/v1/features/{id}

Although the UI is primarily single-page, the API should expose resource retrieval for external integrations.

⸻

Voting

Vote Feature Request

POST /api/v1/features/{id}/vote

Authentication required.

Business rules:

* Prevent self-voting

⸻

Discovery Features

Supported capabilities:

* Search
* Sorting
* Pagination

Search:

title
description

Use LIKE - %q%
Case insensitive.
Empty query show all.


Sorting:

newest
most_voted
trending (default)

⸻

Trending Strategy

The system supports:

Most Voted
Trending
Newest

Trending should combine:

* Vote count
* Recency

Trending combines popularity and recency.

The implementation should favor recently active requests while still considering total vote count.

The exact scoring formula is intentionally implementation-specific and may evolve over time.


API ERRORS:

The API should return success and error codes according to the standard industry.
200 - 201 - 400 - 400 - 401 - 404 - 500
Message errors must structured and clear in JSON format.


⸻

Database Strategy

PostgreSQL was selected because it reflects the target deployment environment more closely while maintaining a simple local development workflow through Docker.

The design should remain compatible with future migration to PostgreSQL.

Indexes should be created for:

feature_requests.created_at
votes.feature_request_id
votes.user_id
users.email

⸻

Testing Strategy

Priority should be given to:

* Service layer tests
* Business rules
* Authentication rules
* Voting rules

Critical scenarios:

* Duplicate vote rejection
* Self-vote rejection
* Authentication validation

⸻

Deployment

The application must run using Docker.

Goals:

* Single command startup
* Reproducible environment
* Reviewer-friendly setup

Expected command:

docker compose up

⸻

Architectural Non-Goals

The following are intentionally excluded:

* Microservices
* CQRS
* Event Sourcing
* Domain Events
* Message Brokers
* OAuth
* Refresh Tokens
* RBAC
* Administration Panels


Unit testing

Do not create data into the database for unit testing.
Use mocks or in-memory test doubles for unit tests.
PostgreSQL integration tests may be added separately.




The architecture should remain intentionally simple and appropriate for the scope of the assessment.

# Docker 
Dockerize the applicacions to make easier the execution.
Use docker compose to start the application.
Create a volume to persist data after a docker restart.
The database is a dependency for the backend so it must start first.

# Seeds
Seeds must be idempotent.
The database should be created from a script if it does't exists.
Add seed data to populate with son feature requests.
Execute these scripts once docker starts only if the db is empty.
