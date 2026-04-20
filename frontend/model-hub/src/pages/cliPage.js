//
// COPYRIGHT OpenDI
//

import React from 'react';
import {
    Box,
    Button,
    Card,
    Typography,
    Container,
    Divider,
    List,
    ListItem,
    ListItemText,
    CircularProgress
} from "@mui/material";
import CodeIcon from '@mui/icons-material/Code';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import APIClient from '../util/ApiClient';
import { useUser } from '../context/UserContext';

const CliDownloadPage = () => {
    const { user, loading } = useUser();

    // Show loading state while checking authentication
    if (loading) {
        return (
            <Container maxWidth="md" sx={{ py: 8, display: 'flex', justifyContent: 'center' }}>
                <CircularProgress />
            </Container>
        );
    }

    // If user is not logged in show login prompt
    if (!user) {
        return (
            <Container maxWidth="md" sx={{ py: 8 }}>
                <Card sx={{ p: 6, textAlign: 'center' }}>
                    <LockOutlinedIcon sx={{ fontSize: 64, color: 'primary.main', mb: 2 }} />
                    <Typography variant="h4" fontWeight="bold" gutterBottom>
                        Login Required
                    </Typography>
                    <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
                        You need to be logged in to get your CLI authentication token.
                    </Typography>
                    <Button 
                        variant="contained" 
                        size="large"
                        onClick={() => { window.location.href = APIClient.getGoogleLoginUrl(); }}
                    >
                        Login with Google
                    </Button>
                </Card>
            </Container>
        );
    }

    return (
        <Container maxWidth="md" sx={{ py: 4 }}>
            <Card sx={{ p: 4 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                    <CodeIcon sx={{ fontSize: 40, mr: 2, color: 'primary.main' }} />
                    <Typography variant="h4" fontWeight="bold">
                        OpenDI CLI
                    </Typography>
                </Box>
                
                <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
                    Command-line interface for the OpenDI Model Hub. Manage and interact with model repositories from your terminal.
                </Typography>

                <Divider sx={{ my: 4 }} />

                <Typography variant="h5" fontWeight="bold" gutterBottom>
                    Installation
                </Typography>

                <Typography variant="h6" sx={{ mt: 2, mb: 1 }}>
                    Requirements
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                    Python 3.10 or higher
                </Typography>

                <Typography variant="h6" sx={{ mt: 2, mb: 1 }}>
                    Installation
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                    pipx provides an isolated environment for CLI tools:
                </Typography>
                <Box sx={{ bgcolor: (theme) => theme.palette.mode === 'dark' ? '#333' : '#f5f5f5', color: 'text.primary', p: 2, borderRadius: 1, mb: 3, fontFamily: 'monospace' }}>
                    pipx install opendi
                </Box>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                    <strong>Note:</strong> Install pipx first if needed: <code style={{ fontFamily: 'monospace' }}>pip install pipx</code>
                </Typography>

                <Divider sx={{ my: 4 }} />

                <Typography variant="h5" fontWeight="bold" gutterBottom>
                    Getting Started
                </Typography>

                <Typography variant="h6" sx={{ mt: 3, mb: 1 }}>
                    1. Verify Installation
                </Typography>
                <Box sx={{ bgcolor: (theme) => theme.palette.mode === 'dark' ? '#333' : '#f5f5f5', color: 'text.primary', p: 2, borderRadius: 1, mb: 2, fontFamily: 'monospace' }}>
                    opendi --help
                </Box>

                <Typography variant="h6" sx={{ mt: 3, mb: 1 }}>
                    2. Log In
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    Authenticate with your OpenDI account. This will open your browser to complete login:
                </Typography>
                <Box sx={{ bgcolor: (theme) => theme.palette.mode === 'dark' ? '#333' : '#f5f5f5', color: 'text.primary', p: 2, borderRadius: 1, mb: 2, fontFamily: 'monospace' }}>
                    opendi login
                </Box>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    Your credentials are securely stored in your OS credential manager (Windows Credential Manager, macOS Keychain, or Linux Secret Service).
                </Typography>

                <Typography variant="h6" sx={{ mt: 3, mb: 1 }}>
                    3. Verify Login
                </Typography>
                <Box sx={{ bgcolor: (theme) => theme.palette.mode === 'dark' ? '#333' : '#f5f5f5', color: 'text.primary', p: 2, borderRadius: 1, mb: 2, fontFamily: 'monospace' }}>
                    opendi whoami
                </Box>

                <Typography variant="h6" sx={{ mt: 3, mb: 1 }}>
                    4. Available Commands
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    Once authenticated, you can use:
                </Typography>

                <List>
                    <ListItem>
                        <ListItemText
                            primary={<code>opendi search &lt;query&gt;</code>}
                            secondary="Search for repositories and models on the hub (works with or without login)"
                            primaryTypographyProps={{ fontFamily: 'monospace' }}
                        />
                    </ListItem>
                    <ListItem>
                        <ListItemText
                            primary={<code>opendi list-repos [owner]</code>}
                            secondary="List repositories (yours or by a specific owner)"
                            primaryTypographyProps={{ fontFamily: 'monospace' }}
                        />
                    </ListItem>
                    <ListItem>
                        <ListItemText
                            primary={<code>opendi create-repo &lt;name&gt; [--description &lt;desc&gt;] [--public]</code>}
                            secondary="Create a new repository (private by default)"
                            primaryTypographyProps={{ fontFamily: 'monospace' }}
                        />
                    </ListItem>
                    <ListItem>
                        <ListItemText
                            primary={<code>opendi delete-repo &lt;owner/slug&gt; [--yes]</code>}
                            secondary="Delete a repository (must be the owner)"
                            primaryTypographyProps={{ fontFamily: 'monospace' }}
                        />
                    </ListItem>
                    <ListItem>
                        <ListItemText
                            primary={<code>opendi pull &lt;name&gt;</code>}
                            secondary="Pull a model or resource (coming soon)"
                            primaryTypographyProps={{ fontFamily: 'monospace' }}
                        />
                    </ListItem>
                    <ListItem>
                        <ListItemText
                            primary={<code>opendi push &lt;path&gt;</code>}
                            secondary="Push a model or resource (coming soon)"
                            primaryTypographyProps={{ fontFamily: 'monospace' }}
                        />
                    </ListItem>
                    <ListItem>
                        <ListItemText
                            primary={<code>opendi logout</code>}
                            secondary="Log out and clear stored credentials"
                            primaryTypographyProps={{ fontFamily: 'monospace' }}
                        />
                    </ListItem>
                </List>

                <Divider sx={{ my: 4 }} />

                <Typography variant="h5" fontWeight="bold" gutterBottom>
                    Help & Documentation
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                    Get help on any command:
                </Typography>
                <Box sx={{ bgcolor: (theme) => theme.palette.mode === 'dark' ? '#333' : '#f5f5f5', color: 'text.primary', p: 2, borderRadius: 1, fontFamily: 'monospace' }}>
                    opendi [command] --help
                </Box>
            </Card>
        </Container>
    );
};

export default CliDownloadPage;