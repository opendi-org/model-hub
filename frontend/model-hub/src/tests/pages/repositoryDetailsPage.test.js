/**
 * RepositoryDetailsPage tests
 */

import React from 'react';
import { render, screen, waitFor, within, fireEvent } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import RepositoryDetailsPage from '../../pages/repositoryDetailsPage';
import APIClient from '../../util/ApiClient';
import { UserProvider } from '../../context/UserContext';
import { RepositoryProvider } from '../../context/RepositoryContext';
import { NotificationProvider } from '../../context/NotificationContext';

jest.mock('../../util/ApiClient');

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function renderPage(owner = 'alice', slug = 'my-repo') {
  return render(
    <NotificationProvider>
      <UserProvider>
        <RepositoryProvider>
          <MemoryRouter initialEntries={[`/repositories/${owner}/${slug}`]}>
            <Routes>
              <Route
                path="/repositories/:owner/:slug"
                element={<RepositoryDetailsPage />}
              />
            </Routes>
          </MemoryRouter>
        </RepositoryProvider>
      </UserProvider>
    </NotificationProvider>
  );
}

function makeRepo(overrides = {}) {
  return {
    id: 1,
    slug: 'my-repo',
    owner: 'alice',
    description: 'A test repository',
    visibility: 'public',
    tags: [],
    collaborators: [],
    ...overrides,
  };
}

function makeTag(name, overrides = {}) {
  return {
    name,
    digest: 'sha256:abc123',
    size: 1024,
    updatedAt: '2024-01-01T00:00:00Z',
    createdBy: 'alice',
    ...overrides,
  };
}

beforeEach(() => {
  jest.clearAllMocks();
  localStorage.clear();
  // Restore default so RepositoryProvider doesn't throw on every test
  APIClient.getRepositories.mockResolvedValue([]);
});

// ── Loading & error states ────────────────────────────────────────────────────
describe('RepositoryDetailsPage — loading & error states', () => {
  test('shows skeleton while fetching repository', () => {
    APIClient.getCurrentUser.mockReturnValue(new Promise(() => {}));
    APIClient.getRepositoryByOwnerSlug.mockReturnValue(new Promise(() => {}));
    renderPage();
    // Loading state: the repo name has not appeared yet
    expect(screen.queryByText('my-repo')).not.toBeInTheDocument();
  });

  test('shows error alert when API call fails', async () => {
    APIClient.getCurrentUser.mockRejectedValue(new Error('not auth'));
    APIClient.getRepositoryByOwnerSlug.mockRejectedValue(new Error('Repository not found'));
    renderPage();
    expect(await screen.findByText(/repository not found/i)).toBeInTheDocument();
  });

  test('renders repository name and description after load', async () => {
    APIClient.getCurrentUser.mockRejectedValue(new Error('not auth'));
    APIClient.getRepositoryByOwnerSlug.mockResolvedValue(makeRepo());
    renderPage();
    expect(await screen.findByText('my-repo')).toBeInTheDocument();
    expect(screen.getByText('A test repository')).toBeInTheDocument();
  });
});

// ── Tag table ─────────────────────────────────────────────────────────────────
describe('RepositoryDetailsPage — tag table', () => {
  test('renders "No tags yet" when tag list is empty', async () => {
    APIClient.getCurrentUser.mockRejectedValue(new Error('not auth'));
    APIClient.getRepositoryByOwnerSlug.mockResolvedValue(makeRepo({ tags: [] }));
    renderPage();
    expect(await screen.findByText(/no tags yet/i)).toBeInTheDocument();
  });

  test('renders a row for each tag returned by the API', async () => {
    APIClient.getCurrentUser.mockRejectedValue(new Error('not auth'));
    APIClient.getRepositoryByOwnerSlug.mockResolvedValue(
      makeRepo({ tags: [makeTag('v1.0'), makeTag('latest')] })
    );
    renderPage();
    expect(await screen.findByText('v1.0')).toBeInTheDocument();
    expect(screen.getByText('latest')).toBeInTheDocument();
  });
});

