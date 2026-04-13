package hub

import (
	"time"

	"gorm.io/gorm"
)

const (
	CollaboratorRoleRead  = "read"
	CollaboratorRoleWrite = "write"
	CollaboratorRoleAdmin = "admin"
	CollaboratorRoleOwner = "owner"
)

// ── Repositories ──────────────────────────────────────────────────────────────
//
// A Repository is the named container for CDM tags, owned by a User.
// (OwnerID, Slug) is the unique natural key used in all :owner/:slug API paths.
//
// Visibility:
//   "public"  — any caller (including unauthenticated) can read tags and model
//               content. The collaborator list is visible only to the owner and
//               to explicit collaborators (UC-05).
//   "private" — only the owner and explicit collaborators can access anything.
//
// Soft-delete: a deleted Repository row is retained so that foreign key
// references (e.g. tags) remain valid, but the slug is mutated on delete so the
// original slug can be reused immediately.
//
// Lineage (UC-08):
//   ForkedFromID NULL  = original repository.
//   ForkedFromID <id>  = fork; <id> is the immediate parent.
//   Ancestry  = walk ForkedFromID upward (stop at NULL).
//   Children  = SELECT * FROM repositories WHERE forked_from_id = ? AND deleted_at IS NULL.
//   GET .../lineage serves both directions.
//
// Transfer (UC-11):
//   UPDATE repositories SET owner_id = <new_user_id> WHERE id = ?
//   No extra table needed. Run inside one transaction with any collaborator
//   cleanup. See open question in planning doc (Section 4): does the previous
//   owner lose access or become an owner collaborator?

type Repository struct {
	gorm.Model
	OwnerID      uint   `gorm:"not null"`
	Slug         string `gorm:"type:text;uniqueIndex:idx_repo_ref;not null"`
	Description  string `gorm:"type:text"`
	Visibility   string `gorm:"type:text;default:'private';not null"` // "public"|"private"
	ForkedFromID *uint  // NULL = original repo

	// Preload only
	Owner      User        `gorm:"foreignKey:OwnerID"`
	ForkedFrom *Repository `gorm:"foreignKey:ForkedFromID"`
}

func (Repository) TableName() string { return "hub_repositories" }

// ── Collaborators ─────────────────────────────────────────────────────────────
//
// Grants a User explicit access to a Repository.
//
// For public repos a collaborator row is only needed for write/owner access, or
// to make the collaborator list visible to that user (UC-05). For private repos
// it is required for any access at all.
//
// The owner is NOT stored as a collaborator row. Repository.OwnerID is the
// canonical ownership record. All access-check helpers must grant the owner
// full access regardless of the collaborator table.
//
// Roles (ascending privilege):
//   "read"  — fetch tags and model content                          (UC-12, UC-20, UC-26)
//   "write" — push and delete tags                                  (UC-23, UC-24, UC-25)
//   "admin" — manage collaborators                                  (UC-10)
//   "owner" — full access including transferring ownership          (UC-11)
//
// Hard-delete: no DeletedAt column. Revoking access is a plain DELETE. A soft-
// deleted row would silently block re-adding the same user because the unique
// constraint on (repo_id, user_id) would still fire against the hidden row.

type Collaborator struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time `gorm:"type:timestamptz;autoCreateTime;not null"`
	UpdatedAt time.Time `gorm:"type:timestamptz;autoUpdateTime;not null"`
	RepoID    uint      `gorm:"uniqueIndex:idx_collab;not null"`
	UserID    uint      `gorm:"uniqueIndex:idx_collab;not null"`
	Role      string    `gorm:"type:text;default:'read';not null"` // hub.CollaboratorRoleRead|Write|Admin|Owner

	// Preload only
	Repo Repository `gorm:"foreignKey:RepoID"`
	User User       `gorm:"foreignKey:UserID"`
}

func (Collaborator) TableName() string { return "hub_collaborators" }
