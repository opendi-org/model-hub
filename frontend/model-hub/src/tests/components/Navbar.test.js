/**
 * Navbar tests — sign-out interaction.
 * Note: unauthenticated appearance is covered in unauthenticated.test.js;
 *       authenticated appearance is covered in authenticated.test.js.
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

const mockLogout = jest.fn();
jest.mock('../../context/UserContext', () => ({
  useUser: () => ({
    user: { username: 'testuser', email: 'test@example.com', picture: null },
    loading: false,
    logout: mockLogout,
  }),
  UserProvider: ({ children }) => children,
}));

jest.mock('../../App', () => ({
  useColorMode: () => ({ mode: 'light', toggleColorMode: jest.fn() }),
  ColorModeContext: { Provider: ({ children }) => children },
}));

jest.mock('../../util/ApiClient');

import Navbar from '../../components/Navbar';

beforeEach(() => {
  mockLogout.mockClear();
});

// ── Sign out ──────────────────────────────────────────────────────────────────
describe('Navbar — sign out', () => {
  test('calls logout when Sign out menu item is clicked', () => {
    render(<MemoryRouter><Navbar /></MemoryRouter>);
    // MUI Tooltip propagates its title as aria-label onto the wrapped button
    fireEvent.click(screen.getByRole('button', { name: 'testuser' }));
    fireEvent.click(screen.getByText(/sign out/i));
    expect(mockLogout).toHaveBeenCalled();
  });
});
