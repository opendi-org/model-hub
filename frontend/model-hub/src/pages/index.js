//
// COPYRIGHT OpenDI
//

import { Button, Container, Typography, Box } from '@mui/material';
import Grid from '@mui/material/Grid';
import ModelMinicard from '../components/ModelMinicard';
import { useEffect, useState } from 'react';
import { useTheme } from '@mui/material/styles';
import API_URL from '../config';

const Home = () => {
    const [models, setModels] = useState([]);
    const theme = useTheme();

    useEffect(() => {
        fetch(`${API_URL}/v0/models`, { credentials: 'include' })
            .then(response => {
                if (!response.ok) throw new Error('Network response was not ok');
                return response.json();
            })
            .then(data => setModels(data))
            .catch(error => console.error('There was a problem with the fetch operation:', error));
    }, []);

    return (
        <Box sx={{ minHeight: '100vh', backgroundColor: theme.palette.background.default }}>
            {/* Hero section */}
            <Box
                sx={{
                    background: theme.palette.mode === 'light'
                        ? 'linear-gradient(135deg, #1E2130 0%, #0D2B55 100%)'
                        : 'linear-gradient(135deg, #0D1117 0%, #0D2244 100%)',
                    color: '#ffffff',
                    py: { xs: 6, md: 8 },
                    px: 3,
                    textAlign: 'center',
                }}
            >
                <Container maxWidth="md">
                    <Typography
                        variant="h4"
                        fontWeight={700}
                        gutterBottom
                        sx={{ letterSpacing: '-0.02em', mb: 2 }}
                    >
                        Get started with OpenDI
                    </Typography>
                    <Typography
                        variant="subtitle1"
                        sx={{
                            color: 'rgba(255,255,255,0.75)',
                            maxWidth: 680,
                            mx: 'auto',
                            lineHeight: 1.7,
                            mb: 3,
                        }}
                    >
                        The purpose of the OpenDI initiative is to foster a vibrant and healthy
                        ecosystem for decision intelligence (DI), which supports innovative DI
                        research, a healthy vendor market, and — ultimately — better decisions
                        in many domains worldwide.
                    </Typography>
                    <Button
                        variant="contained"
                        size="large"
                        href="https://opendi.org"
                        target="_blank"
                        rel="noopener noreferrer"
                        sx={{
                            backgroundColor: '#086DD7',
                            '&:hover': { backgroundColor: '#0558AE' },
                            px: 4,
                            py: 1.25,
                            fontWeight: 700,
                            fontSize: '0.95rem',
                        }}
                    >
                        Start Here
                    </Button>
                </Container>
            </Box>

            {/* Models grid */}
            <Container maxWidth="xl" sx={{ py: 4, px: { xs: 2, md: 4 } }}>
                <Box sx={{ mb: 3, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <Typography variant="h6" fontWeight={600} color="text.primary">
                        All Models
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                        {models.length} {models.length === 1 ? 'result' : 'results'}
                    </Typography>
                </Box>

                {models.length === 0 ? (
                    <Box
                        sx={{
                            textAlign: 'center',
                            py: 8,
                            color: 'text.secondary',
                        }}
                    >
                        <Typography variant="body1">No models found.</Typography>
                        <Typography variant="body2" sx={{ mt: 1 }}>
                            Be the first to upload a model.
                        </Typography>
                        <Button
                            variant="contained"
                            href="/uploadpage"
                            sx={{ mt: 2 }}
                        >
                            Upload a model
                        </Button>
                    </Box>
                ) : (
                    <Grid container spacing={2}>
                        {models.map((model) => (
                            <ModelMinicard
                                key={model.meta.UUID}
                                name={model.meta.name}
                                id={model.meta.UUID}
                                author={model.meta.creator.username}
                                summary={model.meta.summary}
                                version={model.meta.version}
                                updatedDate={model.meta.updatedDate}
                            />
                        ))}
                    </Grid>
                )}
            </Container>
        </Box>
    );
};

export default Home;
