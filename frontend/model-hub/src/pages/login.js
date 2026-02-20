import { 
  Container, 
  Box, 
  Button, 
  Typography, 
} from '@mui/material';
import { Google } from '@mui/icons-material';
import { useTheme } from '@mui/material/styles';
import API_URL from '../config';

const Login = () => {
  const theme = useTheme();

  const handleGoogleLogin = () => {
    window.location.href = `${API_URL}/auth/google/login`;
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