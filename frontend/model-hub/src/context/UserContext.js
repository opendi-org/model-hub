import React, { createContext, useState, useContext, useEffect } from 'react';
import APIClient from '../util/ApiClient';

const UserContext = createContext();

const normalizeUser = (userData) => {
  if (!userData) return userData;
  if (userData.picture) return userData;
  if (userData.avatarURL) return { ...userData, picture: userData.avatarURL };
  return userData;
};

export const UserProvider = ({ children }) => {
  // Initialize from localStorage if available
  const [user, setUser] = useState(() => {
    const savedUser = localStorage.getItem('user');
    return savedUser ? JSON.parse(savedUser) : null;
  });
  const [loading, setLoading] = useState(true);

  // Create a wrapper for setUser that also updates localStorage
  const setUserWithPersistence = (userData) => {
    const normalized = normalizeUser(userData);
    setUser(normalized);
    if (normalized) {
      localStorage.setItem('user', JSON.stringify(normalized));
    } else {
      localStorage.removeItem('user');
    }
  };

  const logout = async () => {
    try {
      await APIClient.logout();
    } catch (err) {
      console.error('Logout failed:', err);
    } finally {
      setUserWithPersistence(null);
    }
  };

  useEffect(() => {
    const checkAuth = async () => {
      try {
        const userData = await APIClient.getCurrentUser();
        setUserWithPersistence(userData);
      } catch (err) {
        localStorage.removeItem('user');
      } finally {
        setLoading(false);
      }
    };

    checkAuth();
  }, []);

  return (
    <UserContext.Provider value={{ user, setUser: setUserWithPersistence, logout, loading }}>
      {children}
    </UserContext.Provider>
  );
};

export const useUser = () => {
  const context = useContext(UserContext);
  if (!context) {
    throw new Error('useUser must be used within UserProvider');
  }
  return context;
};