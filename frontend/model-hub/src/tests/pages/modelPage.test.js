/**
 * ModelPage tests
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

vi.mock('../../util/ApiClient');
vi.mock('../../context/UserContext', () => ({
  useUser: () => ({ user: null, loading: false, logout: jest.fn() }),
  UserProvider: ({ children }) => children,
}));

import APIClient from '../../util/ApiClient';
import ModelPage from '../../pages/modelPage';

beforeEach(() => {
  APIClient.getModelByUUID.mockResolvedValue({});
  APIClient.getLatestCommitByTag.mockResolvedValue({ version: 0 });
  APIClient.getCommitsByTag.mockResolvedValue([]);
  APIClient.getModelVersionByTagAndCommit.mockResolvedValue({});
  APIClient.getModelLineage.mockResolvedValue([]);
  APIClient.getModelChildren.mockResolvedValue([]);
  APIClient.getTransfer.mockRejectedValue({ message: '404' });
});

const renderModelPage = (uuid = 'test-uuid') =>
  render(
    <MemoryRouter initialEntries={[`/model/${uuid}`]}>
      <Routes>
        <Route path="/model/:uuid" element={<ModelPage />} />
      </Routes>
    </MemoryRouter>
  );

// ── ModelPage ─────────────────────────────────────────────────────────────────
describe('ModelPage', () => {
  test('renders all navigation tabs', () => {
    renderModelPage();
    expect(screen.getByRole('tab', { name: /overview/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /documentation/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /commit diff/i })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: /fork info/i })).toBeInTheDocument();
  });

  test('does not show Ownership tab when user is not the owner', () => {
    renderModelPage();
    expect(screen.queryByRole('tab', { name: /ownership/i })).not.toBeInTheDocument();
  });

  test('renders Download and Update buttons', () => {
    renderModelPage();
    expect(screen.getByRole('button', { name: /download/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /update/i })).toBeInTheDocument();
  });

  test('opens Update dialog when Update button is clicked', () => {
    renderModelPage();
    fireEvent.click(screen.getByRole('button', { name: /update/i }));
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(screen.getByText(/update model/i)).toBeInTheDocument();
  });

  test('renders the OpenDI logo image', () => {
    renderModelPage();
    expect(screen.getByAltText(/openDI logo/i)).toBeInTheDocument();
  });
});
