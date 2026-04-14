// Auto-mock for APIClient — all methods return resolved promises by default.
// Individual tests override specific methods via jest.spyOn or mockResolvedValueOnce.
const APIClient = {
  getRepositoryByOwnerSlug: jest.fn(),
  getRepositoryById: jest.fn(),
  getRepositoryTags: jest.fn(),
  getTagModel: jest.fn(),
  createOrUpdateTag: jest.fn(),
  deleteTag: jest.fn(),
  createRepository: jest.fn(),
  updateRepository: jest.fn(),
  deleteRepository: jest.fn(),
  listRepositories: jest.fn(),
  getCurrentUser: jest.fn(),
  handleGoogleCallback: jest.fn(),
  transferRepositoryOwnership: jest.fn(),
  forkRepository: jest.fn(),
  listCollaborators: jest.fn(),
  addCollaborator: jest.fn(),
  removeCollaborator: jest.fn(),
};

export default APIClient;
