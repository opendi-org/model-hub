import React, { useEffect, useState, useRef } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Box, CircularProgress, Typography, Alert } from '@mui/material';
import { useUser } from '../context/UserContext';
import APIClient from '../util/ApiClient';

const AuthCallback = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { setUser } = useUser();
  const [error, setError] = useState(null);
  const hasRun = useRef(false);

  useEffect(() => {
    if (hasRun.current) return;
    hasRun.current = true;

    const handleCallback = async () => {
      const code = searchParams.get('code');
      const state = searchParams.get('state');
    
      if (!code || !state) {
        setError('Missing authentication parameters');
        return;
      }
    
      try {
        const data = await APIClient.handleGoogleCallback(code, state);
        console.log('Full auth response:', data);
        if (data.token) {
          sessionStorage.setItem('auth_token', data.token);
          console.log('Token stored in sessionStorage');
        }
        const userData = data.user || data;
        console.log('Setting user:', userData);
        setUser(userData);
        navigate('/', { replace: true });
      } catch (err) {
        console.error('Auth error:', err);
        setError(err.message);
        
        setTimeout(() => {
          navigate('/login', { replace: true });
        }, 3000);
      }
    };

    handleCallback();
  }, []);

  if (error) {
    return (
      <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', mt: 4 }}>
        <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>
        <Typography variant="body2">Redirecting to login...</Typography>
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