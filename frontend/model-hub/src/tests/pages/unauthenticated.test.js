/**
 * Unauthenticated UI tests
 *
 * Basic element-presence and behavior checks for each page/component
 * when no user is logged in.
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

jest.mock('../../util/ApiClient');

jest.mock('../../context/UserContext', () => ({
  useUser: () => ({ user: null, loading: false, logout: jest.fn() }),
  UserProvider: ({ children }) => children,
}));

jest.mock('../../context/RepositoryContext', () => ({
  useRepositories: () => ({
    repositories: [],
    loading: false,
    scope: 'mine',
    changeScope: jest.fn(),
    addRepository: jest.fn(),
    removeRepository: jest.fn(),
    refreshRepositories: jest.fn(),
  }),
  RepositoryProvider: ({ children }) => children,
}));

jest.mock('../../context/NotificationContext', () => ({
  useNotification: () => ({ showNotification: jest.fn() }),
  NotificationProvider: ({ children }) => children,
}));

const mockToggleColorMode = jest.fn();
jest.mock('../../App', () => ({
  useColorMode: () => ({ mode: 'light', toggleColorMode: mockToggleColorMode }),
  ColorModeContext: { Provider: ({ children }) => children },
}));

import Signin from '../../pages/signin';
import Signup from '../../pages/signup';
import CliDownloadPage from '../../pages/cliPage';
import RepositoriesPage from '../../pages/repositoriesPage';
import Navbar from '../../components/Navbar';

const renderInRouter = (ui) =>
  render(<MemoryRouter>{ui}</MemoryRouter>);

// ── Sign In ──────────────────────────────────────────────────────────────────
describe('Sign In page (unauthenticated)', () => {
  test('renders heading and Google sign-in button with no error alert', () => {
    renderInRouter(<Signin />);
    expect(screen.getByRole('heading', { name: /sign in/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /sign in with google/i })).toBeInTheDocument();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });
});

// ── Sign Up ──────────────────────────────────────────────────────────────────
describe('Sign Up page (unauthenticated)', () => {
  test('renders heading, username field, and Google sign-up button', () => {
    renderInRouter(<Signup />);
    expect(screen.getByRole('heading', { name: /create account/i })).toBeInTheDocument();
    expect(screen.getByLabelText(/username/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /sign up with google/i })).toBeInTheDocument();
  });

  test('shows validation error when submitting without a username', () => {
    renderInRouter(<Signup />);
    fireEvent.click(screen.getByRole('button', { name: /sign up with google/i }));
    expect(screen.getByText(/username is required/i)).toBeInTheDocument();
  });
});

// ── CLI Tool ─────────────────────────────────────────────────────────────────
describe('CLI Tool page (unauthenticated)', () => {
  test('shows CLI docs without requiring login', () => {
    renderInRouter(<CliDownloadPage />);
    // Page is public — no login gate
    expect(screen.queryByRole('heading', { name: /login required/i })).not.toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /openDI CLI/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /getting started/i })).toBeInTheDocument();
  });
});

// ── My Repositories ───────────────────────────────────────────────────────────
describe('My Repositories page (unauthenticated)', () => {
  test('shows login gate and hides the repository UI', () => {
    renderInRouter(<RepositoriesPage />);
    expect(screen.getByRole('heading', { name: /login required/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /login with google/i })).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: /my repositories/i })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /create repository/i })).not.toBeInTheDocument();
  });
});

// ── Navbar ────────────────────────────────────────────────────────────────────
describe('Navbar (unauthenticated)', () => {
  beforeEach(() => {
    mockToggleColorMode.mockClear();
    renderInRouter(<Navbar />);
  });

  test('renders nav links including About pointing to opendi.org', () => {
    expect(screen.getByRole('link', { name: /sign in/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /sign up/i })).toBeInTheDocument();
    const aboutLink = screen.getByRole('link', { name: /about/i });
    expect(aboutLink).toHaveAttribute('href', 'https://opendi.org');
  });

  test('dark/light toggle is present and fires toggleColorMode on click', () => {
    const toggle = screen.getByRole('button', { name: /switch to dark mode/i });
    expect(toggle).toBeInTheDocument();
    fireEvent.click(toggle);
    expect(mockToggleColorMode).toHaveBeenCalledTimes(1);
  });
});
