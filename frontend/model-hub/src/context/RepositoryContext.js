import React, { createContext, useState, useContext, useEffect, useCallback } from 'react';
import APIClient from '../util/ApiClient';
import { useUser } from './UserContext';

const RepositoryContext = createContext();

export const RepositoryProvider = ({ children }) => {
  const { user } = useUser();
  const [repositories, setRepositories] = useState([]);
  const [loading, setLoading] = useState(true);
  const [scope, setScope] = useState(() => {
    // Initialize scope from localStorage, defaulting to 'mine'
    const savedScope = localStorage.getItem('repositoryScope');
    return savedScope || 'mine';
  });

  const refreshRepositories = useCallback(async (newScope) => {
    const activeScope = newScope ?? scope;
    if (!user && activeScope !== 'all') {
      setRepositories([]);
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      const data = await APIClient.getRepositories(activeScope);
      setRepositories(Array.isArray(data) ? data : data.repositories ?? []);
    } catch (err) {
      console.error('Failed to fetch repositories:', err);
      setRepositories([]);
    } finally {
      setLoading(false);
    }
  }, [user, scope]);

  const changeScope = useCallback((newScope) => {
    setScope(newScope);
    localStorage.setItem('repositoryScope', newScope);
    refreshRepositories(newScope);
  }, [refreshRepositories]);

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
        scope,
        changeScope,
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
