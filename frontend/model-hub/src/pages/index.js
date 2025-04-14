//
// COPYRIGHT OpenDI
//

import { Button, Container, Typography, Stack } from '@mui/material';
import Grid from '@mui/material/Grid2';
import Paper from '@mui/material/Paper';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import ModelMinicard from '../components/ModelMinicard';
import { useEffect } from 'react';
import { useState } from 'react';
import { useTheme } from '@mui/material/styles';
import API_URL from '../config';
const Home = () => {
    const [models, setModels] = useState([])
    const theme = useTheme();
    const keywords = ["Financial", "Medical", "Business", "Technical"];
    useEffect(() => {
        fetch(`${API_URL}/v0/models`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(data => {
                setModels(data)})
            .catch(error => console.error('There was a problem with the fetch operation:', error));
    }, []);

    function typeRenderer(category, model) {
        if (model.meta.summary.includes(category)) {
                                        
            return <ModelMinicard key={model.meta.uuid} name={model.meta.name} id = {model.meta.uuid} author={model.meta.creator.username} summary={model.meta.summary} 
            version={model.meta.version} updatedDate={model.meta.updatedDate}/> 
        }
    }
    return (
        <Container maxWidth={false} sx={{ width: '100%', height: '100vh', alignItems: 'center', justifyContent: 'center', padding: 0, margin: 0 }}>
            <Stack sx={{ height: "100%", width: '100%', alignItems: 'center', justifyContent: 'center', display: 'block' }}>
                <Paper elevation={1} sx={{ height: "30%", width: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: 2, backgroundColor: theme.palette.secondary.main}}>
                    <Typography variant="h4" gutterBottom sx={{textAlign:"center"}}>
                        Get started with OpenDI
                    </Typography>
                    <Typography variant="subtitle1" gutterBottom sx={{textAlign:"center"}}>
                    The purpose of the OpenDI initiative is to foster a vibrant and healthy ecosystem for decision intelligence (DI), which supports innovative DI research, <br />
                    a healthy vendor market, and — ultimately — better decisions in many domains worldwide.
                    </Typography>
                    <Button variant="contained" color="primary" href="https://opendi.org" target="_blank" sx={{mt: 2}}>
                        Start Here
                    </Button>
                </Paper>
                <Stack sx={{ width: '100%', display: 'block', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: 2 }}>
                    <Stack spacing={4}>
                        <Grid xs={12}>
                            <Typography variant="h6">
                                Financial Models
                            </Typography>
                        </Grid>
                        <Grid xs={12} container spacing={2}>
                            {
                                models.map((model) => typeRenderer("Financial", model))
                            }
                        </Grid>
                    </Stack>
                </Stack>

                <Stack sx={{ width: '100%', display: 'block', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: 2 }}>
                    <Stack spacing={4}>
                        <Grid xs={12}>
                            <Typography variant="h6">
                                Medical Models
                            </Typography>
                        </Grid>
                        <Grid xs={12} container spacing={2}>
                            {
                                models.map((model) => typeRenderer("Medical", model))
                            }
                        </Grid>
                    </Stack>
                </Stack>

                <Stack sx={{ width: '100%', display: 'block', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: 2 }}>
                    <Stack spacing={4}>
                        <Grid xs={12}>
                            <Typography variant="h6">
                                Business Models
                            </Typography>
                        </Grid>
                        <Grid xs={12} container spacing={2}>
                            {
                                models.map((model) => typeRenderer("Business", model))
                            }
                        </Grid>
                    </Stack>
                </Stack>

                <Stack sx={{ width: '100%', display: 'block', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: 2 }}>
                    <Stack spacing={4}>
                        <Grid xs={12}>
                            <Typography variant="h6">
                                Technical Models
                            </Typography>
                        </Grid>
                        <Grid xs={12} container spacing={2}>
                            {
                                models.map((model) => typeRenderer("Technical", model))
                            }
                        </Grid>
                    </Stack>
                </Stack>
            </Stack>
        </Container>
    );
};

export default Home;
