// Auto-mock for APIClient — all methods return resolved promises by default.
// Individual tests override specific methods via jest.spyOn or mockResolvedValueOnce.
const APIClient = {
  // Repository
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
  getRepositories: jest.fn().mockResolvedValue([]),
  globalSearch: jest.fn().mockResolvedValue([]),
  // Auth
  getCurrentUser: jest.fn(),
  handleGoogleCallback: jest.fn(),
  getGoogleLoginUrl: jest.fn().mockReturnValue('http://localhost/mock-google-login'),
  logout: jest.fn().mockResolvedValue(undefined),
  // Ownership / collaboration
  transferRepositoryOwnership: jest.fn(),
  forkRepository: jest.fn(),
  listCollaborators: jest.fn(),
  addCollaborator: jest.fn(),
  removeCollaborator: jest.fn(),
  // Model (legacy modelPage)
  getModelByUUID: jest.fn().mockResolvedValue({}),
  getLatestCommitByTag: jest.fn().mockResolvedValue({ version: 0 }),
  getCommitsByTag: jest.fn().mockResolvedValue([]),
  getModelVersionByTagAndCommit: jest.fn().mockResolvedValue({}),
  getModelLineage: jest.fn().mockResolvedValue([]),
  getModelChildren: jest.fn().mockResolvedValue([]),
  getTransfer: jest.fn().mockRejectedValue({ message: '404 not found' }),
  createTransfer: jest.fn().mockResolvedValue({}),
  deleteTransfer: jest.fn().mockResolvedValue({}),
  updateModelWithFile: jest.fn().mockResolvedValue({}),
  getModelPrivacy: jest.fn().mockResolvedValue({ shares: [] }),
  updateModelPrivacy: jest.fn().mockResolvedValue({}),
};

export default APIClient;
