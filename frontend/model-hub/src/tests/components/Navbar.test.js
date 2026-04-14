import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import Navbar from '../../components/Navbar';
import { UserProvider } from '../../context/UserContext';
import { NotificationProvider } from '../../context/NotificationContext';

jest.mock('../../util/ApiClient');

beforeEach(() => {
  jest.clearAllMocks();
});

describe('Navbar — unauthenticated', () => {
  test.todo('shows Login link when no user is logged in');
  test.todo('does not show username or logout button');
});

describe('Navbar — authenticated', () => {
  test.todo('shows username when user is logged in');
  test.todo('shows logout button');
  test.todo('clears user context on logout');
});
