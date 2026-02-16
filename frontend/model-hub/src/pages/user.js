import React, { useState, useEffect } from 'react';
import Grid from '@mui/material/Grid';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Container from '@mui/material/Container';
import Avatar from '@mui/material/Avatar';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';
import Button from '@mui/material/Button';
import Stack from '@mui/material/Stack';
import APIClient from '../util/ApiClient';
import ModelMinicard from '../components/ModelMinicard';


const UserPage = () => {
    const [user, setUser] = useState(null);
    const [ownedModels, setOwnedModels] = useState([]);
    const [pendingTransfers, setPendingTransfers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [transferActionStatus, setTransferActionStatus] = useState(null);


    useEffect(() => {
        const fetchData = async () => {
            try {
                setLoading(true);

                const userData = await APIClient.getCurrentUser();
                setUser(userData);

                if (userData && userData.username) {
                    const modelsData = await APIClient.searchModels('user', userData.email);
                    if (modelsData) {
                        const userOwnedModels = modelsData.filter(model => model.addons.ownerID === userData.id);
                        setOwnedModels(userOwnedModels);
                    } else {
                        setOwnedModels([]);
                    }

                    const transfersData = await APIClient.getUserPendingTransfers();
                    setPendingTransfers(transfersData || []);
                }
            } catch (err) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, []);

    const handleAcceptDecline = async (tag, accept) => {
        try {
            await APIClient.deleteTransfer(tag, accept);
            setPendingTransfers(prev => prev.filter(item => item.modelTag !== tag));
            if (accept) {
                window.location.reload();
            }
        } catch (error) {
            console.error('Error during accept/decline:', error);
            setTransferActionStatus(error.message || 'Network error occurred.');
        }
    };

    if (loading) {
        return <Container sx={{ py: 4 }}><Typography>Loading...</Typography></Container>;
    }

    if (error) {
        return <Container sx={{ py: 4 }}><Typography color="error">{error}</Typography></Container>;
    }

    return (
        <Container sx={{ py: 4 }}>
            {user && (
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 5, p: 3, backgroundColor: 'grey.100', borderRadius: 2, boxShadow: 1 }}>
                    <Avatar 
                        src={user.picture} 
                        alt={user.username} 
                        sx={{ width: 64, height: 64, mr: 2 }} 
                    />
                    <Box>
                        <Typography variant="h5" component="h1">{user.username}</Typography>
                        <Typography variant="body1" color="text.secondary">{user.email}</Typography>
                    </Box>
                </Box>
            )}

            {/* Pending Transfers Section */}
            {pendingTransfers.length > 0 && (
                <Box sx={{ mb: 5 }}>
                    <Typography variant="h4" component="h2" gutterBottom color="primary">Incoming Transfer Requests</Typography>
                    {transferActionStatus && <Typography color="error">{transferActionStatus}</Typography>}
                    <Grid container spacing={3}>
                        {pendingTransfers.map((item, index) => (
                            <Grid key={index} size={{ xs: 12, sm: 6, md: 4 }}>
                                <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column', border: '1px solid' }}>
                                    <CardContent sx={{ flexGrow: 1 }}>
                                        <Typography variant="h6" component="div">
                                            {item.modelName}
                                        </Typography>
                                        <Typography sx={{ mb: 1.5 }} color="text.secondary">
                                            Pending Transfer
                                        </Typography>
                                        <Typography variant="body2">
                                            From User ID: {item.fromUserID}
                                            <br />
                                            Tag: {item.modelTag}
                                        </Typography>
                                    </CardContent>
                                    <CardActions>
                                        <Stack direction="row" spacing={1}>
                                            <Button size="small" variant="contained" color="success" onClick={() => handleAcceptDecline(item.modelTag, true)}>
                                                Accept
                                            </Button>
                                            <Button size="small" variant="outlined" color="error" onClick={() => handleAcceptDecline(item.modelTag, false)}>
                                                Decline
                                            </Button>
                                        </Stack>
                                    </CardActions>
                                </Card>
                            </Grid>
                        ))}
                    </Grid>
                </Box>
            )}

            <Typography variant="h4" component="h2" gutterBottom color="primary">Owned Models</Typography>
            {ownedModels.length > 0 ? (
                <Grid container spacing={3}>
                    {ownedModels.map(model => (
                        <ModelMinicard key={model.meta.UUID} name={model.meta.name} id = {model.meta.UUID} author={model.meta.creator.username} summary={model.meta.summary} 
                                version={model.meta.version} updatedDate={model.meta.updatedDate}/>
                    ))}
                </Grid>
            ) : (
                <Typography>You do not own any models.</Typography>
            )}
        </Container>
    );
};

export default UserPage;