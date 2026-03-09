package apiTypes

// Core domain types used to describe the frontend diagram view.

type Repo struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	TagIDs []string `json:"tagIds"`
}

type Model struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	RepoID   string   `json:"repoId"`
	Versions []string `json:"versions"`
	TagIDs   []string `json:"tagIds"`
}

type Tag struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	FileTypes []string `json:"fileTypes,omitempty"`
	Notes     string   `json:"notes,omitempty"`
}

// DiagramUser is a lightweight view that links a user to the repos
// they care about for diagram purposes.
type DiagramUser struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	RepoIDs []string `json:"repoIds"`
}

// Node and edge types are optimized for frontend consumption to render graphs.

type DiagramNode struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

type DiagramEdge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label,omitempty"`
}

type DiagramPayload struct {
	RepoID string        `json:"repoId"`
	Nodes  []DiagramNode `json:"nodes"`
	Edges  []DiagramEdge `json:"edges"`
}

// In-memory seed data that mirrors the relationships on the whiteboard diagram.

var (
	tagsByID = map[string]Tag{
		"tag-a": {
			ID:   "tag-a",
			Name: "Tag A",
			FileTypes: []string{
				"notebook",
				"dataset",
			},
		},
		"tag-b": {
			ID:   "tag-b",
			Name: "Tag B",
			FileTypes: []string{
				"model",
			},
		},
	}

	reposByID = map[string]Repo{
		"repo-1": {
			ID:     "repo-1",
			Name:   "Financial Models Repo",
			TagIDs: []string{"tag-a"},
		},
		"repo-2": {
			ID:     "repo-2",
			Name:   "ML Experiments Repo",
			TagIDs: []string{"tag-b"},
		},
	}

	modelsByID = map[string]Model{
		"model-financial": {
			ID:       "model-financial",
			Name:     "Financial Model",
			RepoID:   "repo-1",
			Versions: []string{"v1.0", "v1.5"},
			TagIDs:   []string{"tag-a"},
		},
		"model-ml": {
			ID:       "model-ml",
			Name:     "Machine Learning Model",
			RepoID:   "repo-2",
			Versions: []string{"v1.0", "v1.5"},
			TagIDs:   []string{"tag-b"},
		},
		"model-com": {
			ID:       "model-com",
			Name:     "COM",
			RepoID:   "repo-1",
			Versions: []string{"v1.0"},
			TagIDs:   []string{"tag-a", "tag-b"},
		},
	}

	usersByID = map[string]DiagramUser{
		"user-1": {
			ID:      "user-1",
			Name:    "Example User",
			RepoIDs: []string{"repo-1", "repo-2"},
		},
	}
)

// GetReposForUser returns the repos a user can see, along with a flag for existence.
func GetReposForUser(userID string) ([]Repo, bool) {
	user, ok := usersByID[userID]
	if !ok {
		return nil, false
	}

	repos := make([]Repo, 0, len(user.RepoIDs))
	for _, repoID := range user.RepoIDs {
		if repo, exists := reposByID[repoID]; exists {
			repos = append(repos, repo)
		}
	}

	return repos, true
}

// GetModelsForRepo returns all models that belong to a given repo.
func GetModelsForRepo(repoID string) ([]Model, bool) {
	if _, ok := reposByID[repoID]; !ok {
		return nil, false
	}

	models := make([]Model, 0)
	for _, m := range modelsByID {
		if m.RepoID == repoID {
			models = append(models, m)
		}
	}

	return models, true
}

// GetTag looks up a tag by ID.
func GetTag(tagID string) (Tag, bool) {
	t, ok := tagsByID[tagID]
	return t, ok
}

// BuildDiagramForRepo constructs a node/edge graph centered on a repo and
// includes models in that repo, their tags, and any models in other repos that
// share those tags (mirroring the cross-repo relationships on the whiteboard).
func BuildDiagramForRepo(repoID string) (DiagramPayload, bool) {
	repo, ok := reposByID[repoID]
	if !ok {
		return DiagramPayload{}, false
	}

	nodes := make([]DiagramNode, 0)
	edges := make([]DiagramEdge, 0)

	nodeSeen := make(map[string]bool)

	addNode := func(id, nodeType, label string) {
		if nodeSeen[id] {
			return
		}
		nodes = append(nodes, DiagramNode{
			ID:    id,
			Type:  nodeType,
			Label: label,
		})
		nodeSeen[id] = true
	}

	addEdge := func(from, to, label string) {
		edges = append(edges, DiagramEdge{
			From:  from,
			To:    to,
			Label: label,
		})
	}

	repoNodeID := "repo:" + repo.ID
	addNode(repoNodeID, "repo", repo.Name)

	relevantTagIDs := make(map[string]bool)

	// First, add models in the selected repo and their tags.
	for _, m := range modelsByID {
		if m.RepoID != repoID {
			continue
		}
		modelNodeID := "model:" + m.ID
		addNode(modelNodeID, "model", m.Name)
		addEdge(repoNodeID, modelNodeID, "contains")

		for _, tagID := range m.TagIDs {
			relevantTagIDs[tagID] = true
			if tag, exists := tagsByID[tagID]; exists {
				tagNodeID := "tag:" + tag.ID
				addNode(tagNodeID, "tag", tag.Name)
				addEdge(modelNodeID, tagNodeID, "tagged")
			}
		}
	}

	// Then, add models in other repos that share those tags and link their repos.
	for _, m := range modelsByID {
		if m.RepoID == repoID {
			continue
		}

		shared := false
		for _, tagID := range m.TagIDs {
			if relevantTagIDs[tagID] {
				shared = true
				break
			}
		}
		if !shared {
			continue
		}

		otherRepo, exists := reposByID[m.RepoID]
		if !exists {
			continue
		}

		otherRepoNodeID := "repo:" + otherRepo.ID
		addNode(otherRepoNodeID, "repo", otherRepo.Name)

		modelNodeID := "model:" + m.ID
		addNode(modelNodeID, "model", m.Name)
		addEdge(otherRepoNodeID, modelNodeID, "contains")

		for _, tagID := range m.TagIDs {
			tag, exists := tagsByID[tagID]
			if !exists {
				continue
			}
			tagNodeID := "tag:" + tag.ID
			addNode(tagNodeID, "tag", tag.Name)
			addEdge(modelNodeID, tagNodeID, "tagged")
		}
	}

	return DiagramPayload{
		RepoID: repo.ID,
		Nodes:  nodes,
		Edges:  edges,
	}, true
}

