import { Button, Container, Typography, Stack } from '@mui/material';
import Paper from '@mui/material/Paper';
import Card from '@mui/material/Card';
import CardActions from '@mui/material/CardActions';
import CardContent from '@mui/material/CardContent';
import opendiIcon from '../opendi-icon.png';
const Home = () => {
    return (
        <Container maxWidth={false} sx={{ width:'100%', height: '100vh', alignItems: 'center', justifyContent: 'center', padding: 0, margin: 0 }}>
            <Stack sx={{ height:"100%", width: '100%', alignItems: 'center', justifyContent: 'center' }}>
                <Paper elevation={1} sx={{ height:"50%", width: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: 2, backgroundColor: 'grey.300' }}>
                    <Typography>
                        Get Started With OpenDI
                    </Typography>
                    <Button>
                        Start Here
                    </Button>
                </Paper>
                <Stack sx={{ height:"50%", width: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: 2 }}>
                    <Stack direction="row" spacing={8}>
                        <Card sx={{ minWidth: 275 }}>
                            <CardContent>
                                <Typography gutterBottom sx={{ color: 'text.secondary', fontSize: 14 }}>
                                Model 1
                                </Typography>
                                <Typography variant="body2">
                                Description of model 1
                                <br />
                                </Typography>
                            </CardContent>
                            <CardActions>
                                <Button size="small">View</Button>
                            </CardActions>
                        </Card>

                        <Card sx={{ minWidth: 275 }}>
                            <CardContent>
                                <Typography gutterBottom sx={{ color: 'text.secondary', fontSize: 14 }}>
                                Model 2
                                </Typography>
                                <Typography variant="body2">
                                Description of model 2
                                <br />
                                </Typography>
                            </CardContent>
                            <CardActions>
                                <Button size="small">View</Button>
                            </CardActions>
                        </Card>
                    </Stack>
                    
                </Stack>
            </Stack>
        </Container>
    );
};

export default Home;
