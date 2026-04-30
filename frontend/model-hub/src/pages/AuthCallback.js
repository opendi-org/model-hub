import React, { useEffect, useState, useRef } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Box, CircularProgress, Typography } from '@mui/material';
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
        
        // Fetch user data from /auth/me after callback to refresh UI context.
        const userData = await APIClient.getCurrentUser();
        setUser(userData);

        const cliCode = data.cliCode || searchParams.get('cli_code');
        if (cliCode) {
          navigate('/auth/cli-approved', {
            replace: true,
            state: {
              message: 'CLI login approved. You can return to your terminal.',
            },
          });
          return;
        }

        navigate('/', { replace: true });
      } catch (err) {
        console.error('Auth error:', err);
        const errorMsg = err.message || 'Authentication failed';
        const encodedError = encodeURIComponent(errorMsg);
        const cliCode = err.cliCode || searchParams.get('cli_code');
        const cliCodeQuery = cliCode ? `&cli_code=${encodeURIComponent(cliCode)}` : '';
        
        // Determine which page to redirect to based on error message
        let redirectPath = '/signin';
        if (errorMsg.includes('account not found')) {
          redirectPath = `/signup?error=${encodedError}${cliCodeQuery}`;
        } else if (errorMsg.includes('account already exists')) {
          redirectPath = `/signin?error=${encodedError}${cliCodeQuery}`;
        } else {
          // Default to signin with error
          redirectPath = `/signin?error=${encodedError}${cliCodeQuery}`;
        }
        
        // Redirect immediately with error in query params
        navigate(redirectPath, { replace: true });
      }
    };

    handleCallback();
  }, [searchParams, setUser, navigate]);

  if (error) {
    return (
      <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', mt: 4 }}>
        <CircularProgress size={60} />
        <Typography variant="h6" sx={{ mt: 2 }}>
          Signing you in...
        </Typography>
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