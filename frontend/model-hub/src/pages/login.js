import { 
  Container, 
  Box, 
  Button, 
  Typography, 
  TextField,
} from '@mui/material';
import { Google } from '@mui/icons-material';
import { useTheme } from '@mui/material/styles';
import { useState } from 'react';
import APIClient from '../util/ApiClient';

const Login = () => {
  const theme = useTheme();
  const [username, setUsername] = useState('');
  const [usernameError, setUsernameError] = useState('');

  const handleGoogleLogin = () => {
    const normalized = username.trim().toLowerCase();
    if (!normalized) {
      setUsernameError('Username is required for first sign-in');
      return;
    }
    setUsernameError('');
    window.location.href = APIClient.getGoogleLoginUrl(normalized);
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
          OpenDI Model Hub
        </Typography>
        
        <TextField
          fullWidth
          label="Username"
          placeholder="choose-a-username"
          value={username}
          error={Boolean(usernameError)}
          helperText={usernameError || 'Required for first login; ignored for existing accounts.'}
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
          onClick={handleGoogleLogin}
          sx={{ 
            mt: 2,
            mb: 2,
            textTransform: 'none',
            fontSize: '1rem',
            py: 1.5,
            backgroundColor: '#1976d2', // or match your theme
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

export default Login;