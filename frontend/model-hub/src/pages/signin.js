import { 
  Container, 
  Box, 
  Button, 
  Typography, 
  Alert,
} from '@mui/material';
import { Google } from '@mui/icons-material';
import { useTheme } from '@mui/material/styles';
import { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import APIClient from '../util/ApiClient';

const Signin = () => {
  const theme = useTheme();
  const [searchParams] = useSearchParams();
  const [error, setError] = useState(null);
  const cliCode = searchParams.get('cli_code');

  useEffect(() => {
    // Check if there's an error message from the callback
    const errorMsg = searchParams.get('error');
    if (errorMsg) {
      setError(decodeURIComponent(errorMsg));
    }
  }, [searchParams]);

  const handleGoogleSignin = () => {
    // Sign in without a username - backend will check if account exists
    window.location.href = APIClient.getGoogleLoginUrl(undefined, cliCode);
  };

  return (
    <Container component="main" maxWidth="xs">
      <Box
        sx={{
          marginTop: 8,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          backgroundColor: theme.palette.background.paper,
          border: `1px solid ${theme.palette.divider}`,
          padding: 4,
          borderRadius: 1
        }}
      >
        <Typography component="h1" variant="h5" mb={4} color="text.primary">
          Sign In
        </Typography>

        {error && (
          <Alert severity="error" sx={{ mb: 2, width: '100%' }}>
            {error.includes('account not found') 
              ? 'Your Google account is not linked to a ModelHub account. Please create an account first.' 
              : error}
          </Alert>
        )}

        <Typography variant="body2" color="text.secondary" sx={{ mb: 3, textAlign: 'center' }}>
          Sign in with your Google account to access ModelHub.
        </Typography>

        <Button
          fullWidth
          variant="contained"
          startIcon={<Google />}
          onClick={handleGoogleSignin}
          sx={{ 
            mt: 2,
            mb: 2,
            textTransform: 'none',
            fontSize: '1rem',
            py: 1.5,
            backgroundColor: '#1976d2',
            '&:hover': {
              backgroundColor: '#1565c0',
            }
          }}
        >
          Sign in with Google
        </Button>
      </Box>
    </Container>
  );
};

export default Signin;
