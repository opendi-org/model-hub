import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Box, CircularProgress, Typography, Alert } from '@mui/material';
import { useUser } from '../context/UserContext';
import API_URL from '../config';

const AuthCallback = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { setUser } = useUser();
  const [error, setError] = useState(null);

  useEffect(() => {
    const handleCallback = async () => {
      const code = searchParams.get('code');
      const state = searchParams.get('state');

      if (!code || !state) {
        setError('Missing authentication parameters');
        return;
      }

      try {
        const response = await fetch(
          `${API_URL}/auth/google/callback?code=${code}&state=${state}`,
          {
            method: 'POST',
            credentials: 'include',
          }
        );

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.error || 'Authentication failed');
        }

        const userData = await response.json();
        setUser(userData);
        navigate('/');
      } catch (err) {
        console.error('Auth error:', err);
        setError(err.message);
      }
    };

    handleCallback();
  }, [searchParams, navigate, setUser]);

  if (error) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}>
        <Alert severity="error">{error}</Alert>
      </Box>
    );
  }

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'center',
        alignItems: 'center',
        height: '100vh',
      }}
    >
      <CircularProgress size={60} />
      <Typography variant="h6" sx={{ mt: 2 }}>
        Signing you in...
      </Typography>
    </Box>
  );
};

export default AuthCallback;