// ── Add Tag button permissions ────────────────────────────────────────────────
describe('RepositoryDetailsPage — add tag permissions', () => {
  test('Add Tag button is hidden when user is not the owner', async () => {
    APIClient.getCurrentUser.mockResolvedValue({ username: 'bob', email: 'bob@example.com' });
    APIClient.getRepositoryByOwnerSlug.mockResolvedValue(makeRepo()); // owned by alice
    renderPage();
    await screen.findByText('my-repo');
    expect(screen.queryByRole('button', { name: /add tag/i })).not.toBeInTheDocument();
  });

  test('Add Tag button is visible for the repo owner', async () => {
    APIClient.getCurrentUser.mockResolvedValue({ username: 'alice', email: 'alice@example.com' });
    APIClient.getRepositoryByOwnerSlug.mockResolvedValue(makeRepo()); // owned by alice
    renderPage();
    expect(await screen.findByRole('button', { name: /add tag/i })).toBeInTheDocument();
  });
});

// ── Owner-only controls ───────────────────────────────────────────────────────
describe('RepositoryDetailsPage — owner-only controls', () => {
  test('Collaborators button is visible to the repo owner', async () => {
    APIClient.getCurrentUser.mockResolvedValue({ username: 'alice', email: 'alice@example.com' });
    APIClient.getRepositoryByOwnerSlug.mockResolvedValue(makeRepo());
    renderPage();
    expect(await screen.findByRole('button', { name: /collaborators/i })).toBeInTheDocument();
  });

  test('Collaborators button is not shown to non-owners', async () => {
    APIClient.getCurrentUser.mockResolvedValue({ username: 'bob', email: 'bob@example.com' });
    APIClient.getRepositoryByOwnerSlug.mockResolvedValue(makeRepo()); // owned by alice
    renderPage();
    await screen.findByText('my-repo');
    expect(screen.queryByRole('button', { name: /collaborators/i })).not.toBeInTheDocument();
  });
});

// ── Delete tag dialog ─────────────────────────────────────────────────────────
describe('RepositoryDetailsPage — delete tag dialog', () => {
  test('opens delete confirmation dialog when delete icon is clicked', async () => {
    APIClient.getCurrentUser.mockResolvedValue({ username: 'alice', email: 'alice@example.com' });
    APIClient.getRepositoryByOwnerSlug.mockResolvedValue(
      makeRepo({ tags: [makeTag('v1.0')] })
    );
    renderPage();
    await screen.findByText('v1.0');

    // Owner row has 4 buttons: [copy-digest, edit-tag, delete-tag, download-model]
    const dataRow = screen.getAllByRole('row')[1];
    fireEvent.click(within(dataRow).getAllByRole('button')[2]); // delete-tag

    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText(/are you sure you want to delete tag/i)).toBeInTheDocument();
  });

  test('calls deleteTag and removes the row on confirm', async () => {
    APIClient.getCurrentUser.mockResolvedValue({ username: 'alice', email: 'alice@example.com' });
    APIClient.getRepositoryByOwnerSlug.mockResolvedValue(
      makeRepo({ tags: [makeTag('v1.0')] })
    );
    APIClient.deleteTag.mockResolvedValue({});
    renderPage();
    await screen.findByText('v1.0');

    // Open delete dialog
    const dataRow = screen.getAllByRole('row')[1];
    fireEvent.click(within(dataRow).getAllByRole('button')[2]); // delete-tag

    // Confirm in the dialog
    fireEvent.click(
      within(screen.getByRole('dialog')).getByRole('button', { name: /^delete$/i })
    );

    await waitFor(() => {
      expect(APIClient.deleteTag).toHaveBeenCalledWith(1, 'v1.0');
    });
    expect(screen.queryByText('v1.0')).not.toBeInTheDocument();
  });
});
