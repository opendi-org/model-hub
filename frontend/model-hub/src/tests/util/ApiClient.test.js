import APIClient from '../../util/ApiClient';
import HTTPClient from '../../util/HttpClient';

jest.mock('../../util/HttpClient');

beforeEach(() => {
  jest.clearAllMocks();
});

describe('APIClient.getRepositoryByOwnerSlug', () => {
  test.todo('calls correct endpoint with encoded owner and slug');
  test.todo('throws when HTTP call fails');
});

describe('APIClient.getTagModel', () => {
  test.todo('calls correct endpoint with encoded repo id and tag name');
  test.todo('returns parsed JSON model');
});

describe('APIClient.createOrUpdateTag', () => {
  test.todo('calls PUT without overwrite param by default');
  test.todo('appends ?overwrite=true when option is set');
});

describe('APIClient.deleteTag', () => {
  test.todo('calls DELETE on correct endpoint');
});
