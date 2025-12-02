//
// COPYRIGHT OpenDI
//

import React, { useState, useEffect } from 'react';
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
    Alert,
    TextField,
    InputAdornment,
    IconButton,
    CircularProgress
} from "@mui/material";
import DownloadIcon from '@mui/icons-material/Download';
import CodeIcon from '@mui/icons-material/Code';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import API_URL from '../config';
import { useUser } from '../context/UserContext';

// CLI executable
const CLI_DOWNLOAD_URL = "/opendi-cli.exe";

const CliDownloadPage = () => {
    const { user, loading } = useUser();
    const [token, setToken] = useState('');
    const [copySuccess, setCopySuccess] = useState(false);

    const getCookie = (name) => {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        if (parts.length === 2) {
            const cookieValue = parts.pop().split(';').shift();
            return cookieValue;
        }
        return null;
    };
    
    useEffect(() => {
        if (user) {
            const authToken = sessionStorage.getItem('auth_token');
            if (authToken) {
                setToken(authToken);
            }
        }
    }, [user]);

    const handleCopyToken = () => {
        navigator.clipboard.writeText(token);
        setCopySuccess(true);
        setTimeout(() => setCopySuccess(false), 2000);
    };

    const handleDownloadCLI = () => {
        // Create a link to download the CLI executable
        const link = document.createElement('a');
        link.href = CLI_DOWNLOAD_URL;
        link.download = 'opendi-cli.exe';
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    };

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
                        onClick={() => window.location.href = `${API_URL}/auth/google/login`}
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
                    Download the OpenDI Model Hub command-line interface to manage your models from the terminal.
                </Typography>

                <Button
                    variant="contained"
                    size="large"
                    startIcon={<DownloadIcon />}
                    onClick={handleDownloadCLI}
                    sx={{ mb: 4 }}
                >
                    Download CLI
                </Button>

                <Divider sx={{ my: 4 }} />

                {/* Authentication Token Section */}
                <Typography variant="h5" fontWeight="bold" gutterBottom>
                    Your Authentication Token
                </Typography>
                
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    Use this token to authenticate the CLI with your account:
                </Typography>

                {token ? (
                    <>
                        <TextField
                            fullWidth
                            value={token}
                            InputProps={{
                                readOnly: true,
                                endAdornment: (
                                    <InputAdornment position="end">
                                        <IconButton onClick={handleCopyToken} edge="end">
                                            <ContentCopyIcon />
                                        </IconButton>
                                    </InputAdornment>
                                ),
                                sx: { fontFamily: 'monospace', fontSize: '0.875rem' }
                            }}
                            sx={{ mb: 2 }}
                        />
                        {copySuccess && (
                            <Alert severity="success" sx={{ mb: 2 }}>
                                Token copied to clipboard!
                            </Alert>
                        )}
                    </>
                ) : (
                    <Alert severity="warning" sx={{ mb: 2 }}>
                        Unable to retrieve authentication token. Please try logging out and back in.
                    </Alert>
                )}

                <Divider sx={{ my: 4 }} />

                <Typography variant="h5" fontWeight="bold" gutterBottom>
                    Installation
                </Typography>
                
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    <strong>Step 1:</strong> Download the CLI executable using the button above
                </Typography>

                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    <strong>Step 2:</strong> Place the CLI executable in your working directory
                </Typography>

                <Divider sx={{ my: 4 }} />

                <Typography variant="h5" fontWeight="bold" gutterBottom>
                    Getting Started
                </Typography>

                <Typography variant="h6" sx={{ mt: 3, mb: 1 }}>
                    1. Set the Remote URL
                </Typography>
                <Box sx={{ bgcolor: 'grey.100', p: 2, borderRadius: 1, mb: 2, fontFamily: 'monospace' }}>
                    opendi-cli.exe set-url http://opendi-modelhub.org
                </Box>

                <Typography variant="h6" sx={{ mt: 3, mb: 1 }}>
                    2. Set Your Authentication Token
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                    Copy your token from above and run:
                </Typography>
                <Box sx={{ bgcolor: 'grey.100', p: 2, borderRadius: 1, mb: 2, fontFamily: 'monospace' }}>
                    opendi-cli.exe set-token YOUR_TOKEN_HERE
                </Box>

                <Typography variant="h6" sx={{ mt: 3, mb: 1 }}>
                    3. Start Using the CLI
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    You're all set! Here are some common commands:
                </Typography>

                <List>
                    <ListItem>
                        <ListItemText
                            primary={<code>opendi-cli.exe pull &lt;tag&gt;</code>}
                            secondary="Pull a model from the remote hub"
                            primaryTypographyProps={{ fontFamily: 'monospace' }}
                        />
                    </ListItem>
                    <ListItem>
                        <ListItemText
                            primary={<code>opendi-cli.exe push &lt;tag&gt;</code>}
                            secondary="Push a local model to the remote hub"
                            primaryTypographyProps={{ fontFamily: 'monospace' }}
                        />
                    </ListItem>
                    <ListItem>
                        <ListItemText
                            primary={<code>opendi-cli.exe commit &lt;tag&gt;</code>}
                            secondary="Create a local commit for a model"
                            primaryTypographyProps={{ fontFamily: 'monospace' }}
                        />
                    </ListItem>
                    <ListItem>
                        <ListItemText
                            primary={<code>opendi-cli.exe init &lt;path&gt;</code>}
                            secondary="Initialize a local model from a JSON file"
                            primaryTypographyProps={{ fontFamily: 'monospace' }}
                        />
                    </ListItem>
                    <ListItem>
                        <ListItemText
                            primary={<code>opendi-cli.exe get-commits -r &lt;tag&gt;</code>}
                            secondary="Show commits for a remote model"
                            primaryTypographyProps={{ fontFamily: 'monospace' }}
                        />
                    </ListItem>
                    <ListItem>
                        <ListItemText
                            primary={<code>opendi-cli.exe get-lineage -r &lt;tag&gt;</code>}
                            secondary="Show lineage for a remote model"
                            primaryTypographyProps={{ fontFamily: 'monospace' }}
                        />
                    </ListItem>
                    <ListItem>
                        <ListItemText
                            primary={<code>opendi-cli.exe get-models</code>}
                            secondary="List all local models"
                            primaryTypographyProps={{ fontFamily: 'monospace' }}
                        />
                    </ListItem>
                    <ListItem>
                        <ListItemText
                            primary={<code>opendi-cli.exe clear-token</code>}
                            secondary="Logout (clear stored token)"
                            primaryTypographyProps={{ fontFamily: 'monospace' }}
                        />
                    </ListItem>
                </List>

                <Divider sx={{ my: 4 }} />

                <Typography variant="h5" fontWeight="bold" gutterBottom>
                    Help
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                    For more information and options, run:
                </Typography>
                <Box sx={{ bgcolor: 'grey.100', p: 2, borderRadius: 1, fontFamily: 'monospace' }}>
                    opendi-cli.exe -h
                </Box>
            </Card>
        </Container>
    );
};

export default CliDownloadPage;