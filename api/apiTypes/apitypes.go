package apiTypes

import (
	"encoding/json"
	"time"
)

type CausalDecisionModel struct {
	ID        int       `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Schema    string    `json:"$schema"`
	MetaID    int       `json:"-"`
	Meta      Meta      `json:"meta"`
	Diagrams  []Diagram `gorm:"many2many:cdm_diagrams" json:"diagrams,omitempty"`
}

type Meta struct {
	ID            int             `gorm:"primaryKey" json:"-"`
	CreatedAt     time.Time       `json:"-"`
	UpdatedAt     time.Time       `json:"-"`
	UUID          string          `json:"uuid"`
	Name          string          `json:"name,omitempty"`
	Summary       string          `json:"summary,omitempty"`
	Documentation json.RawMessage `json:"documentation,omitempty"`
	Version       string          `json:"version,omitempty"`
	Draft         bool            `json:"draft,omitempty"`
	Creator       string          `json:"creator,omitempty"`
	CreatedDate   string          `json:"createdDate,omitempty"`
	Updator       string          `json:"updator,omitempty"`
	UpdatedDate   string          `json:"updatedDate,omitempty"`
}

type Diagram struct {
	ID           int                `gorm:"primaryKey" json:"-"`
	CreatedAt    time.Time          `json:"-"`
	UpdatedAt    time.Time          `json:"-"`
	MetaID       int                `json:"-"`
	Meta         Meta               `json:"meta"`
	Elements     []DiaElement       `gorm:"many2many:diagram_elements" json:"elements,omitempty"`
	Dependencies []CausalDependency `gorm:"many2many:diagram_dependencies" json:"dependencies,omitempty"`
	Addons       json.RawMessage    `json:"addons,omitempty"`
}

type DiaElement struct {
	ID                 int             `gorm:"primaryKey" json:"-"`
	CreatedAt          time.Time       `json:"-"`
	UpdatedAt          time.Time       `json:"-"`
	MetaID             int             `json:"-"`
	Meta               Meta            `json:"meta"`
	CausalType         string          `json:"causalType"`
	DiagramType        string          `json:"diaType"`
	Content            json.RawMessage `json:"content"`
	AssociatedElements json.RawMessage `json:"associatedEvalElements,omitempty"`
}

type CausalDependency struct {
	ID        int       `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	MetaID    int       `json:"-"`
	Meta      Meta      `json:"meta"`
	Source    string    `json:"source"`
	Target    string    `json:"target"`
}
