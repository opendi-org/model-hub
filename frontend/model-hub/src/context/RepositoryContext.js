import React, { createContext, useState, useContext, useEffect, useCallback } from 'react';
import APIClient from '../util/ApiClient';
import { useUser } from './UserContext';

const RepositoryContext = createContext();

export const RepositoryProvider = ({ children }) => {
  const { user } = useUser();
  const [repositories, setRepositories] = useState([]);
  const [loading, setLoading] = useState(true);

  const refreshRepositories = useCallback(async () => {
    if (!user) {
      setRepositories([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      const data = await APIClient.getRepositories('mine');
      setRepositories(Array.isArray(data) ? data : data.repos ?? []);
    } catch (err) {
      console.error('Failed to fetch repositories:', err);
      setRepositories([]);
    } finally {
      setLoading(false);
    }
  }, [user]);

  useEffect(() => {
    refreshRepositories();
  }, [refreshRepositories]);

  const addRepository = (repo) => {
    setRepositories((prev) => [...prev, repo]);
  };

  const removeRepository = (id) => {
    setRepositories((prev) => prev.filter((r) => r.id !== id));
  };

  const updateRepository = (id, data) => {
    setRepositories((prev) =>
      prev.map((r) => (r.id === id ? { ...r, ...data } : r))
    );
  };

  return (
    <RepositoryContext.Provider
      value={{
        repositories,
        loading,
        addRepository,
        removeRepository,
        updateRepository,
        refreshRepositories,
      }}
    >
      {children}
    </RepositoryContext.Provider>
  );
};

export const useRepositories = () => {
  const context = useContext(RepositoryContext);
  if (!context) {
    throw new Error('useRepositories must be used within RepositoryProvider');
  }
  return context;
};
