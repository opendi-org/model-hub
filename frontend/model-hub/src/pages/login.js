import React, { useState } from 'react';
import { 
  Container, 
  Box, 
  TextField, 
  Button, 
  Typography, 
} from '@mui/material';
import API_URL from '../config'

const Login = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loginError, setLoginError] = useState(false);

  const handleSubmit = (event) => {
    event.preventDefault();
    
    fetch(`${API_URL}/login?email=${email}&password=${password}`, { method: 'POST', })
        .then(response => {
            if (!response.ok) {
                return response.json().then(error => {
                    setLoginError(true)
                    throw new Error(error.message);
                });
            }
            setLoginError(false)
            return response.json();
        })
        .then(data => {
            sessionStorage.setItem('username', data.username)
            sessionStorage.setItem('email', data.email)
            window.location.href = '/';
        })
        .catch(error => {
            console.error('Error:', error.message);
        });
  };

  return (
    <Container component="main" maxWidth="xs">
      <Box
        sx={{
          marginTop: 8,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          backgroundColor: '#f5f5f5',
          padding: 4,
          borderRadius: 1
        }}
      >
        <Typography component="h1" variant="h5" mb={4}>
          OpenDI Model Hub
        </Typography>
        
        <Box component="form" onSubmit={handleSubmit} noValidate sx={{ mt: 1, width: '100%' }}>
          <TextField
            margin="normal"
            required
            fullWidth
            id="email"
            label="Email"
            name="email"
            autoComplete="email"
            autoFocus
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          <TextField
            margin="normal"
            required
            fullWidth
            name="password"
            label="Password"
            type="password"
            id="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />

          <Box 
            sx={{
              display: loginError ? 'block' : 'none',
              backgroundColor: '#ffebee',
              color: '#c62828',
              padding: 2,
              borderRadius: 1,
              mt: 2
            }}
          >
            <Typography variant="body2">
              Username or password is incorrect
            </Typography>
          </Box>
          <Button
            type="submit"
            fullWidth
            variant="contained"
            color='dark'
            sx={{ 
              mt: 3, 
              mb: 2, 
            }}
          >
            Log in
          </Button>
        </Box>
      </Box>
    </Container>
  );
};

export default Login;