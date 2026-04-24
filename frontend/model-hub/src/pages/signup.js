import { 
  Container, 
  Box, 
  Button, 
  Typography, 
  TextField,
  Alert,
} from '@mui/material';
import { Google } from '@mui/icons-material';
import { useTheme } from '@mui/material/styles';
import { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import APIClient from '../util/ApiClient';

const Signup = () => {
  const theme = useTheme();
  const [searchParams] = useSearchParams();
  const cliCode = searchParams.get('cli_code');
  const [username, setUsername] = useState('');
  const [usernameError, setUsernameError] = useState('');
  const [redirectError, setRedirectError] = useState(null);

  useEffect(() => {
    // Check if there's an error message from the callback
    const errorMsg = searchParams.get('error');
    if (errorMsg) {
      setRedirectError(decodeURIComponent(errorMsg));
    }
  }, [searchParams]);

  const handleGoogleSignup = () => {
    const normalized = username.trim().toLowerCase();
    if (!normalized) {
      setUsernameError('Username is required to create an account');
      return;
    }
    setUsernameError('');
    window.location.href = APIClient.getGoogleLoginUrl(normalized, cliCode);
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
          Create Account
        </Typography>

        {redirectError && (
          <Alert severity="error" sx={{ mb: 2, width: '100%' }}>
            {redirectError}
          </Alert>
        )}
        
        <TextField
          fullWidth
          label="Username"
          placeholder="choose-a-username"
          value={username}
          error={Boolean(usernameError)}
          helperText={usernameError || 'Choose a unique username for your ModelHub account.'}
          onChange={(e) => {
            setUsername(e.target.value);
            if (usernameError) setUsernameError('');
          }}
          sx={{ mb: 1 }}
        />

        <Button
          fullWidth
          variant="contained"
          startIcon={<Google />}
          onClick={handleGoogleSignup}
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
          Sign up with Google
        </Button>
      </Box>
    </Container>
  );
};

export default Signup;
