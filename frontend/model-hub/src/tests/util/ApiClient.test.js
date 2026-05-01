/**
 * APIClient unit tests — verifies URL construction and HTTP delegation.
 */

import APIClient from '../../util/ApiClient';
import HTTPClient from '../../util/HttpClient';

vi.mock('../../util/HttpClient');

beforeEach(() => {
  jest.clearAllMocks();
});

// ── getRepositoryByOwnerSlug ───────────────────────────────────────────────────
describe('APIClient.getRepositoryByOwnerSlug', () => {
  test('calls correct endpoint with encoded owner and slug', async () => {
    HTTPClient.get.mockResolvedValue({ id: 1 });
    await APIClient.getRepositoryByOwnerSlug('alice smith', 'my repo');
    expect(HTTPClient.get).toHaveBeenCalledWith(
      '/v0/repositories/alice%20smith/my%20repo'
    );
  });

  test('throws when HTTP call fails', async () => {
    HTTPClient.get.mockRejectedValue(new Error('Network error'));
    await expect(
      APIClient.getRepositoryByOwnerSlug('alice', 'repo')
    ).rejects.toThrow('Network error');
  });
});

// ── getTagModel ───────────────────────────────────────────────────────────────
describe('APIClient.getTagModel', () => {
  test('calls correct endpoint with encoded repo id and tag name', async () => {
    HTTPClient.get.mockResolvedValue({});
    await APIClient.getTagModel('repo/42', 'v1.0');
    expect(HTTPClient.get).toHaveBeenCalledWith(
      '/v0/repo/repo%2F42/tags/v1.0/model'
    );
  });

  test('returns the parsed JSON model', async () => {
    const model = { name: 'test-model', version: '1.0' };
    HTTPClient.get.mockResolvedValue(model);
    const result = await APIClient.getTagModel(1, 'latest');
    expect(result).toEqual(model);
  });
});

// ── createOrUpdateTag ─────────────────────────────────────────────────────────
describe('APIClient.createOrUpdateTag', () => {
  test('calls PUT without overwrite param by default', async () => {
    HTTPClient.put.mockResolvedValue({});
    await APIClient.createOrUpdateTag(1, 'v1.0', { data: true });
    expect(HTTPClient.put).toHaveBeenCalledWith('/v0/repo/1/tags/v1.0', { data: true });
  });

  test('appends ?overwrite=true when option is set', async () => {
    HTTPClient.put.mockResolvedValue({});
    await APIClient.createOrUpdateTag(1, 'v1.0', { data: true }, { overwrite: true });
    expect(HTTPClient.put).toHaveBeenCalledWith(
      '/v0/repo/1/tags/v1.0?overwrite=true',
      { data: true }
    );
  });
});

// ── deleteTag ─────────────────────────────────────────────────────────────────
describe('APIClient.deleteTag', () => {
  test('calls DELETE on correct endpoint', async () => {
    HTTPClient.delete.mockResolvedValue({});
    await APIClient.deleteTag(1, 'v1.0');
    expect(HTTPClient.delete).toHaveBeenCalledWith('/v0/repo/1/tags/v1.0');
  });
});
