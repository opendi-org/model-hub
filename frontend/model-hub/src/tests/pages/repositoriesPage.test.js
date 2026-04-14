import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import RepositoriesPage from '../../pages/repositoriesPage';
import APIClient from '../../util/ApiClient';
import { UserProvider } from '../../context/UserContext';
import { RepositoryProvider } from '../../context/RepositoryContext';
import { NotificationProvider } from '../../context/NotificationContext';

jest.mock('../../util/ApiClient');

beforeEach(() => {
  jest.clearAllMocks();
});

describe('RepositoriesPage', () => {
  test.todo('shows loading state while fetching');
  test.todo('renders list of repositories after load');
  test.todo('shows empty state when no repositories exist');
  test.todo('navigates to repository details on card click');
});
