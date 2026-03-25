import HTTPClient from "./HttpClient";

export default class APIClient {
  // --- Authentication ---

  /** GET /auth/me - current user (session cookie). */
  static async getCurrentUser() {
    return HTTPClient.get("/auth/me");
  }

  /** POST /auth/logout - logout. */
  static async logout() {
    const response = await fetch(HTTPClient.baseURL + "/auth/logout", {
      method: "POST",
      credentials: "include",
    });
    if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
    return response.json().catch(() => ({}));
  }

  /** Full URL for Google OAuth login redirect. Backend serves /auth/... at root (no /api prefix). */
  static getGoogleLoginUrl() {
    return `${HTTPClient.baseURL}/auth/google/login`;
  }

  /** GET /auth/google/callback - exchange code/state for user; throws with error message on failure. */
  static async handleGoogleCallback(code, state) {
    const url = `${HTTPClient.baseURL}/auth/google/callback?code=${encodeURIComponent(code)}&state=${encodeURIComponent(state)}`;
    const response = await fetch(url, { method: "GET", credentials: "include" });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(data.error || "Authentication failed");
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

  /** GET /v0/repo/:id - get repository by ID. */
  static async getRepositoryById(id) {
    return HTTPClient.get(`/v0/repo/${encodeURIComponent(id)}`);
  }

  /** POST /v0/repositories - create a new repository. */
  static async createRepository(data) {
    return HTTPClient.post("/v0/repositories", data);
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

  /** PUT /v0/repo/:id/tags/:tagName - create or overwrite a tag (upload CDM JSON). */
  static async createOrUpdateTag(repoId, tagName, cdmJson) {
    return HTTPClient.put(
      `/v0/repo/${encodeURIComponent(repoId)}/tags/${encodeURIComponent(tagName)}`,
      cdmJson
    );
  }

  /** DELETE /v0/repo/:id/tags/:tagName - delete a tag. */
  static async deleteTag(repoId, tagName) {
    return HTTPClient.delete(
      `/v0/repo/${encodeURIComponent(repoId)}/tags/${encodeURIComponent(tagName)}`
    );
  }
}
