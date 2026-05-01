/**
 * Misc page tests: NotFound, UserPage, AuthCallback
 */

import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';

jest.mock('../../util/ApiClient');

let mockUserValue = { user: null, loading: false, logout: jest.fn(), setUser: jest.fn() };
jest.mock('../../context/UserContext', () => ({
  useUser: () => mockUserValue,
  UserProvider: ({ children }) => children,
}));

import APIClient from '../../util/ApiClient';
import NotFound from '../../pages/NotFound';
import UserPage from '../../pages/user';
import AuthCallback from '../../pages/AuthCallback';

function LocationDisplay() {
  const location = useLocation();
  return <div data-testid="location">{location.pathname}</div>;
}

beforeEach(() => {
  jest.clearAllMocks();
  mockUserValue = { user: null, loading: false, logout: jest.fn(), setUser: jest.fn() };
});

// ── NotFound ──────────────────────────────────────────────────────────────────
describe('NotFound page', () => {
  test('shows "Page not found" with Go home and Go to login buttons', () => {
    render(<MemoryRouter initialEntries={['/some-random-page']}><NotFound /></MemoryRouter>);
    expect(screen.getByText(/page not found/i)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /go home/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /go to login/i })).toBeInTheDocument();
  });

  test('shows auth redirect hint when path looks like an auth callback', () => {
    render(<MemoryRouter initialEntries={['/auth/callback']}><NotFound /></MemoryRouter>);
    expect(screen.getByText(/authentication redirect/i)).toBeInTheDocument();
  });

  test('shows generic not found message for normal paths', () => {
    render(<MemoryRouter initialEntries={['/does-not-exist']}><NotFound /></MemoryRouter>);
    expect(screen.getByText(/the page you requested does not exist/i)).toBeInTheDocument();
  });
});

// ── UserPage ──────────────────────────────────────────────────────────────────
describe('UserPage', () => {
  test('shows loading state while auth is pending', () => {
    mockUserValue = { user: null, loading: true, logout: jest.fn() };
    render(<MemoryRouter><UserPage /></MemoryRouter>);
    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  test('shows sign-in prompt when no user is logged in', () => {
    mockUserValue = { user: null, loading: false, logout: jest.fn(), setUser: jest.fn() };
    render(<MemoryRouter><UserPage /></MemoryRouter>);
    expect(screen.getByText(/please sign in/i)).toBeInTheDocument();
  });

  test('shows username and email for logged-in user', () => {
    mockUserValue = { user: { username: 'testuser', email: 'test@example.com', picture: null }, loading: false, logout: jest.fn() };
    render(<MemoryRouter><UserPage /></MemoryRouter>);
    expect(screen.getByText('testuser')).toBeInTheDocument();
    expect(screen.getByText('test@example.com')).toBeInTheDocument();
  });
});

// ── AuthCallback ──────────────────────────────────────────────────────────────
describe('AuthCallback', () => {
  test('shows "Signing you in..." spinner on initial render', () => {
    APIClient.handleGoogleCallback.mockReturnValue(new Promise(() => {}));
    render(
      <MemoryRouter initialEntries={['/auth/callback?code=abc&state=xyz']}>
        <AuthCallback />
      </MemoryRouter>
    );
    expect(screen.getByText(/signing you in/i)).toBeInTheDocument();
  });

  test('calls handleGoogleCallback with code and state from URL', async () => {
    APIClient.handleGoogleCallback.mockResolvedValue({ access_token: 'tok' });
    APIClient.getCurrentUser.mockResolvedValue({ username: 'testuser' });
    render(
      <MemoryRouter initialEntries={['/auth/callback?code=mycode&state=mystate']}>
        <AuthCallback />
      </MemoryRouter>
    );
    await waitFor(() => {
      expect(APIClient.handleGoogleCallback).toHaveBeenCalledWith('mycode', 'mystate');
    });
  });

  test('navigates to home on successful authentication', async () => {
    APIClient.handleGoogleCallback.mockResolvedValue({ access_token: 'tok' });
    APIClient.getCurrentUser.mockResolvedValue({ username: 'testuser' });
    render(
      <MemoryRouter initialEntries={['/auth/callback?code=abc&state=xyz']}>
        <Routes>
          <Route path="/auth/callback" element={<AuthCallback />} />
          <Route path="/" element={<LocationDisplay />} />
        </Routes>
      </MemoryRouter>
    );
    await waitFor(() => {
      expect(screen.getByTestId('location')).toHaveTextContent('/');
    });
  });

  test('navigates to /signin on authentication error', async () => {
    APIClient.handleGoogleCallback.mockRejectedValue(new Error('auth failed'));
    render(
      <MemoryRouter initialEntries={['/auth/callback?code=abc&state=xyz']}>
        <Routes>
          <Route path="/auth/callback" element={<AuthCallback />} />
          <Route path="/signin" element={<LocationDisplay />} />
        </Routes>
      </MemoryRouter>
    );
    await waitFor(() => {
      expect(screen.getByTestId('location')).toHaveTextContent('/signin');
    });
  });
});
