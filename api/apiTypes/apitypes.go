//
// COPYRIGHT OpenDI
//

package apiTypes

import (
	"bytes"
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
	UUID          string          `gorm:"unique" json:"uuid"`
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

// equals methods for all structs for testing purposes
func (cdm CausalDecisionModel) Equals(other CausalDecisionModel) bool {
	if cdm.ID != other.ID || cdm.Schema != other.Schema || cdm.MetaID != other.MetaID {
		return false
	}

	if !cdm.Meta.Equals(other.Meta) {
		return false
	}

	if len(cdm.Diagrams) != len(other.Diagrams) {
		return false
	}

	for i, d := range cdm.Diagrams {
		if !d.Equals(other.Diagrams[i]) {
			return false
		}
	}

	return true
}

func (m Meta) Equals(other Meta) bool {
	return m.UUID == other.UUID &&
		m.Name == other.Name &&
		m.Summary == other.Summary &&
		bytes.Equal(m.Documentation, other.Documentation) &&
		m.Version == other.Version &&
		m.Draft == other.Draft &&
		m.Creator == other.Creator &&
		m.CreatedDate == other.CreatedDate &&
		m.Updator == other.Updator &&
		m.UpdatedDate == other.UpdatedDate
}

func (d Diagram) Equals(other Diagram) bool {
	if d.ID != other.ID || d.MetaID != other.MetaID || !d.Meta.Equals(other.Meta) {
		return false
	}

	if len(d.Elements) != len(other.Elements) || len(d.Dependencies) != len(other.Dependencies) {
		return false
	}

	for i, e := range d.Elements {
		if !e.Equals(other.Elements[i]) {
			return false
		}
	}

	for i, dep := range d.Dependencies {
		if !dep.Equals(other.Dependencies[i]) {
			return false
		}
	}

	return bytes.Equal(d.Addons, other.Addons)
}

func (e DiaElement) Equals(other DiaElement) bool {
	return e.ID == other.ID && e.MetaID == other.MetaID && e.Meta.Equals(other.Meta) &&
		e.CausalType == other.CausalType && e.DiagramType == other.DiagramType &&
		bytes.Equal(e.Content, other.Content) &&
		bytes.Equal(e.AssociatedElements, other.AssociatedElements)
}

func (dep CausalDependency) Equals(other CausalDependency) bool {
	return dep.ID == other.ID && dep.MetaID == other.MetaID && dep.Meta.Equals(other.Meta) &&
		dep.Source == other.Source && dep.Target == other.Target
}
