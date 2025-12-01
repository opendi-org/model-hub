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
      // Check if token is in the userData and store it
      if (userData.token) {
        sessionStorage.setItem('auth_token', userData.token);
      }
    } else {
      localStorage.removeItem('user');
      sessionStorage.removeItem('auth_token');
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
      // Clear user locally even if backend call fails
      setUserWithPersistence(null);
    }
  };

  useEffect(() => {
    // Check if user is already logged in
    const checkAuth = async () => {
      try {
        const response = await fetch(`${API_URL}/auth/me`, {
          credentials: 'include',
        });
        
        if (response.ok) {
          const userData = await response.json();
          console.log('Loaded user from /auth/me:', userData);
          
          // Store the token if its included
          if (userData.token) {
            sessionStorage.setItem('auth_token', userData.token);
          }
          
          setUserWithPersistence(userData);
        } else {
          console.log('/auth/me failed with status:', response.status);
          localStorage.removeItem('user');
          sessionStorage.removeItem('auth_token');
        }
      } catch (err) {
        console.error('Auth check failed:', err);
        // Clear stale localStorage on error
        localStorage.removeItem('user');
        sessionStorage.removeItem('auth_token');
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