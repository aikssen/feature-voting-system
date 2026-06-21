PROJECT.md

SoundFlow Feature Voting System

Vision

SoundFlow is a fictional mobile-first music streaming platform.

This project represents the public feature voting portal used by SoundFlow users to help shape the future of the product.

Users can submit ideas, discover requests from the community, and vote on the features they believe should be prioritized by the product team.

The goal is to create a simple, modern, and production-minded platform that allows product feedback to be collected and prioritized efficiently.

⸻

Product Goals

The platform should allow users to:

1. Submit new feature requests
2. Discover existing feature requests
3. Prioritize features through voting

The platform should make it easy for product teams to understand which requests matter most to users.

⸻

Core User Journey

1. User signs up or logs in
2. User discovers existing feature requests
3. User searches or sorts requests
4. User votes on requests submitted by other users
5. User submits new feature requests
6. Popular requests rise in visibility and ranking

⸻

Core Features

Authentication

Users can:

* Sign Up
* Log In
* Log Out

Authentication should remain intentionally simple.
For simplicity and assessment scope, JWT tokens are stored client-side with 24h expiration.
Re-login required.
In production, httpOnly cookies would be preferred.

No refresh token.

A user must be able to see the homepage and its content but must be logged in to add feature requests and vote.

User creation:

User email must be unique.
Password min 4 characters, max 12 characters. At least 1 special char is needed.

⸻

Feature Requests

Users can:

* Create feature requests
* View feature requests
* Search feature requests
* Sort feature requests

Each feature request contains:

* Title
* Description
* Author
* Creation Date

⸻

Voting

Authenticated users can vote on feature requests.

The system tracks:

* Total votes
* Ranking position
* Trending score

⸻

Business Rules

Voting Rules

A user may vote only once per feature request.

A user cannot vote for their own feature request.

Votes are associated with authenticated users.

Unvote / Edit / Delete are out of scope.


⸻

Feature Request Rules

A feature request must contain:

* Title (min=2 char, max=100 char)
* Description (min=2 char, max=200 char)

Both fields are required.

Feature requests cannot be edited or deleted.

The system should attempt to prevent duplicate requests when possible.
An error should be raise to indicate duplicate.

Edit / Delete feature requests are out of scope.

⸻

Discovery

Users should be able to quickly discover relevant feature requests.

Supported discovery mechanisms:

* Search
* Sorting
* Ranking
* Trending

⸻

Prioritization

The platform should help identify the most valuable feature requests.

Supported prioritization methods:

* Most Voted
* Trending
* Newest

Trending may combine vote count and recency to surface emerging requests.
Trending should favor recently active requests over historically popular requests.

⸻

Non-Goals

The following capabilities are intentionally out of scope:

* OAuth providers
* Email verification
* Password recovery
* Role-based access control
* Administration panel
* Moderation workflows
* Notifications
* Comments
* Multi-product support

⸻

Success Criteria

A successful implementation allows users to:

* Create an account
* Log in
* Submit feature requests
* Discover requests through search and sorting
* Vote on requests from other users
* View rankings based on popularity and trending score

The application should feel like a realistic product rather than a simple CRUD demonstration.
