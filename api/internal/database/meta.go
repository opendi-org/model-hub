package database

// meta.go — conversions between cdm.Asset (domain) and MetaFields (DB row).
//
// These are the only two functions allowed to read or write MetaFields.
// Everything else in the package goes through them so the mapping lives in
// exactly one place.

import (
	"encoding/json"

	"gorm.io/datatypes"

	"opendi.org/model-hub/api/internal/models/cdm"
)

// metaToFields converts a domain cdm.Asset into flat MetaFields for DB storage.
// creator/updator/createdDate/updatedDate are intentionally excluded — they are
// hub-managed and must never be stored as part of the content hash.
func metaToFields(a cdm.Asset) MetaFields {
	f := MetaFields{
		UUID:    a.UUID,
		Name:    a.Name,
		Summary: a.Summary,
		Version: a.Version,
		Draft:   a.Draft,
	}
	if len(a.Documentation) > 0 {
		f.Documentation = datatypes.JSON(a.Documentation)
	}
	return f
}

// fieldsToMeta converts stored MetaFields back into a domain cdm.Asset.
func fieldsToMeta(f MetaFields) cdm.Asset {
	a := cdm.Asset{
		UUID:    f.UUID,
		Name:    f.Name,
		Summary: f.Summary,
		Version: f.Version,
		Draft:   f.Draft,
	}
	if len(f.Documentation) > 0 {
		a.Documentation = json.RawMessage(f.Documentation)
	}
	return a
}
