//
// COPYRIGHT OpenDI
//

import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Box,
    Button,
    Card,
    Alert,
    Typography,
    Container,
    CircularProgress
} from "@mui/material";
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';
import API_URL from '../config';
import { useDropzone } from "react-dropzone";
import { useUser } from '../context/UserContext';

const UploadPage = () => {
    const [uploadStatus, setUploadStatus] = useState(null);
    const [errorMessage, setErrorMessage] = useState("");
    const { user, loading } = useUser();
    const navigate = useNavigate();
    const currentUserID = user ? user.id : null;

    // Drag and drop functionality
    const onDrop = useCallback(async (acceptedFiles) => {
        console.log(acceptedFiles);

        const file = acceptedFiles[0];

        try {
            const fileText = await file.text();

            let fileData;
            try {
                fileData = JSON.parse(fileText);
            } catch (parseError) {
                throw new Error("Invalid JSON file format.");
            }

            fileData.id = null
            fileData.meta = fileData.meta || {};
            fileData.meta.creator = fileData.meta.creator || {}; 
            fileData.meta.creator.email = user.email;

            const response = await fetch(`${API_URL}/v0/models`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                credentials: "include",
                body: JSON.stringify(fileData)
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.error || `Upload failed: ${response.statusText}`);
            }

            const result = await response.json();
            console.log("Upload success:", result);

            setUploadStatus("success");
            setErrorMessage("");

        } catch (error) {
            console.error("Error uploading file:", error);
            setUploadStatus("error");
            setErrorMessage(error.message || "Upload failed.");
        }
    }, [user]);

    const { getRootProps, getInputProps, isDragActive } = useDropzone({ 
        onDrop,
        accept: {
            'application/json': ['.json']
        },
        multiple: false
    });

    // Show loading state while checking authentication
    if (loading) {
        return (
            <Container maxWidth="md" sx={{ py: 8, display: 'flex', justifyContent: 'center' }}>
                <CircularProgress />
            </Container>
        );
    }

    // If user is not logged in, show login prompt
    if (!user) {
        return (
            <Container maxWidth="md" sx={{ py: 8 }}>
                <Card sx={{ p: 6, textAlign: 'center' }}>
                    <LockOutlinedIcon sx={{ fontSize: 64, color: 'primary.main', mb: 2 }} />
                    <Typography variant="h4" fontWeight="bold" gutterBottom>
                        Login Required
                    </Typography>
                    <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
                        You need to be logged in to upload models to the OpenDI Model Hub.
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

    // If user is logged in, show upload functionality
    return (
        <Container maxWidth="lg" sx={{ py: 4 }}>
            <Box
                sx={{
                    display: 'flex',
                    flexDirection: { xs: 'column', md: 'row' },
                    gap: 3,
                    alignItems: 'flex-start',
                }}
            >
                <Box sx={{ flex: 1, minWidth: 0 }}>
                    <Card sx={{ p: 4 }}>
                        <Typography variant="h4" fontWeight="bold" gutterBottom>
                            Upload Model
                        </Typography>

                        <Typography variant="body1" color="text.secondary" sx={{ mb: 3 }}>
                            Upload a JSON file containing your Causal Decision Model
                        </Typography>

                        {/* Show Success Alert */}
                        {uploadStatus === "success" && (
                            <Alert severity="success" sx={{ mb: 2 }} onClose={() => setUploadStatus(null)}>
                                File uploaded successfully!
                            </Alert>
                        )}

                        {/* Show Error Alert */}
                        {uploadStatus === "error" && (
                            <Alert severity="error" sx={{ mb: 2 }} onClose={() => setUploadStatus(null)}>
                                {errorMessage}
                            </Alert>
                        )}

                        <Box
                            {...getRootProps()}
                            sx={{
                                border: "2px dashed",
                                borderColor: isDragActive ? "primary.main" : "grey.400",
                                borderRadius: 2,
                                p: 6,
                                textAlign: "center",
                                cursor: "pointer",
                                backgroundColor: isDragActive ? "action.hover" : "background.paper",
                                transition: "all 0.2s ease-in-out",
                                "&:hover": {
                                    backgroundColor: "action.hover",
                                    borderColor: "primary.main",
                                },
                            }}
                        >
                            <input {...getInputProps()} />
                            <Typography variant="h6" gutterBottom>
                                {isDragActive ? "Drop the JSON file here" : "Drag & drop a JSON file here"}
                            </Typography>
                            <Typography variant="body2" color="text.secondary">
                                or click to select a file
                            </Typography>
                            <Typography variant="caption" display="block" sx={{ mt: 2 }} color="text.secondary">
                                Only .json files are accepted
                            </Typography>
                        </Box>
                    </Card>
                </Box>

                <Box sx={{ flexShrink: 0, width: { xs: '100%', md: 280 } }}>
                    <Card sx={{ p: 3 }}>
                        <Typography variant="overline" color="text.secondary" display="block" sx={{ mb: 2 }}>
                            Resources
                        </Typography>
                        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                            <Box>
                                <Typography variant="subtitle2" gutterBottom>
                                    API Documentation
                                </Typography>
                                <Button
                                    component="a"
                                    href="https://opendi-org.github.io/api-specification/"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    variant="outlined"
                                    size="small"
                                    endIcon={<OpenInNewIcon />}
                                    fullWidth
                                >
                                    View API Spec
                                </Button>
                            </Box>
                            <Box>
                                <Typography variant="subtitle2" gutterBottom>
                                    Authoring Tool
                                </Typography>
                                <Button
                                    component="a"
                                    href="https://opendi-org.github.io/cdd-authoring-tool/"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    variant="outlined"
                                    size="small"
                                    endIcon={<OpenInNewIcon />}
                                    fullWidth
                                >
                                    Open Authoring Tool
                                </Button>
                            </Box>
                        </Box>
                    </Card>
                </Box>
            </Box>
        </Container>
    );
};

export default UploadPage;