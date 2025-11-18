import React, { useState, useEffect } from 'react';
import Grid from '@mui/material/Grid2';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Container from '@mui/material/Container';
import Avatar from '@mui/material/Avatar';
import API_URL from '../config';
import ModelMinicard from '../components/ModelMinicard';


const UserPage = () => {
    const [user, setUser] = useState(null);
    const [ownedModels, setOwnedModels] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);


    useEffect(() => {
        const fetchData = async () => {
            try {
                setLoading(true);

                const userResponse = await fetch (`${API_URL}/auth/me`, {
                    credentials: 'include',
                });

                if(!userResponse.ok) {
                    if(userResponse.status === 401) {
                        throw new Error('Not logged in.');
                    }
                    throw new Error('Could not fetch user.');
                }

                const userData = await userResponse.json();
                setUser(userData);

                if(userData && userData.username) {
                    const modelsResponse = await fetch(`${API_URL}/v0/models/search/user/${userData.id}`);

                    if (!modelsResponse.ok) {
                        throw new Error('Could not fetch user models.');
                    }

                    const modelsData = await modelsResponse.json();

                    if (modelsData) {
                        const userOwnedModels = modelsData.filter(model => model.addons.ownerID === userData.id);
                        setOwnedModels(userOwnedModels);
                    } else {
                        setOwnedModels([]);
                    }
                }
            } catch(err) {
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, []);

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
            <Typography variant="h4" component="h2" gutterBottom>Owned Models</Typography>
            {ownedModels.length > 0 ? (
                <Grid container spacing={3}>
                    {ownedModels.map(model => (
                        <ModelMinicard
                            key={model.meta.uuid}
                            id={model.meta.uuid}
                            name={model.meta.name}
                            author={model.meta.creator.username}
                            summary={model.meta.summary}
                            version={model.meta.version}
                            updatedDate={new Date(model.meta.updatedAt).toLocaleDateString()}
                        />
                    ))}
                </Grid>
            ) : (
                <Typography>You do not own any models.</Typography>
            )}
        </Container>
    );
};

export default UserPage;