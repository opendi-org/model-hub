import React from 'react';
import { Box, Button, Container, Typography } from '@mui/material';
import { Link as RouterLink, useLocation } from 'react-router-dom';

const NotFound = () => {
  const location = useLocation();
  const looksLikeAuthRedirect = location.pathname.includes('/auth') || location.search.includes('code=');

  return (
    <Container maxWidth="sm">
      <Box sx={{ mt: 8, textAlign: 'center' }}>
        <Typography variant="h5" gutterBottom>
          Page not found
        </Typography>
        <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
          {looksLikeAuthRedirect
            ? 'This looks like an authentication redirect to an unknown frontend route. Check your GOOGLE_REDIRECT_URL.'
            : 'The page you requested does not exist.'}
        </Typography>
        <Box sx={{ display: 'flex', gap: 1.5, justifyContent: 'center' }}>
          <Button component={RouterLink} to="/" variant="contained">
            Go home
          </Button>
          <Button component={RouterLink} to="/login" variant="outlined">
            Go to login
          </Button>
        </Box>
      </Box>
    </Container>
  );
};

export default NotFound;
