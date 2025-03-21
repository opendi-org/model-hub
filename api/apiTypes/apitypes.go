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
	ID         int                  `gorm:"primaryKey" json:"-"`
	CreatedAt  time.Time            `json:"-"`
	UpdatedAt  time.Time            `json:"-"`
	Schema     string               `json:"$schema"`
	MetaID     int                  `json:"-"`
	Meta       Meta                 `json:"meta"`
	ParentUUID string               `json:"parentUUID,omitempty"`
	ParentID   *int                 `json:"-"`
	Parent     *CausalDecisionModel `json:"-"`
	Diagrams   []Diagram            `gorm:"many2many:cdm_diagrams" json:"diagrams,omitempty"`
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
	CreatorID     int             `json:"-"`
	Creator       User            `json:"creator,omitempty"`
	CreatedDate   string          `json:"createdDate,omitempty"`
	Updaters      []User          `gorm:"many2many:meta_updaters" json:"updaters,omitempty"`
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

type User struct {
	ID       int    `gorm:"primaryKey" json:"-"`
	UUID     string `gorm:"unique" json:"uuid"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

type Commit struct {
	ID             int       `gorm:"primaryKey" json:"-"`
	ParentCommitID string    `json:"parentCommitID"`
	Diff           string    `json:"diff"`
	UserUUID       string    `json:"useruuid"`
	CDMUUID        string    `json:"cdmuuid"`
	CreatedAt      time.Time `json:"CreatedAt"`
}

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

	// Even if the models somehow have other parents, we don't care
	// about that for equality. In fact, one might consider
	// changing this code to simply check that the meta is
	// equal and return true. After all, if two models have
	// the same name, summary, documentation, version, draft,
	// especially UUID, etc, then it stands to reason that
	// they are the same model.

	return true
}

func (m Meta) Equals(other Meta) bool {
	return m.UUID == other.UUID
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

func (u User) Equals(other User) bool {
	return u.Username == other.Username
}

func (c Commit) Equals(other Commit) bool {
	return c.ParentCommitID == other.ParentCommitID && c.CDMUUID == other.CDMUUID
}
