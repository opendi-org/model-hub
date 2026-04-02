package apiTypes

// DTOs optimized for what the frontend needs to render the diagram.

type TagSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RepoSummary struct {
	ID   string       `json:"id"`
	Name string       `json:"name"`
	Tags []TagSummary `json:"tags"`
}

type ModelSummary struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	RepoID   string       `json:"repoId"`
	Versions []string     `json:"versions"`
	Tags     []TagSummary `json:"tags"`
}

// ToTagSummaries converts a list of tag IDs into lightweight summaries.
func ToTagSummaries(tagIDs []string) []TagSummary {
	result := make([]TagSummary, 0, len(tagIDs))
	for _, id := range tagIDs {
		if tag, ok := GetTag(id); ok {
			result = append(result, TagSummary{
				ID:   tag.ID,
				Name: tag.Name,
			})
		}
	}
	return result
}

// ToRepoSummaries converts repos into their frontend summaries.
func ToRepoSummaries(repos []Repo) []RepoSummary {
	result := make([]RepoSummary, 0, len(repos))
	for _, r := range repos {
		result = append(result, RepoSummary{
			ID:   r.ID,
			Name: r.Name,
			Tags: ToTagSummaries(r.TagIDs),
		})
	}
	return result
}

// ToModelSummaries converts models into their frontend summaries.
func ToModelSummaries(models []Model) []ModelSummary {
	result := make([]ModelSummary, 0, len(models))
	for _, m := range models {
		result = append(result, ModelSummary{
			ID:       m.ID,
			Name:     m.Name,
			RepoID:   m.RepoID,
			Versions: append([]string{}, m.Versions...),
			Tags:     ToTagSummaries(m.TagIDs),
		})
	}
	return result
}

