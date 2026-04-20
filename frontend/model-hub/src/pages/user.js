import React, { useState, useEffect } from 'react';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Container from '@mui/material/Container';
import Avatar from '@mui/material/Avatar';
import APIClient from '../util/ApiClient';


const UserPage = () => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);


    useEffect(() => {
        const fetchData = async () => {
            try {
                setLoading(true);

                const userData = await APIClient.getCurrentUser();
                setUser(userData);
            } catch (err) {
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
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 5, p: 3, backgroundColor: 'background.paper', borderRadius: 2, boxShadow: 1 }}>
                    <Avatar 
                        src={user.picture} 
                        alt={user.username} 
                        sx={{ width: 64, height: 64, mr: 2 }} 
                    />
                    <Box>
                        <Typography variant="h5" component="h1" color="text.primary">{user.username}</Typography>
                        <Typography variant="body1" color="text.primary">{user.email}</Typography>
                    </Box>
                </Box>
            )}
        </Container>
    );
};

export default UserPage;