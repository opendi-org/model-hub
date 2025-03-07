import React from 'react';
import Grid from '@mui/material/Grid2';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import styled from '@mui/material/styles/styled';
import Container from '@mui/material/Container';
import { useTheme } from '@mui/material/styles';


const UserPage = () => {
    const theme = useTheme();

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

      return (
        <Container sx={{height: '90%', minHeight: '70vh', padding: 0, margin: 0}}>
          <Grid container spacing={2} height="100%">
            <Grid size={4} height="85vh">
              <Item sx={{height:'100%'}}>size=8</Item>
            </Grid>
            <Grid size={8} >
              <Item sx={{backgroundColor: "#FFFFFF", height: '85vh'}}>size=4</Item>
            </Grid>
          </Grid>
        </Container>
      );
};

export default UserPage;