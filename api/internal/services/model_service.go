package services

import (
	"errors"

	"gorm.io/gorm"

	"opendi.org/model-hub/api/internal/database"
	"opendi.org/model-hub/api/internal/models/cdm"
	"opendi.org/model-hub/api/internal/models/hub"
)

// ErrTagNotFound is returned when the requested tag does not exist.
var ErrTagNotFound = errors.New("tag not found")

// ErrTagAlreadyExists is returned when trying to create a tag that already exists.
var ErrTagAlreadyExists = errors.New("tag already exists")

// UploadModelResult holds the outcome of a successful model upload.
type UploadModelResult struct {
	Tag    string
	Digest string
	Size   int
}

// UploadModel validates, stores the CDM payload, and creates the tag record.
// repoID is the repository to attach the tag to; createdByID is the uploader's user ID.
func UploadModel(db *gorm.DB, repoID uint, tagName string, raw []byte, createdByID uint, overwrite bool) (*UploadModelResult, error) {
	rootUUID, err := database.SaveCDM(db, raw)
	if err != nil {
		return nil, err
	}

	tag := hub.CDMTag{
		RepoID:      repoID,
		Name:        tagName,
		ModelUUID:   rootUUID,
		SizeBytes:   int64(len(raw)),
		CreatedByID: createdByID,
	}
	if err := db.Create(&tag).Error; err != nil {
		if !isUniqueViolation(err) {
			return nil, err
		}
		if !overwrite {
			return nil, ErrTagAlreadyExists
		}
		// Overwrite requested: update existing tag pointer + size.
		if err := db.Model(&hub.CDMTag{}).
			Where("repo_id = ? AND name = ?", repoID, tagName).
			Updates(map[string]any{
				"model_uuid": rootUUID,
				"size_bytes": int64(len(raw)),
			}).Error; err != nil {
			return nil, err
		}
	}

	return &UploadModelResult{
		Tag:    tagName,
		Digest: rootUUID,
		Size:   len(raw),
	}, nil
}

// DownloadModelResult holds the outcome of a successful model download.
type DownloadModelResult struct {
	Model  *cdm.CausalDecisionModel
	Digest string // tag.ModelUUID content digest
}

// DownloadModel retrieves the CDM for the named tag in the given repository.
// Returns ErrTagNotFound if the tag does not exist.
func DownloadModel(db *gorm.DB, repoID uint, tagName string) (*DownloadModelResult, error) {
	var tag hub.CDMTag
	if err := db.Where("repo_id = ? AND name = ?", repoID, tagName).First(&tag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}

	model, err := database.LoadCDM(db, tag.ModelUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, err
	}

	return &DownloadModelResult{Model: model, Digest: tag.ModelUUID}, nil
}

// DeleteTag deletes the named tag in the given repository.
// Returns ErrTagNotFound if the tag does not exist.
func DeleteTag(db *gorm.DB, repoID uint, tagName string) error {
	result := db.Where("repo_id = ? AND name = ?", repoID, tagName).Delete(&hub.CDMTag{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTagNotFound
	}
	return nil
}
