package hub

import "time"

// ── CDM Tags ──────────────────────────────────────────────────────────────────
//
// A mutable pointer from (RepoID, Name) → a CDM content digest (ModelUUID).
//
// On every push UpdatedAt is bumped and ModelUUID + SizeBytes are overwritten.
// No history is kept — UpdatedAt is the authoritative "last pushed" timestamp.
//
// CreatedByID is written once on first push and never updated. This lets
// GET .../tags/:tag (UC-19) report the original publisher even after the tag
// has been overwritten many times.
//
// SizeBytes is the byte count of the CDM JSON as served by
// GET .../tags/:tag/model (reconstructed output, not raw upload bytes — the hub
// rewrites meta.uuid, meta.name, meta.version, and meta.createdDate on egress,
// so the sizes are not identical). Stored here so tag-listing responses can
// include size without touching the CDM content tables.
//
// ModelUUID references cdm_models.meta_uuid. No FK constraint is declared; this
// is intentional — see entities.go for the content-addressed storage rationale.
//
// Hard-delete: no DeletedAt column. A soft-deleted tag would block reuse of the
// tag name on the same repo because the unique constraint on (repo_id, name)
// would fire against the hidden row. Tag deletion is final.

type CDMTag struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	CreatedAt   time.Time `gorm:"type:timestamptz;autoCreateTime;not null"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;autoUpdateTime;not null"`
	RepoID      uint      `gorm:"uniqueIndex:idx_tag_ref;not null"`
	Name        string    `gorm:"type:text;uniqueIndex:idx_tag_ref;not null"` // e.g. "latest", "v1.2.0"
	ModelUUID   string    `gorm:"type:uuid;not null"`
	SizeBytes   int64     `gorm:"not null;default:0"`
	CreatedByID uint      `gorm:"not null"`

	// Preload only
	Repo      Repository `gorm:"foreignKey:RepoID"`
	CreatedBy User       `gorm:"foreignKey:CreatedByID"`
}

func (CDMTag) TableName() string { return "hub_cdm_tags" }

