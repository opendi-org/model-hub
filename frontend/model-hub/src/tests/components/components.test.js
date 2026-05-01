/**
 * Component tests: ModelMinicard, JsonDiffViewer
 */

import React from 'react';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

vi.mock('../../App', () => ({
  useColorMode: () => ({ mode: 'light', toggleColorMode: jest.fn() }),
  ColorModeContext: { Provider: ({ children }) => children },
}));

import ModelMinicard from '../../components/ModelMinicard';
import JsonDiffViewer from '../../components/JsonDiffViewer';

const renderInRouter = (ui) => render(<MemoryRouter>{ui}</MemoryRouter>);

// ── ModelMinicard ─────────────────────────────────────────────────────────────
describe('ModelMinicard', () => {
  const baseProps = {
    id: 'abc-123',
    name: 'My Model',
    author: 'testuser',
    summary: 'A great model',
    version: '1.0',
    updatedDate: '2024-01-01',
  };

  test('renders name, author, summary, version, and updated date', () => {
    renderInRouter(<ModelMinicard {...baseProps} />);
    expect(screen.getByText('My Model')).toBeInTheDocument();
    expect(screen.getByText('by testuser')).toBeInTheDocument();
    expect(screen.getByText('A great model')).toBeInTheDocument();
    expect(screen.getByText('v1.0')).toBeInTheDocument();
    expect(screen.getByText('Updated 2024-01-01')).toBeInTheDocument();
  });

  test('links to the correct model page', () => {
    renderInRouter(<ModelMinicard {...baseProps} />);
    expect(screen.getByRole('link')).toHaveAttribute('href', '/model/abc-123');
  });

  test('shows fallback text when no summary is provided', () => {
    renderInRouter(<ModelMinicard {...baseProps} summary={undefined} />);
    expect(screen.getByText('No description provided.')).toBeInTheDocument();
  });

  test('hides version chip when version is not provided', () => {
    renderInRouter(<ModelMinicard {...baseProps} version={undefined} />);
    expect(screen.queryByText(/^v/)).not.toBeInTheDocument();
  });

  test('hides updated date when not provided', () => {
    renderInRouter(<ModelMinicard {...baseProps} updatedDate={undefined} />);
    expect(screen.queryByText(/updated/i)).not.toBeInTheDocument();
  });
});

// ── JsonDiffViewer ────────────────────────────────────────────────────────────
describe('JsonDiffViewer', () => {
  const model = { name: 'coffee', version: '1.0' };

  test('shows "No previous version" when version is 0 and no model exists', () => {
    render(<JsonDiffViewer lastVersionOfModel={null} commit={{ version: 0 }} />);
    expect(screen.getByText('No previous version')).toBeInTheDocument();
  });

  test('shows raw JSON when version is 0 and a model exists', () => {
    render(<JsonDiffViewer lastVersionOfModel={model} commit={{ version: 0 }} />);
    expect(screen.getByText(/coffee/)).toBeInTheDocument();
  });

  test('renders the JSON tree for a replace patch', () => {
    const commit = {
      version: 1,
      diff: JSON.stringify([{ op: 'replace', path: '/name', value: 'espresso' }]),
    };
    render(<JsonDiffViewer lastVersionOfModel={model} commit={commit} />);
    expect(screen.getByTestId('json-tree')).toBeInTheDocument();
  });

  test('renders the JSON tree for an add patch', () => {
    const commit = {
      version: 1,
      diff: JSON.stringify([{ op: 'add', path: '/roast', value: 'dark' }]),
    };
    render(<JsonDiffViewer lastVersionOfModel={model} commit={commit} />);
    expect(screen.getByTestId('json-tree')).toBeInTheDocument();
  });

  test('renders the JSON tree for a remove patch', () => {
    const commit = {
      version: 1,
      diff: JSON.stringify([{ op: 'remove', path: '/version' }]),
    };
    render(<JsonDiffViewer lastVersionOfModel={model} commit={commit} />);
    expect(screen.getByTestId('json-tree')).toBeInTheDocument();
  });
});

