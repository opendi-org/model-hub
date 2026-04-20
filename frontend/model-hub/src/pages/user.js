import React, { useState } from 'react';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Container from '@mui/material/Container';
import Avatar from '@mui/material/Avatar';
import { useUser } from '../context/UserContext';


const UserPage = () => {
    const { user, loading } = useUser();
    const [imageError, setImageError] = useState(false);

    const getInitials = () => {
        if (!user) return 'U';
        if (user.username) return user.username[0].toUpperCase();
        if (user.email) return user.email[0].toUpperCase();
        return 'U';
    };

    if (loading) {
        return <Container sx={{ py: 4 }}><Typography>Loading...</Typography></Container>;
    }

    if (!user) {
        return <Container sx={{ py: 4 }}><Typography color="text.secondary">Please sign in to view your profile.</Typography></Container>;
    }

    return (
        <Container sx={{ py: 4 }}>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 5, p: 3, backgroundColor: 'background.paper', borderRadius: 2, boxShadow: 1 }}>
                <Avatar
                    src={!imageError ? (user.picture || user.avatarURL) : undefined}
                    alt={user.username}
                    imgProps={{
                        onError: () => setImageError(true),
                        referrerPolicy: 'no-referrer',
                    }}
                    sx={{ width: 64, height: 64, mr: 2, bgcolor: (imageError || !(user.picture || user.avatarURL)) ? '#086DD7' : undefined }}
                >
                    {(imageError || !(user.picture || user.avatarURL)) && getInitials()}
                </Avatar>
                <Box>
                    <Typography variant="h5" component="h1">{user.username}</Typography>
                    <Typography variant="body1" color="text.secondary">{user.email}</Typography>
                </Box>
            </Box>
        </Container>
    );
};

export default UserPage;