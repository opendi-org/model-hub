import { Button, Container, Typography, Stack } from '@mui/material';
import Grid from '@mui/material/Grid2';
import Paper from '@mui/material/Paper';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import ModelMinicard from '../components/ModelMinicard';
import { useEffect } from 'react';
import { useState } from 'react';
import { useTheme } from '@mui/material/styles';
const Home = () => {
    const [models, setModels] = useState([])
    const theme = useTheme();
    useEffect(() => {
        fetch(`http://${process.env.REACT_APP_MODEL_HUB_ADDRESS}:${process.env.REACT_APP_MODEL_HUB_PORT}/v0/models`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(data => {console.log(data);
                setModels(data)})
            .catch(error => console.error('There was a problem with the fetch operation:', error));
    }, []);
    return (
        <Container maxWidth={false} sx={{ width: '100%', height: '100vh', alignItems: 'center', justifyContent: 'center', padding: 0, margin: 0 }}>
            <Stack sx={{ height: "100%", width: '100%', alignItems: 'center', justifyContent: 'center' }}>
                <Paper elevation={1} sx={{ height: "30%", width: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: 2, backgroundColor: theme.palette.secondary.main}}>
                    <Typography variant="h4" gutterBottom>
                        Get started with OpenDI
                    </Typography>
                    <Typography variant="subtitle1" gutterBottom>
                        Ut enim ad minim veniam, quis nostrud exercitation ullamco
                    </Typography>
                    <Button variant="contained" color="primary" href="https://opendi.org" target="_blank">
                        Start Here
                    </Button>
                </Paper>
                <Stack sx={{ height: "70%", width: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: 2 }}>
                    <Stack spacing={4}>
                        <Grid xs={12}>
                            <Typography variant="h6">
                                Category X
                            </Typography>
                        </Grid>
                        <Grid xs={12} container spacing={2}>
                            {
                                models.map((model) => 
                                <ModelMinicard key={model.id} name={model.meta.name} id = {model.id} author={model.meta.creator} />)
                            }
                        </Grid>
                        <Grid xs={12}>
                            <Typography variant="h6">
                                Category Y
                            </Typography>
                        </Grid>
                        <Grid xs={12} container spacing={2}>
                            <Grid xs={4}>
                                <Card sx={{ minWidth: 275 }}>
                                    <CardContent>
                                        <Typography gutterBottom sx={{ color: 'text.secondary', fontSize: 14 }}>
                                            Lorem Ipsum
                                        </Typography>
                                        <Typography variant="body2">
                                            Lorem Ipsum
                                            <br />
                                            Lorem Ipsum
                                            <br />
                                            Lorem Ipsum
                                        </Typography>
                                    </CardContent>
                                </Card>
                            </Grid>
                            <Grid xs={4}>
                                <Card sx={{ minWidth: 275 }}>
                                    <CardContent>
                                        <Typography gutterBottom sx={{ color: 'text.secondary', fontSize: 14 }}>
                                            Lorem Ipsum
                                        </Typography>
                                        <Typography variant="body2">
                                            Lorem Ipsum
                                            <br />
                                            Lorem Ipsum
                                            <br />
                                            Lorem Ipsum
                                        </Typography>
                                    </CardContent>
                                </Card>
                            </Grid>
                            <Grid xs={4}>
                                <Card sx={{ minWidth: 275 }}>
                                    <CardContent>
                                        <Typography gutterBottom sx={{ color: 'text.secondary', fontSize: 14 }}>
                                            Lorem Ipsum
                                        </Typography>
                                        <Typography variant="body2">
                                            Lorem Ipsum
                                            <br />
                                            Lorem Ipsum
                                            <br />
                                            Lorem Ipsum
                                        </Typography>
                                    </CardContent>
                                </Card>
                            </Grid>
                        </Grid>
                    </Stack>
                </Stack>
            </Stack>
        </Container>
    );
};

export default Home;
