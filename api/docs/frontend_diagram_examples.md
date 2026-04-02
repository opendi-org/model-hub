# Frontend Diagram API Examples

This document shows example JSON responses for the diagram-related endpoints.

## GET `/v0/users/{userID}/repos`

Example request:

`GET /v0/users/user-1/repos`

Example response body:

```json
{
  "userId": "user-1",
  "repos": [
    {
      "id": "repo-1",
      "name": "Financial Models Repo",
      "tags": [
        { "id": "tag-a", "name": "Tag A" }
      ]
    },
    {
      "id": "repo-2",
      "name": "ML Experiments Repo",
      "tags": [
        { "id": "tag-b", "name": "Tag B" }
      ]
    }
  ]
}
```

## GET `/v0/repos/{repoID}/models`

Example request:

`GET /v0/repos/repo-1/models`

Example response body:

```json
[
  {
    "id": "model-financial",
    "name": "Financial Model",
    "repoId": "repo-1",
    "versions": ["v1.0", "v1.5"],
    "tags": [
      { "id": "tag-a", "name": "Tag A" }
    ]
  },
  {
    "id": "model-com",
    "name": "COM",
    "repoId": "repo-1",
    "versions": ["v1.0"],
    "tags": [
      { "id": "tag-a", "name": "Tag A" },
      { "id": "tag-b", "name": "Tag B" }
    ]
  }
]
```

## GET `/v0/diagram/repos/{repoID}`

Example request:

`GET /v0/diagram/repos/repo-1`

Example response body:

```json
{
  "repoId": "repo-1",
  "nodes": [
    { "id": "repo:repo-1", "type": "repo", "label": "Financial Models Repo" },
    { "id": "model:model-financial", "type": "model", "label": "Financial Model" },
    { "id": "model:model-com", "type": "model", "label": "COM" },
    { "id": "tag:tag-a", "type": "tag", "label": "Tag A" },
    { "id": "repo:repo-2", "type": "repo", "label": "ML Experiments Repo" },
    { "id": "model:model-ml", "type": "model", "label": "Machine Learning Model" },
    { "id": "tag:tag-b", "type": "tag", "label": "Tag B" }
  ],
  "edges": [
    { "from": "repo:repo-1", "to": "model:model-financial", "label": "contains" },
    { "from": "repo:repo-1", "to": "model:model-com", "label": "contains" },
    { "from": "model:model-financial", "to": "tag:tag-a", "label": "tagged" },
    { "from": "model:model-com", "to": "tag:tag-a", "label": "tagged" },
    { "from": "repo:repo-2", "to": "model:model-ml", "label": "contains" },
    { "from": "model:model-ml", "to": "tag:tag-b", "label": "tagged" },
    { "from": "model:model-com", "to": "tag:tag-b", "label": "tagged" }
  ]
}
```

