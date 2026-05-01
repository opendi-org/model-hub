/**
 * Authenticated UI tests
 *
 * Basic element-presence and behavior checks for each page/component
 * when a user is logged in.
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

jest.mock('../../util/ApiClient');

const mockUser = { username: 'testuser', email: 'test@example.com', picture: null };

jest.mock('../../context/UserContext', () => ({
  useUser: () => ({ user: mockUser, loading: false, logout: jest.fn() }),
  UserProvider: ({ children }) => children,
}));

const mockRepos = [
  { id: '1', slug: 'my-model', owner: 'testuser', description: 'A test model', visibility: 'public', updatedAt: new Date().toISOString() },
  { id: '2', slug: 'private-repo', owner: 'testuser', description: '', visibility: 'private', updatedAt: new Date().toISOString() },
];

const mockRepositoriesContext = {
  repositories: mockRepos,
  loading: false,
  scope: 'mine',
  changeScope: jest.fn(),
  addRepository: jest.fn(),
  removeRepository: jest.fn(),
  refreshRepositories: jest.fn(),
};

jest.mock('../../context/RepositoryContext', () => ({
  useRepositories: () => mockRepositoriesContext,
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

import CliDownloadPage from '../../pages/cliPage';
import RepositoriesPage from '../../pages/repositoriesPage';
import Navbar from '../../components/Navbar';

const renderInRouter = (ui) =>
  render(<MemoryRouter>{ui}</MemoryRouter>);

// ── CLI Tool ─────────────────────────────────────────────────────────────────
describe('CLI Tool page (authenticated)', () => {
  test('shows full CLI docs', () => {
    renderInRouter(<CliDownloadPage />);
    expect(screen.getByRole('heading', { name: /openDI CLI/i })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: /getting started/i })).toBeInTheDocument();
    expect(screen.getAllByText(/pipx install opendi/i).length).toBeGreaterThan(0);
  });
});

// ── My Repositories ───────────────────────────────────────────────────────────
describe('My Repositories page (authenticated)', () => {
  test('shows page header and Create Repository button', () => {
    renderInRouter(<RepositoriesPage />);
    expect(screen.getByRole('heading', { name: /my repositories/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /create repository/i })).toBeInTheDocument();
    expect(screen.queryByRole('heading', { name: /login required/i })).not.toBeInTheDocument();
  });

  test('renders a card for each repository', () => {
    renderInRouter(<RepositoriesPage />);
    expect(screen.getByText('my-model')).toBeInTheDocument();
    expect(screen.getByText('private-repo')).toBeInTheDocument();
  });

  test('opens Create Repository dialog when button is clicked', () => {
    renderInRouter(<RepositoriesPage />);
    fireEvent.click(screen.getByRole('button', { name: /create repository/i }));
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByLabelText(/repository name/i)).toBeInTheDocument();
  });
});

// ── Navbar ────────────────────────────────────────────────────────────────────
describe('Navbar (authenticated)', () => {
  beforeEach(() => renderInRouter(<Navbar />));

  test('shows user initials and hides Sign in / Sign up', () => {
    // Avatar renders the first letter of the username as initials
    expect(screen.getByText('T')).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: /sign in/i })).not.toBeInTheDocument();
    expect(screen.queryByRole('link', { name: /sign up/i })).not.toBeInTheDocument();
  });

  test('opens user menu with username and Sign out on avatar click', () => {
    fireEvent.click(screen.getByText('T'));
    expect(screen.getByText('testuser')).toBeInTheDocument();
    expect(screen.getByText('test@example.com')).toBeInTheDocument();
    expect(screen.getByText(/sign out/i)).toBeInTheDocument();
  });
});
