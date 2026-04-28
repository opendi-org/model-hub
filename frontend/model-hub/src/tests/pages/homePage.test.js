/**
 * Home page and Explore (search) page tests
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

jest.mock('../../util/ApiClient');

import APIClient from '../../util/ApiClient';
import Home from '../../pages/index';
import ExplorePage from '../../pages/search';

const renderInRouter = (ui) => render(<MemoryRouter>{ui}</MemoryRouter>);

const makeRepo = (overrides = {}) => ({
  id: '1',
  slug: 'test-repo',
  owner: 'testuser',
  description: 'A test repo',
  visibility: 'public',
  updatedAt: new Date().toISOString(),
  ...overrides,
});

beforeEach(() => jest.clearAllMocks());

// ── Home page ─────────────────────────────────────────────────────────────────
describe('Home page', () => {
  test('renders hero section with Start Here button', async () => {
    APIClient.globalSearch.mockResolvedValue([]);
    renderInRouter(<Home />);
    expect(screen.getByText(/get started with openDI/i)).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /start here/i })).toHaveAttribute('href', 'https://opendi.org');
  });

  test('shows search and filter controls', () => {
    APIClient.globalSearch.mockResolvedValue([]);
    renderInRouter(<Home />);
    expect(screen.getByPlaceholderText(/search repositories/i)).toBeInTheDocument();
  });

  test('shows empty state when no repositories exist', async () => {
    APIClient.globalSearch.mockResolvedValue([]);
    renderInRouter(<Home />);
    expect(await screen.findByText(/no repositories found/i)).toBeInTheDocument();
  });

  test('renders a card for each repository after loading', async () => {
    APIClient.globalSearch.mockResolvedValue([
      makeRepo({ id: '1', slug: 'coffee-model' }),
      makeRepo({ id: '2', slug: 'tea-model' }),
    ]);
    renderInRouter(<Home />);
    expect(await screen.findByText('coffee-model')).toBeInTheDocument();
    expect(screen.getByText('tea-model')).toBeInTheDocument();
  });

  test('shows repository count', async () => {
    APIClient.globalSearch.mockResolvedValue([makeRepo()]);
    renderInRouter(<Home />);
    expect(await screen.findByText(/1 repository/)).toBeInTheDocument();
  });
});

// ── Explore (Search) page ─────────────────────────────────────────────────────
describe('Explore page', () => {
  test('renders heading and empty state when no public repos exist', async () => {
    APIClient.getRepositories.mockResolvedValue([]);
    renderInRouter(<ExplorePage />);
    expect(screen.getByText(/explore public repositories/i)).toBeInTheDocument();
    expect(await screen.findByText(/no public repositories yet/i)).toBeInTheDocument();
  });

  test('renders repository cards after loading', async () => {
    APIClient.getRepositories.mockResolvedValue([
      makeRepo({ id: '1', slug: 'my-repo', owner: 'alice' }),
    ]);
    renderInRouter(<ExplorePage />);
    expect(await screen.findByText('my-repo')).toBeInTheDocument();
    expect(screen.getByText('alice')).toBeInTheDocument();
  });

  test('shows filter input', () => {
    APIClient.getRepositories.mockResolvedValue([]);
    renderInRouter(<ExplorePage />);
    expect(screen.getByPlaceholderText(/filter repositories/i)).toBeInTheDocument();
  });
});
