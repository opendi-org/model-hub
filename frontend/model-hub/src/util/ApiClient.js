import HTTPClient from "./HttpClient";

export default class APIClient {
  // --- Authentication ---

  /** GET /auth/me - current user (session cookie). */
  static async getCurrentUser() {
    return HTTPClient.get("/v0/auth/me");
  }

  /** POST /auth/logout - logout. */
  static async logout() {
    const response = await fetch(HTTPClient.baseURL + "/v0/auth/logout", {
      method: "POST",
      credentials: "include",
    });
    if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
    return response.json().catch(() => ({}));
  }

  /** Full URL for Google OAuth login redirect. Backend serves /auth/... at root (no /api prefix). */
  static getGoogleLoginUrl(username, cliCode) {
    const params = new URLSearchParams();
    if (username) params.set('username', username);
    if (cliCode) params.set('cli_code', cliCode);
    const qs = params.toString();
    return `${HTTPClient.baseURL}/v0/auth/login/google/start${qs ? `?${qs}` : ''}`;
  }

  /** GET /auth/google/callback - exchange code/state for user; throws with error message on failure. */
  static async handleGoogleCallback(code, state) {
    const url = `${HTTPClient.baseURL}/v0/auth/login/google/callback?code=${encodeURIComponent(code)}&state=${encodeURIComponent(state)}`;
    const response = await fetch(url, { method: "GET", credentials: "include" });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      const error = new Error(data.error || "Authentication failed");
      if (data.cli_code) {
        error.cliCode = data.cli_code;
      }
      throw error;
    }
    return data;
  }

  // --- Models ---

  /** GET /v0/models - list models. */
  static async getModels() {
    return HTTPClient.get("/v0/models");
  }

  /** GET /v0/models/:uuid - model by UUID. */
  static async getModelByUUID(uuid) {
    return HTTPClient.get(`/v0/models/${encodeURIComponent(uuid)}`);
  }

  /** GET /v0/models/search/:type/:name - search (type e.g. model, user; name is encoded). */
  static async searchModels(type, name) {
    const encoded = encodeURIComponent(name);
    return HTTPClient.get(`/v0/models/search/${encodeURIComponent(type)}/${encoded}`);
  }

  /** GET /v0/models/tag/:tag - model by tag. */
  static async getModelByTag(tag) {
    return HTTPClient.get(`/v0/models/tag/${encodeURIComponent(tag)}`);
  }

  /** POST /v0/models - create model (JSON body). */
  static async uploadModel(data) {
    return HTTPClient.post("/v0/models", data);
  }

  /** GET /v0/models/commits/:tag - commits for model tag. */
  static async getCommitsByTag(tag) {
    return HTTPClient.get(`/v0/models/commits/${encodeURIComponent(tag)}`);
  }

  /** GET /v0/models/commits/latest/:tag - latest commit by tag. */
  static async getLatestCommitByTag(tag) {
    return HTTPClient.get(`/v0/models/commits/latest/${encodeURIComponent(tag)}`);
  }

  /** GET /v0/models/lineage/:tag - lineage. */
  static async getModelLineage(tag) {
    return HTTPClient.get(`/v0/models/lineage/${encodeURIComponent(tag)}`);
  }

  /** GET /v0/models/children/:tag - children. */
  static async getModelChildren(tag) {
    return HTTPClient.get(`/v0/models/children/${encodeURIComponent(tag)}`);
  }

  /** GET /v0/models/version/:uuid/:version - version of model (uuid + version string). */
  static async getModelVersion(uuid, version) {
    return HTTPClient.get(`/v0/models/version/${encodeURIComponent(uuid)}/${encodeURIComponent(version)}`);
  }

  /** GET /v0/models/modelVersion/:tag/:commitVersion - model at commit version (frontend convention). */
  static async getModelVersionByTagAndCommit(tag, commitVersion) {
    return HTTPClient.get(`/v0/models/modelVersion/${encodeURIComponent(tag)}/${encodeURIComponent(commitVersion)}`);
  }

  /** GET /v0/models/privacy/:tag - privacy/shares. */
  static async getModelPrivacy(tag) {
    return HTTPClient.get(`/v0/models/privacy/${encodeURIComponent(tag)}`);
  }

  /** PUT /v0/models/privacy/:tag - update privacy/shares. */
  static async updateModelPrivacy(tag, body) {
    return HTTPClient.put(`/v0/models/privacy/${encodeURIComponent(tag)}`, body);
  }

  // --- Transfers ---

  /** GET /v0/models/transfer/:tag - get transfer. */
  static async getTransfer(tag) {
    return HTTPClient.get(`/v0/models/transfer/${encodeURIComponent(tag)}`);
  }

  /** POST /v0/models/transfer/:tag - create transfer (owner query param in URL). */
  static async createTransfer(tag, ownerEmail) {
    const q = ownerEmail != null ? `?owner=${encodeURIComponent(ownerEmail)}` : "";
    return HTTPClient.post(`/v0/models/transfer/${encodeURIComponent(tag)}${q}`, {});
  }

  /** DELETE /v0/models/transfer/:tag?accept=true|false - accept or decline transfer. */
  static async deleteTransfer(tag, accept) {
    const acceptValue = accept === true ? "true" : "false";
    const url = `/v0/models/transfer/${encodeURIComponent(tag)}?accept=${acceptValue}`;
    const response = await fetch(HTTPClient.baseURL + url, {
      method: "DELETE",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
    });
    if (!response.ok) {
      const err = await response.json().catch(() => ({}));
      throw new Error(err.error || `HTTP ${response.status}`);
    }
    return response.json().catch(() => ({}));
  }

  /** GET /v0/user/transfers - current user's pending transfers. */
  static async getUserPendingTransfers() {
    return HTTPClient.get("/v0/user/transfers");
  }

  /** PUT /v0/models - update model with file body (e.g. JSON file). */
  static async updateModelWithFile(file) {
    const response = await fetch(HTTPClient.baseURL + "/v0/models", {
      method: "PUT",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: file,
    });
    if (!response.ok) throw new Error(`Upload failed: ${response.statusText}`);
    return response.json();
  }

  // --- Repositories ---

  /** GET /v0/repositories - list repos. scope: "mine" | "shared-with-me" | "all". */
  static async getRepositories(scope = "mine") {
    return HTTPClient.get(`/v0/repositories?scope=${encodeURIComponent(scope)}`);
  }

  /** GET /v0/search - global search repositories. Shows results based on auth status. Optional query params: q, visibility, owner, sortBy, sortOrder. */
  static async globalSearch(q = '', visibility = '', owner = '', sortBy = 'updated', sortOrder = 'desc') {
    const params = new URLSearchParams();
    if (q) params.append('q', q);
    if (visibility) params.append('visibility', visibility);
    if (owner) params.append('owner', owner);
    if (sortBy) params.append('sortBy', sortBy);
    if (sortOrder) params.append('sortOrder', sortOrder);
    const queryString = params.toString();
    return HTTPClient.get(`/v0/search${queryString ? '?' + queryString : ''}`);
  }

  /** GET /v0/repo/:id - get repository by ID. */
  static async getRepositoryById(id) {
    return HTTPClient.get(`/v0/repo/${encodeURIComponent(id)}`);
  }

  /** GET /v0/repositories/:owner/:slug - get repository by owner/slug. */
  static async getRepositoryByOwnerSlug(owner, slug) {
    return HTTPClient.get(
      `/v0/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(slug)}`
    );
  }

  /** POST /v0/repositories - create a new repository. */
  static async createRepository(data) {
    return HTTPClient.post("/v0/repositories", data);
  }

  /** PATCH /v0/repo/:id - update repository metadata. */
  static async updateRepository(id, data) {
    return HTTPClient.patch(`/v0/repo/${encodeURIComponent(id)}`, data);
  }

  /** DELETE /v0/repo/:id - delete a repository (204 No Content). */
  static async deleteRepository(id) {
    const response = await fetch(HTTPClient.baseURL + `/v0/repo/${encodeURIComponent(id)}`, {
      method: "DELETE",
      credentials: "include",
    });
    if (!response.ok) {
      const err = await response.json().catch(() => ({}));
      throw new Error(err.error || `HTTP ${response.status}`);
    }
    return {};
  }

  // --- Tags ---

  /** GET /v0/repo/:id/tags - list tags for a repository. */
  static async getRepositoryTags(repoId) {
    return HTTPClient.get(`/v0/repo/${encodeURIComponent(repoId)}/tags`);
  }

  /** GET /v0/repo/:id/tags/:tagName/model - download full CDM JSON for a tag. */
  static async getTagModel(repoId, tagName) {
    return HTTPClient.get(
      `/v0/repo/${encodeURIComponent(repoId)}/tags/${encodeURIComponent(tagName)}/model`
    );
  }

  /** PUT /v0/repo/:id/tags/:tagName - create a tag; overwrite only when options.overwrite=true. */
  static async createOrUpdateTag(repoId, tagName, cdmJson, options = {}) {
    const qs = options.overwrite === true ? '?overwrite=true' : '';
    return HTTPClient.put(
      `/v0/repo/${encodeURIComponent(repoId)}/tags/${encodeURIComponent(tagName)}${qs}`,
      cdmJson
    );
  }

  /** DELETE /v0/repo/:id/tags/:tagName - delete a tag. */
  static async deleteTag(repoId, tagName) {
    return HTTPClient.delete(
      `/v0/repo/${encodeURIComponent(repoId)}/tags/${encodeURIComponent(tagName)}`
    );
  }

  // --- Repository Management (UC-08, UC-09, UC-10, UC-11) ---

  /** POST /v0/repositories/:owner/:slug/fork - fork a repository. */
  static async forkRepository(owner, slug, data) {
    return HTTPClient.post(
      `/v0/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(slug)}/fork`,
      data
    );
  }

  /** PATCH /v0/repositories/:owner/:slug - set repository privacy. */
  static async setRepositoryPrivacy(owner, slug, visibility) {
    return HTTPClient.patch(
      `/v0/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(slug)}`,
      { visibility }
    );
  }

  /** GET /v0/repositories/:owner/:slug/collaborators - list collaborators. */
  static async listCollaborators(owner, slug) {
    return HTTPClient.get(
      `/v0/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(slug)}/collaborators`
    );
  }

  /** PUT /v0/repositories/:owner/:slug/collaborators/:username - add/update collaborator. */
  static async addCollaborator(owner, slug, username, role) {
    return HTTPClient.put(
      `/v0/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(slug)}/collaborators/${encodeURIComponent(username)}`,
      { username, role }
    );
  }

  /** DELETE /v0/repositories/:owner/:slug/collaborators/:username - remove collaborator. */
  static async removeCollaborator(owner, slug, username) {
    return HTTPClient.delete(
      `/v0/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(slug)}/collaborators/${encodeURIComponent(username)}`
    );
  }

  /** POST /v0/repositories/:owner/:slug/transfer - transfer repository ownership. */
  static async transferRepositoryOwnership(owner, slug, newUsername, previousOwnerAccess = 'none') {
    return HTTPClient.post(
      `/v0/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(slug)}/transfer`,
      { username: newUsername, previousOwnerAccess }
    );
  }

  /** GET /v0/repositories/:owner/:slug/lineage - get repository lineage (parent + children). */
  static async getRepositoryLineage(owner, slug) {
    return HTTPClient.get(
      `/v0/repositories/${encodeURIComponent(owner)}/${encodeURIComponent(slug)}/lineage`
    );
  }
}
