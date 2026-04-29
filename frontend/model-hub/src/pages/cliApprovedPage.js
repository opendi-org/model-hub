import React from 'react';
import { Box, Button, Card, Container, Stack, Typography } from '@mui/material';
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline';
import { Link as RouterLink, useLocation } from 'react-router-dom';

const CliApprovedPage = () => {
  const location = useLocation();
  const message = location.state?.message || 'CLI login approved. You can return to your terminal.';

  return (
    <Container maxWidth="md" sx={{ py: 8 }}>
      <Card
        sx={{
          p: { xs: 4, md: 6 },
          borderRadius: 3,
        }}
      >
        <Stack spacing={3} alignItems="center" sx={{ textAlign: 'center' }}>
          <Box
            sx={{
              width: 84,
              height: 84,
              borderRadius: '50%',
              display: 'grid',
              placeItems: 'center',
              bgcolor: 'rgba(8,109,215,0.12)',
              color: 'primary.main',
            }}
          >
            <CheckCircleOutlineIcon sx={{ fontSize: 48 }} />
          </Box>

          <Typography variant="overline" sx={{ letterSpacing: 2, color: 'text.secondary' }}>
            OpenDI CLI
          </Typography>

          <Typography variant="h4" component="h1">
            You're signed in
          </Typography>

          <Typography variant="h6" sx={{ maxWidth: 620, color: 'text.secondary', fontWeight: 500 }}>
            {message}
          </Typography>

          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} sx={{ pt: 1 }}>
            <Button component={RouterLink} to="/" variant="contained" size="large">
              Go to Model Hub
            </Button>
            <Button component={RouterLink} to="/cli-tool" variant="outlined" size="large">
              CLI Guide
            </Button>
          </Stack>
        </Stack>
      </Card>
    </Container>
  );
};

export default CliApprovedPage;
