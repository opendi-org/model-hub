import React from 'react';
import Grid from '@mui/material/Grid2';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import styled from '@mui/material/styles/styled';
import Container from '@mui/material/Container';
import { useTheme } from '@mui/material/styles';
import TextField from '@mui/material/TextField';
import { useState } from 'react';



const UserPage = () => {
    const theme = useTheme();
    const [menu, setMenu] = useState('Profile');

    const Item = styled(Paper)(() => ({
        backgroundColor: theme.palette.secondary.main,
        padding: theme.spacing(1),
        textAlign: 'center',
        color: theme.secondary,
        ...theme.applyStyles('dark', {
          backgroundColor: '#1A2027',
        width: '100%',
        height: '100%',
        }),
      }));

      function handleProfileUpdate() {
        const username = document.querySelector('input[type="username"]').value;
        const email = document.querySelector('input[type="email"]').value;
        return console.log(username, email);
      }

      function RightHandSide() {
        if (menu === 'Profile') {
          return (
            <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 3, py: 4 }}>
              <Typography variant="h4" gutterBottom>Profile Information</Typography>
              
              <Box sx={{ width: '100%', maxWidth: 400 }}>
                <TextField 
                  fullWidth
                  label="Username"
                  type='username'
                  variant="outlined"
                  margin="normal"
                  defaultValue={localStorage.getItem('username') || ''}
                />
                
                <TextField 
                  fullWidth
                  label="Email"
                  type="email"
                  variant="outlined"
                  margin="normal"
                  defaultValue={localStorage.getItem('email') || ''}
                />
                
                <TextField 
                  fullWidth
                  label="Password"
                  type="password"
                  variant="outlined"
                  margin="normal"
                />
                
                <Button 
                  variant="contained" 
                  color="primary"
                  sx={{ mt: 3 }}
                  onClick={handleProfileUpdate}
                >
                  Update Profile
                </Button>
              </Box>
            </Box>
          )
        }
      }

      return (
        <Container sx={{height: '90%', minHeight: '70vh', padding: 0, margin: 0}}>
          <Grid container spacing={2} height="100%">
            <Grid size={2} spacing={2} height="85vh">
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2, height: '100%' }}>
                <Item 
                  sx={{ 
                    cursor: 'pointer', 
                    '&:hover': { backgroundColor: theme.palette.action.hover } 
                  }}
                  onClick={() => console.log('Profile clicked')}
                >
                  <Typography variant="subtitle1">Profile</Typography>
                </Item>
                <Item 
                  sx={{ 
                    cursor: 'pointer', 
                    '&:hover': { backgroundColor: theme.palette.action.hover } 
                  }}
                  onClick={() => console.log('Assets clicked')}
                >
                  <Typography variant="subtitle1">Assets</Typography>
                </Item>
                <Item 
                  sx={{ 
                    cursor: 'pointer', 
                    '&:hover': { backgroundColor: theme.palette.action.hover } 
                  }}
                  onClick={() => console.log('Settings clicked')}
                >
                  <Typography variant="subtitle1">Settings</Typography>
                </Item>
              </Box>
            </Grid>
            <Grid size={10} >
              <RightHandSide />
            </Grid>
          </Grid>
        </Container>
      );
};

export default UserPage;