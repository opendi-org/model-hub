package hub

// Package hub defines the mutable, hub-layer DB entities used by the OpenDI
// model hub: users, auth identities, repositories, collaborators, and CDM tag
// pointers. Unlike the immutable CDM content tables in the database layer,
// these tables support updates (and, for some, hard deletes). See auth.go,
// repository.go, and tag.go for the concrete models and rationale.
