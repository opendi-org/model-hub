import React, { createContext, useState, useContext, useEffect } from 'react';
import API_URL from '../config';

const UserContext = createContext();

export const UserProvider = ({ children }) => {
  // Initialize from localStorage if available
  const [user, setUser] = useState(() => {
    const savedUser = localStorage.getItem('user');
    return savedUser ? JSON.parse(savedUser) : null;
  });
  const [loading, setLoading] = useState(true);

  // Create a wrapper for setUser that also updates localStorage
  const setUserWithPersistence = (userData) => {
    setUser(userData);
    if (userData) {
      localStorage.setItem('user', JSON.stringify(userData));
    } else {
      localStorage.removeItem('user');
    }
  };

  const logout = async () => {
    try {
      await fetch(`${API_URL}/auth/logout`, {
        method: 'POST',
        credentials: 'include',
      });
      setUserWithPersistence(null);
    } catch (err) {
      console.error('Logout failed:', err);
      // Still clear user locally even if backend call fails
      setUserWithPersistence(null);
    }
  };

  useEffect(() => {
    // Check if user is already logged in on mount
    const checkAuth = async () => {
      try {
        const response = await fetch(`${API_URL}/auth/me`, {
          credentials: 'include',
        });
        
        if (response.ok) {
          const userData = await response.json();
          console.log('Loaded user from /auth/me:', userData);
          setUserWithPersistence(userData);
        } else {
          // If auth check fails, clear any stale localStorage
          localStorage.removeItem('user');
        }
      } catch (err) {
        console.error('Auth check failed:', err);
        // Clear stalelocalStorage on error
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