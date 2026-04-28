/**
 * RepositoriesPage tests
 */

import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import RepositoriesPage from '../../pages/repositoriesPage';
import APIClient from '../../util/ApiClient';
import { UserProvider } from '../../context/UserContext';
import { RepositoryProvider } from '../../context/RepositoryContext';
import { NotificationProvider } from '../../context/NotificationContext';

jest.mock('../../util/ApiClient');

function LocationDisplay() {
  const location = useLocation();
  return <div data-testid="location">{location.pathname}</div>;
}

function renderPage() {
  return render(
    <NotificationProvider>
      <UserProvider>
        <RepositoryProvider>
          <MemoryRouter initialEntries={['/repositories']}>
            <Routes>
              <Route path="/repositories" element={<RepositoriesPage />} />
              <Route path="/repositories/:owner/:slug" element={<LocationDisplay />} />
            </Routes>
          </MemoryRouter>
        </RepositoryProvider>
      </UserProvider>
    </NotificationProvider>
  );
}

beforeEach(() => {
  jest.clearAllMocks();
  localStorage.clear();
  APIClient.getRepositories.mockResolvedValue([]);
});

// ── RepositoriesPage ──────────────────────────────────────────────────────────
describe('RepositoriesPage', () => {
  test('shows loading skeleton while user auth is pending', () => {
    APIClient.getCurrentUser.mockReturnValue(new Promise(() => {}));
    APIClient.getRepositories.mockResolvedValue([]);
    renderPage();
    // While userLoading is true the main heading is replaced by skeleton placeholders
    expect(screen.queryByRole('heading', { name: /my repositories/i })).not.toBeInTheDocument();
  });

  test('navigates to repository details on card click', async () => {
    APIClient.getCurrentUser.mockResolvedValue({ username: 'alice', email: 'alice@example.com' });
    APIClient.getRepositories.mockResolvedValue([
      { id: 1, slug: 'my-repo', owner: 'alice', visibility: 'private', tags: [] },
    ]);
    renderPage();
    await userEvent.click(await screen.findByText('my-repo'));
    await waitFor(() => {
      expect(screen.getByTestId('location')).toHaveTextContent('/repositories/alice/my-repo');
    });
  });
});
