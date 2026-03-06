package database

import (
	"time"

	"gorm.io/datatypes"
)

// MetaFields embeds DI-Asset fields as meta_* columns.
// We intentionally omit creator/updator/createdDate/updatedDate for now.
type MetaFields struct {
	UUID          string         `gorm:"primaryKey;type:uuid;not null"`
	Name          string         `gorm:"type:text"`
	Summary       string         `gorm:"type:text"`
	Documentation datatypes.JSON `gorm:"type:jsonb"`
	Version       string         `gorm:"type:text"`
	Draft         *bool          `gorm:"type:boolean"`
}

// CDMModel is the root entity representing api/internal/models/cdm/model.go.
type CDMModel struct {
	CreatedAt time.Time  `gorm:"type:timestamptz;not null"`
	Schema    string     `gorm:"type:text"`
	Meta      MetaFields `gorm:"embedded;embeddedPrefix:meta_"`
	Addons    datatypes.JSON `gorm:"type:jsonb"`
}

func (CDMModel) TableName() string { return "cdm_models" }

// CDMRunnableModel is one runnable model (globally deduped by meta_uuid).
type CDMRunnableModel struct {
	CreatedAt time.Time `gorm:"type:timestamptz;not null"`
	Meta      MetaFields `gorm:"embedded;embeddedPrefix:meta_"`
	Elements  datatypes.JSON `gorm:"type:jsonb"`
	Addons    datatypes.JSON `gorm:"type:jsonb"`
}

func (CDMRunnableModel) TableName() string { return "cdm_runnable_models" }

// CDMDiagram is one diagram (globally deduped by meta_uuid).
type CDMDiagram struct {
	CreatedAt    time.Time `gorm:"type:timestamptz;not null"`
	Meta         MetaFields     `gorm:"embedded;embeddedPrefix:meta_"`
	Addons       datatypes.JSON `gorm:"type:jsonb"`
	Elements     datatypes.JSON `gorm:"type:jsonb"`
	Dependencies datatypes.JSON `gorm:"type:jsonb"`
}

func (CDMDiagram) TableName() string { return "cdm_diagrams" }

// CDMEvaluatableAsset is one evaluatable asset (globally deduped by meta_uuid).
type CDMEvaluatableAsset struct {
	CreatedAt time.Time `gorm:"type:timestamptz;not null"`
	Meta      MetaFields     `gorm:"embedded;embeddedPrefix:meta_"`
	EvalType  string         `gorm:"type:text;not null"`
	Content   datatypes.JSON `gorm:"type:jsonb"`
}

func (CDMEvaluatableAsset) TableName() string { return "cdm_evaluatable_assets" }

// CDMIOValue is one I/O value (globally deduped by meta_uuid).
type CDMIOValue struct {
	CreatedAt time.Time `gorm:"type:timestamptz;not null"`
	Meta      MetaFields     `gorm:"embedded;embeddedPrefix:meta_"`
	Data      datatypes.JSON `gorm:"type:jsonb"`
}

func (CDMIOValue) TableName() string { return "cdm_io_values" }

// CDMControl is one control (globally deduped by meta_uuid).
type CDMControl struct {
	CreatedAt time.Time `gorm:"type:timestamptz;not null"`
	Meta      MetaFields     `gorm:"embedded;embeddedPrefix:meta_"`
	InputOutputValues datatypes.JSON `gorm:"type:jsonb;column:input_output_values"`
	Displays  datatypes.JSON `gorm:"type:jsonb"`
}

func (CDMControl) TableName() string { return "cdm_controls" }

// Link tables: attach globally-deduped children to a model.
// No ordering; only CreatedAt plus the two UUIDs.

type CDMModelRunnableModel struct {
	CreatedAt    time.Time `gorm:"type:timestamptz;not null"`
	ModelUUID    string    `gorm:"primaryKey;type:uuid;not null;column:model_uuid"`
	RunnableUUID string    `gorm:"primaryKey;type:uuid;not null;column:runnable_uuid"`
}

func (CDMModelRunnableModel) TableName() string { return "cdm_model_runnable_models" }

type CDMModelDiagram struct {
	CreatedAt  time.Time `gorm:"type:timestamptz;not null"`
	ModelUUID  string    `gorm:"primaryKey;type:uuid;not null;column:model_uuid"`
	DiagramUUID string   `gorm:"primaryKey;type:uuid;not null;column:diagram_uuid"`
}

func (CDMModelDiagram) TableName() string { return "cdm_model_diagrams" }

type CDMModelEvaluatableAsset struct {
	CreatedAt      time.Time `gorm:"type:timestamptz;not null"`
	ModelUUID      string    `gorm:"primaryKey;type:uuid;not null;column:model_uuid"`
	EvaluatableUUID string   `gorm:"primaryKey;type:uuid;not null;column:evaluatable_uuid"`
}

func (CDMModelEvaluatableAsset) TableName() string { return "cdm_model_evaluatable_assets" }

type CDMModelIOValue struct {
	CreatedAt time.Time `gorm:"type:timestamptz;not null"`
	ModelUUID string    `gorm:"primaryKey;type:uuid;not null;column:model_uuid"`
	IOValueUUID string  `gorm:"primaryKey;type:uuid;not null;column:io_value_uuid"`
}

func (CDMModelIOValue) TableName() string { return "cdm_model_io_values" }

type CDMModelControl struct {
	CreatedAt  time.Time `gorm:"type:timestamptz;not null"`
	ModelUUID  string    `gorm:"primaryKey;type:uuid;not null;column:model_uuid"`
	ControlUUID string   `gorm:"primaryKey;type:uuid;not null;column:control_uuid"`
}

func (CDMModelControl) TableName() string { return "cdm_model_controls" }
