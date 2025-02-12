import { Button, Container, Typography, Stack } from '@mui/material';
import Grid from '@mui/material/Grid2';
import Paper from '@mui/material/Paper';
import Card from '@mui/material/Card';
import CardActions from '@mui/material/CardActions';
import CardContent from '@mui/material/CardContent';
import opendiIcon from '../opendi-icon.png';
import ModelCard from '../components/ModelCard';
const Home = () => {
    return (
        <Container maxWidth={false} sx={{ width: '100%', height: '100vh', alignItems: 'center', justifyContent: 'center', padding: 0, margin: 0 }}>
            <Stack sx={{ height: "100%", width: '100%', alignItems: 'center', justifyContent: 'center' }}>
                <Paper elevation={1} sx={{ height: "30%", width: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: 2, backgroundColor: 'grey.300' }}>
                    <Typography variant="h4" gutterBottom>
                        Get started with OpenDI
                    </Typography>
                    <Typography variant="subtitle1" gutterBottom>
                        Ut enim ad minim veniam, quis nostrud exercitation ullamco
                    </Typography>
                    <Button variant="contained" sx={{ backgroundColor: 'lightblue' }} href="https://opendi.org" target="_blank">
                        Start Here
                    </Button>
                </Paper>
                <Stack sx={{ height: "70%", width: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: 2 }}>
                    <Stack container spacing={4}>
                        <Grid item xs={12}>
                            <Typography variant="h6">
                                Category X
                            </Typography>
                        </Grid>
                        <Grid item xs={12} container spacing={2}>
                            <ModelCard>

                            </ModelCard>
                            <ModelCard>

                            </ModelCard>
                            <ModelCard>
                            
                            </ModelCard>
                        </Grid>
                        <Grid item xs={12}>
                            <Typography variant="h6">
                                Category Y
                            </Typography>
                        </Grid>
                        <Grid item xs={12} container spacing={2}>
                            <Grid item xs={4}>
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
                            <Grid item xs={4}>
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
                            <Grid item xs={4}>
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
