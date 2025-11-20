//
// COPYRIGHT OpenDI
//

import * as React from 'react';
import { useEffect } from 'react';
import { useState } from 'react';
import opendiIcon from '../opendi-icon.png';
import API_URL from '../config';
import { useMemo } from 'react';
import { JSONTree } from 'react-json-tree';
import {
    Alert,
    Box,
    Button,
    Tabs,
    Tab,
    Typography,
    Link,
    Stack,
    Breadcrumbs,
    Chip,
    FormControlLabel,
    FormGroup,
    Card,
    CardContent,
    TextField,
    FormControl,
    InputLabel,
    Select,
    MenuItem,
    Switch,
    IconButton,
    List,
    ListItem,
    ListItemText,
    Modal,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle
} from "@mui/material";
import ShareIcon from '@mui/icons-material/Share'; // Import the Share icon
import DeleteIcon from '@mui/icons-material/Delete';
import { useParams } from "react-router-dom";
import { useDropzone } from "react-dropzone";
import { useCallback } from "react";
import { useUser } from '../context/UserContext';
import JsonDiffViewer from "../components/JsonDiffViewer";


function Ownership({ 
    tag, 
    isOwner, 
    pendingTransfer, 
    onTransferUpdate, 
    fetchStatus 
} = {}) {
    const [status, setStatus] = useState('idle'); // 'idle', 'pending', 'success', 'error'
    const [message, setMessage] = useState('');
    const [newOwnerEmail, setNewOwnerEmail] = useState('');
    const handleNewOwnerEmailChange = (event) => {
        setNewOwnerEmail(event.target.value);
    };

    const handleTransfer = async () => {
        if (!tag || !newOwnerEmail) { 
            setMessage('New owner email must be provided.');
            setStatus('error');
            return;
        }

        setStatus('pending');
        setMessage('Sending ownership transfer request...');

        try {
            const url = `${API_URL}/v0/models/transfer/${tag}?owner=${newOwnerEmail}`;

            const response = await fetch(url, {
                method: 'POST',
                credentials: 'include', 
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({})
            });

            if (response.ok) {
                const data = await response.json();
                setStatus('success');
                setMessage(`Ownership transfer request sent successfully to ${newOwnerEmail}.`);
                // The actual ToUserID is returned, but we show the email for user confirmation
                onTransferUpdate(data); 
                setNewOwnerEmail(''); 
            } else {
                const errorData = await response.json();
                setStatus('error');
                setMessage(errorData.error || 'An unexpected error occurred.');
            }
        } catch (error) {
            console.error('Error during ownership transfer:', error);
            setStatus('error');
            setMessage('An error occurred during the API call.');
        }
    };

    const handleAcceptDecline = async (accept) => {
        setStatus('pending');
        const action = accept ? 'Accepting' : 'Declining';
        setMessage(`${action} ownership transfer request...`);

        try {
            // API Endpoint: DELETE /v0/models/transfer/{tag}?accept={boolean}
            const acceptValue = accept ? 'true' : 'false';
            const url = `${API_URL}/v0/models/transfer/${tag}?accept=${acceptValue}`;

            const response = await fetch(url, {
                method: 'DELETE', 
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json'
                },
            });

            if (response.ok) {
                setStatus('success');
                const successMsg = accept 
                    ? 'Ownership transfer accepted successfully.' 
                    : 'Ownership transfer declined successfully.';
                setMessage(successMsg);
                onTransferUpdate(null);
            } else {
                // Read the error body for a specific message
                const errorData = await response.json();
                setStatus('error');
                // Use the error message from the API or a fallback
                setMessage(errorData.error || `Error (${response.status}): Could not complete the action.`);
            }
        } catch (error) {
            console.error('Error during accept/decline:', error);
            setStatus('error');
            setMessage('A network error occurred during the accept/decline API call.');
        }
    };

    return (
        <div>
            <Stack spacing={2}>
                {isOwner && (
                    <>
                        {/* The TextField and its handler now live here, in the component that owns the state */}
                        <TextField
                            label="Transfer Email"
                            variant="outlined"
                            value={newOwnerEmail} 
                            onChange={handleNewOwnerEmailChange} 
                            sx={{ width: "30%" }}
                        />
                        
                        {/* The button that triggers the API call */}
                        <Button 
                            variant="outlined" 
                            sx={{ width: "30%" }} 
                            onClick={handleTransfer} 
                            disabled={status === 'pending' || !!pendingTransfer}
                        >
                            Transfer Ownership
                        </Button>
                    </>
                )}
                
                {!!pendingTransfer && (
                    // Display the ToUserID (which we temporarily set to email on success, or the ID from GET)
                    // <p>A transfer request is currently pending to user ID: <strong>{pendingTransfer.toUserID}</strong></p>
                    <p>A transfer request is currently pending</p>
                )}

                {fetchStatus === 'loading' && <p>Checking for pending transfer...</p>}
                {fetchStatus === 'failed' && <p style={{ color: 'red' }}>Could not check for pending transfer.</p>}

                {/* Display the status and messages */}
                {status === 'pending' && <p>{message}</p>}
                {status === 'success' && <p style={{ color: 'green' }}>{message}</p>}
                {status === 'error' && <p style={{ color: 'red' }}>{message}</p>}
            </Stack>
        </div>
    );
}

const MemoizedOwnership = React.memo(Ownership);

const DownloadPage = () => {
    const { user } = useUser();
    const [uploadStatus, setUploadStatus] = useState(null);
    const [errorMessage, setErrorMessage] = useState("");
    const [open, setOpen] = React.useState(false);

    // New state for all commits and selected version
    const [allCommits, setAllCommits] = useState([]);
    const [selectedVersion, setSelectedVersion] = useState(null);
    const [selectedVersionModel, setSelectedVersionModel] = useState(null);
    const [prevVersionModel, setPrevVersionModel] = useState(null);

    const handleClickOpen = () => {
        setOpen(true);
    };

    const handleClose = () => {
        setOpen(false);
    };

    const cdm = {
        creator: 'No CDM loaded'
    };

    //hook that extract route parameters from URL
    const { uuid } = useParams();
    // console.log( uuid );
    

    //useState returns an array of two elements that contain a state variable and a method to change the variable (and in doing so, re-render)
    const [model, setModel] = useState({})

    /*
    Runs after the component renders.

    Fetches model data from an API.

    Updates model using setModel(data).

    If tag changes, useEffect runs again (because tag is in the dependency array).

    */

    const [pendingTransfer, setPendingTransfer] = useState(null);
    const [fetchStatus, setFetchStatus] = useState('idle'); // 'idle', 'loading', 'loaded', 'failed'

    useEffect(() => {
        // Don't run if we don't have a token or tag
        if (!model?.addons?.tag) return;

        const fetchTransfer = async () => {
            setFetchStatus('loading');
            try {
                const url = `${API_URL}/v0/models/transfer/${model?.addons?.tag}`;
                const response = await fetch(url, {
                    method: 'GET',
                    credentials: 'include',
                    headers: {
                        'Content-Type': 'application/json'
                    }
                });

                if (response.ok) {
                    const data = await response.json();
                    setPendingTransfer(data);
                    setFetchStatus('loaded');
                } else if (response.status === 404) {
                    // This is expected if no transfer is pending
                    setPendingTransfer(null);
                    setFetchStatus('loaded');
                } else {
                    // Handle other errors (401, 403, 500, etc.)
                    const errorData = await response.json();
                    console.error('Error fetching transfer:', errorData.error);
                    setFetchStatus('failed');
                }
            } catch (error) {
                console.error('Network error fetching transfer:', error);
                setFetchStatus('failed');
            }
        }; 

        fetchTransfer();
    }, [model]); // Re-run effect if model or token changes

    const currentUserID = user ? user.id : null;
    
    // setHeader("Authorization", `Bearer ${jwtToken}`);

    useEffect(() => {
        fetch(`${API_URL}/v0/models/${uuid}`, { credentials: 'include' })
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(data => {
                setModel(data);
                if (data.addons && data.addons.hasOwnProperty('isPublic')) {
                    setIsPrivate(data.addons.isPublic);
                }
            })
            .catch(error => console.error('There was a problem with the fetch operation:', error));
    }, [uuid]);

    const [commit, setCommit] = useState({})
    useEffect(() => {
        fetch(`${API_URL}/v0/models/commits/latest/${model?.addons?.tag}`, { credentials: 'include' })
            .then(response => {
                if (response.status === 404) {
                    return { version: 0 }; // Exit early if not found
                }
    
                if (!response.ok) {
                    throw new Error('Network response was not ok for getting latest commit');
                }
                return response.json();
            })
            .then(data => {
                    setCommit(data); // Set commit data if the response was valid
                    if (data.version > 0 && !selectedVersion) {
                        setSelectedVersion(data.version);
                    }
            })
            .catch(error => console.error('There was a problem with the fetch operation:', error));
    }, [model, selectedVersion]);

    // Fetch all commits for the model
    useEffect(() => {
        if (!model?.addons?.tag) return;
        
        fetch(`${API_URL}/v0/models/commits/${model?.addons?.tag}`, { credentials: 'include' })
            .then(response => {
                if (response.status === 404) {
                    setAllCommits([]); // Set empty array if no commits found
                    return [];
                }
                if (!response.ok) {
                    throw new Error('Network response was not ok for getting commits');
                }
                return response.json();
            })
            .then(data => {
                setAllCommits(data);
            })
            .catch(error => console.error('Error fetching commits:', error));
    }, [model]);

    // Get previous version of model
    //note that for development purposes, in react strict mode, useEffect invokes twice. 
    const [lastVersionOfModel, setLastVersionOfModel] = useState(null);

    useEffect(() => {
        const fetchData = async () => {
            if (Object.keys(model).length === 0 || Object.keys(commit).length === 0) {
                return;
            }

            if (commit['version'] === 0) {
                setLastVersionOfModel("No previous version");
                return;
            }

            try {
                const response = await fetch(`${API_URL}/v0/models/modelVersion/${model?.addons?.tag}/${commit.version - 1}`);
                if (!response.ok) {
                    throw new Error('Network response was not ok for getting model version');
                }
                const data = await response.json();
                console.log(commit['diff'])
                setLastVersionOfModel(data);
            } catch (error) {
                console.error('There was a problem with the fetch operation:', error);
            }
        };

        fetchData();
    }, [model, commit]); // Re-run when model or commit changes
    
    // Effect for fetching the selected version model
    useEffect(() => {
        if (!model?.addons?.tag || !selectedVersion) return;
        
        const fetchSelectedVersion = async () => {
            try {
                const response = await fetch(`${API_URL}/v0/models/modelVersion/${model?.addons?.tag}/${selectedVersion}`);
                if (!response.ok) {
                    throw new Error('Network response was not ok for getting selected model version');
                }
                const data = await response.json();
                setSelectedVersionModel(data);
                
                // Always try to fetch the previous version, even for version 1
                const prevResponse = await fetch(`${API_URL}/v0/models/modelVersion/${model?.addons?.tag}/${selectedVersion - 1}`);
                if (!prevResponse.ok) {
                    // For version 1, we need to handle the special case where version 0 might not be directly accessible
                    if (selectedVersion === 1) {
                        console.log("Fetching version 0 (original state)");
                        // The backend should reconstruct version 0 from the version 1 diff
                    } else {
                        console.error('Error fetching previous version:', prevResponse.statusText);
                    }
                    setPrevVersionModel("No previous version");
                } else {
                    const prevData = await prevResponse.json();
                    setPrevVersionModel(prevData);
                }
            } catch (error) {
                console.error('Error fetching model versions:', error);
            }
        };
        
        fetchSelectedVersion();
    }, [model, selectedVersion]);

    // Handle version selection
    const handleVersionChange = (event) => {
        setSelectedVersion(Number(event.target.value));
    };

    // Find the selected commit from allCommits
    const selectedCommit = useMemo(() => {
        if (!selectedVersion || !allCommits.length) return null;
        return allCommits.find(c => c.version === selectedVersion) || null;
    }, [selectedVersion, allCommits]);

    async function getCDM() {
        // console.log("Creator: " + cdm.creator);
        try {
            const json = model;
            cdm.creator = json.meta.creator.Username;
            // console.log("Creator: " + cdm.creator);

            const jsonString = JSON.stringify(json, null, 2);
            const dataUri = "data:application/json;charset=utf-8," + encodeURIComponent(jsonString);

            const link = document.createElement("a");
            link.href = dataUri;
            link.download = "model.json";
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);

            //   document.getElementById("cdm").innerHTML = cdm.creator;
        } catch (error) {
            console.error(error.message);
        }
    }

    const onDrop = useCallback(async (acceptedFiles) => {
        console.log(acceptedFiles);

        const file = acceptedFiles[0];

        try {
            // const fileText = await file.text();
            const response = await fetch(`${API_URL}/v0/models`, {
                method: "PUT",
                credentials: 'include',
                headers: {
                    "Content-Type": "application/json"
                },
                body: file
            });

            if (!response.ok) {
                throw new Error(`Upload failed: ${response.statusText}`);
            }

            const result = await response.json();
            console.log("Updated success:", result);

            setUploadStatus("success");
            setErrorMessage("");
            handleClose();

        } catch (error) {
            console.error("Error uploading file:", error);
            setUploadStatus("error");
            setErrorMessage(error.message || "Update failed.");
            handleClose();
        }
    }, []);

    const { getRootProps, getInputProps, isDragActive } = useDropzone({ onDrop });

    const [isPrivate, setIsPrivate] = useState(false); //Privacy Toggle

    const handlePrivacyToggle = async (event) => {
        let data;
        const newIsPublic = event.target.checked;
        const previousIsPrivate = isPrivate; // Store the old value to revert to

        setIsPrivate(newIsPublic); // Optimistically update the state

        try {
            const response = await fetch(apiEndpoint, {
                method: 'GET',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to fetch share settings.');
            }

            data = await response.json();
        } catch (err) {
            setError(err.message);
            setIsPrivate(previousIsPrivate); // Revert on GET failure
            return; // <-- Stop execution
        }

        let currentShares = data.shares || [];
        console.log(newIsPublic, currentShares);
        if(newIsPublic) {
            currentShares = currentShares.filter(share => share.level === 'write');
        }

        try {
            const response = await fetch(`${API_URL}/v0/models/privacy/${model?.addons?.tag}`, {
                method: 'PUT',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ 
                    isPublic: newIsPublic,
                    shares: currentShares
                 })
            });

            if (!response.ok) {
                // Revert the state change if the API call fails
                setIsPrivate(previousIsPrivate); // Revert on PUT failure
                const error = await response.json();
                console.error("Failed to update privacy settings:", error.error);
                setError(error.error || "Failed to update privacy.");
            } else {
                setError(null); // Clear any previous errors on success
            }
        } catch (error) {
            // Revert the state change if the API call fails
            setIsPrivate(previousIsPrivate); // Revert on PUT network error
            console.error("An error occurred during the API call:", error);
            setError("An network error occurred.");
        }
    };


    const [modalIsOpen, setModalIsOpen] = useState(false);
    const [shares, setShares] = useState([]); // [{email: "...", level: "..."}, ...]
    const [searchTerm, setSearchTerm] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState(null);

    const apiEndpoint = `${API_URL}/v0/models/privacy/${model?.addons?.tag}`;

    const style = {
        position: 'absolute',
        top: '50%',
        left: '50%',
        transform: 'translate(-50%, -50%)',
        width: 400,
        bgcolor: 'background.paper',
        border: '2px solid #000',
        boxShadow: 24,
        p: 4,
    };

    const handleShareClick = () => {
        setModalIsOpen(true);
    };

    const closeModal = () => {
        setModalIsOpen(false);
        setSearchTerm('');
    };

    const handleSearchChange = (event) => {
        setSearchTerm(event.target.value);
    };

    const [permissionLevel, setPermissionLevel] = useState('read');

    const handlePermissionChange = (event) => {
        setPermissionLevel(event.target.value);
    };

    const handleSave = async () => {
        let data;

        try {
            const response = await fetch(apiEndpoint, {
                method: 'GET',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Failed to fetch share settings.');
            }


            data = await response.json();
        } catch (err) {
            setError(err.message);
            return;
        }
        const url = `${API_URL}/v0/models/privacy/${model?.addons?.tag}`;
        // searchTerm is now the new user's email
        const newEmail = searchTerm; 
        
        if (!newEmail) {
            console.error("Email cannot be empty.");
            setError("Email cannot be empty.");
            return;
        }

        const newShare = {
            email: newEmail, // Use email field
            level: permissionLevel
        }
        
        // Filter out the new share if the email already exists to prevent duplicates
        const existingShares = (data.shares || []).filter(share => share.email !== newShare.email);
        
        const bodyData = {
            'isPublic': isPrivate,
            'shares': [ 
                ...existingShares,
                newShare
            ]
        };
        

        try {
            const response = await fetch(url, {
                method: 'PUT',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(bodyData)
            });

            if (response.ok) {
                console.log("Share settings updated successfully.");
                closeModal();
            } else {
                const errorData = await response.json();
                console.error("API Error:", errorData.error);
                setError(errorData.error || "Failed to save share settings.");
                // Handle specific error messages based on status codes (e.g., show a user-friendly message)
            }
        } catch (error) {
            console.error("Network or other error:", error);
            setError("A network error occurred.");
            // Handle network errors
        }
    };

    const handleRemoveShare = async (emailToRemove) => {
        // Create the new list of shares by filtering out the user to remove
        const newSharesList = shares.filter(share => share.email !== emailToRemove);
        
        const bodyData = {
            'isPublic': !isPrivate, // Use the current state of the public/private toggle
            'shares': newSharesList
        };

        try {
            const response = await fetch(apiEndpoint, {
                method: 'PUT',
                credentials: 'include',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(bodyData)
            });

            if (response.ok) {
                console.log("Share removed successfully.");
                // Update the local state to reflect the change immediately
                setShares(newSharesList); 
            } else {
                const errorData = await response.json();
                console.error("API Error removing share:", errorData.error);
                // Optionally, set an error message to display to the user
                setError("Failed to remove share."); 
            }
        } catch (error) {
            console.error("Network or other error:", error);
            // Optionally, set an error message
            setError("A network error occurred.");
        }
    };

    // Fetch data from the API when the modal is opened
    useEffect(() => {
        if (modalIsOpen) {
            const fetchShares = async () => {
                setIsLoading(true);
                setError(null);
                try {
                    const response = await fetch(apiEndpoint, {
                        method: 'GET',
                        credentials: 'include',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                    });

                    if (!response.ok) {
                        const errorData = await response.json();
                        throw new Error(errorData.error || 'Failed to fetch share settings.');
                    }

                    const data = await response.json();
                    setShares(data.shares || []); // data.shares now contains email
                } catch (err) {
                    setError(err.message);
                } finally {
                    setIsLoading(false);
                }
            };
            fetchShares();
        }
    }, [modalIsOpen, apiEndpoint]);

    function displayUploadMenu() {
        return (
            <Box
                {...getRootProps()}
                sx={{
                    border: "2px dashed gray",
                    borderRadius: 2,
                    p: 4,
                    height: '5em',
                    alignContent: 'center',
                    textAlign: "center",
                    cursor: "pointer",
                    backgroundColor: isDragActive ? "lightblue" : "transparent",
                    transition: "background-color 0.2s ease-in-out",
                    "&:hover": {
                        backgroundColor: "lightgray",
                    },
                }}
            >
                <input {...getInputProps()} />
                <Typography variant="body1">
                    {isDragActive ? "Drop the files here..." : "Drag 'n' drop some files here, or click to select files"}
                </Typography>
            </Box>
        );
    }

    const breadcrumbs = [
        <Link underline="hover" key="1" color="inherit" href="/">
            MUI
        </Link>,
        <Link
            underline="hover"
            key="2"
            color="inherit"
            href="/material-ui/getting-started/installation/"
        >
            Core
        </Link>,
        <Typography key="3" sx={{ color: 'text.primary' }}>
            Breadcrumb
        </Typography>,
    ];

    function CustomTabPanel(props) {
        const { children, value, index, ...other } = props;

        return (
            <div
                role="tabpanel"
                hidden={value !== index}
                id={`simple-tabpanel-${index}`}
                aria-labelledby={`simple-tab-${index}`}
                {...other}
            >
                {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
            </div>
        );
    }

    const [value, setValue] = React.useState(0);

    const handleChange = (event, newValue) => {
        setValue(newValue);
    };

    const ownerID = model?.addons?.ownerID;
    const isOwner = currentUserID && ownerID && currentUserID === ownerID;
    
    // The tab should only show if you are the owner OR the target of a transfer
    const showOwnershipTab = isOwner;

    function CollapsedParentLineage() {
        const [lineage, setLineage] = React.useState(null);

        React.useEffect(() => {
            async function fetchLineage() {
                try {
                    const res = await fetch(`${API_URL}/v0/models/lineage/${model?.addons?.tag}`, { credentials: 'include' });
                    if (!res.ok) {
                        throw new Error('Failed to fetch lineage');
                    }
                    const data = await res.json();
                    setLineage(data);
                } catch (error) {
                    console.error('Error fetching lineage:', error);
                }
            }
            fetchLineage();
        }, [model]);

        if (!lineage) {
            return;
        }

        if (lineage.length === 0) {
            return (
                <div role="presentation">
                    <Typography variant="h6" gutterBottom>
                        Parent Lineage
                    </Typography>
                    <Typography variant="body1">
                        No parent lineage present. This is a root model.
                    </Typography>
                </div>
            );
        }

        return (
            <div role="presentation">
                <Typography variant="h6" gutterBottom>
                    Parent Lineage
                </Typography>
                <Breadcrumbs maxItems={4} separator="›" aria-label="breadcrumb" sx={{ mb: "2em" }}>
                    {lineage.map((parent, index) => (
                        <Link
                            key={parent.id || `lineage-${index}`}
                            underline="hover"
                            color="inherit"
                            href={`/model/${parent?.addons?.tag}`}
                        >
                            {parent.meta ? parent.meta.name : parent.name}
                        </Link>
                    ))}
                </Breadcrumbs>
            </div>
        );
    }

    function ModelChildren() {
        const [children, setChildren] = React.useState(null);

        React.useEffect(() => {
            async function fetchChildren() {
                try {
                    const res = await fetch(`${API_URL}/v0/models/children/${model?.addons?.tag}`, { credentials: 'include' });
                    if (!res.ok) {
                        throw new Error('Failed to fetch children');
                    }
                    const data = await res.json();
                    setChildren(data);
                } catch (error) {
                    console.error('Error fetching children:', error);
                }
            }
            fetchChildren();
        }, [model]);

        if (!children || children.length === 0) {
            return;
        }

        return (
            <div role="presentation">
                <Typography variant="h6" gutterBottom>
                    Children
                </Typography>
                <Box sx={{ display: 'flex', gap: '1em' }}>
                    {children.map((child, index) => (
                        <Link
                            key={child.id || `children-${index}`}
                            underline="hover"
                            color="gray"
                            href={`/model/${child?.addons?.tag}`}
                        >
                            {child.meta ? child.meta.name : child.name}
                        </Link>
                    ))}
                </Box>
            </div>
        );
    }

    return (
        <Box sx={{ display: "flex", flexDirection: "column", p: 3 }}>
            <Stack spacing={2} sx={{ p: 3, pb: 0 }}>
                <Breadcrumbs separator="›" aria-label="breadcrumb">
                    {breadcrumbs}
                </Breadcrumbs>
            </Stack>


            <Box sx={{ display: "flex", flexDirection: "row", p: 3 }}>
                <Box
                    component="img"
                    src={opendiIcon}
                    alt="Description of image"
                    sx={{ width: '20em', height: 'auto' }}
                />

                <Box sx={{ display: "flex", flexDirection: "column", p: 3, flex: 1 }}>



                    {/* Show Success Alert */}
                    {uploadStatus === "success" && (
                        <Alert severity="success" sx={{ mb: 2 }} onClose={() => setUploadStatus(null)}>
                            File uploaded successfully! Please refresh the page to see the changes. 
                        </Alert>
                    )}
    
                    {/* Show Error Alert */}
                    {uploadStatus === "error" && (
                        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setUploadStatus(null)}>
                            {errorMessage}
                        </Alert>
                    )}



                    <Typography variant="h4" sx={{ pb: 1 }}>{model.meta ? model.meta.name : ""}</Typography>
                    <Typography variant="subtitle1" sx={{ pb: 2 }}> By: {model && model.meta && model.meta.creator ? model.meta.creator.username : ""} </Typography>



                    <Stack direction="row" spacing={1} sx={{ pb: 8 }}>
                    <Chip label="Tag 1" />
                    <Chip label="Tag 2" />
                    <Chip label="Tag 3" />
                    <Chip label="Tag 4" />
                    </Stack>

                    {/*  Privacy Toggle  */}
                    <div>
                        <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                            {isOwner && (<FormGroup>
                                <FormControlLabel
                                    control={
                                        <Switch
                                            checked={isPrivate}
                                            // onChange={(event) => setIsPrivate(event.target.checked)}
                                            onChange={handlePrivacyToggle}
                                            
                                        />
                                    }
                                    label={isPrivate ? "Privacy: Public" : "Privacy: Private"}
                                />
                            </FormGroup>)}

                            {isOwner && (<IconButton
                                aria-label="share"
                                onClick={handleShareClick}
                                sx={{ ml: 2 }}
                            >
                                <ShareIcon />
                            </IconButton>)}
                        </Box>

                        <Modal
                            open={modalIsOpen}
                            onClose={closeModal}
                            // contentLabel="Share Modal"
                        >
                            <Box sx={style}>
                                <Typography variant="h6" component="h2" sx={{ mb: 2 }}>
                                    Share Settings
                                </Typography>
                                <TextField
                                    fullWidth
                                    label="Search users"
                                    variant="outlined"
                                    value={searchTerm}
                                    onChange={handleSearchChange}
                                    sx={{ mb: 2 }}
                                />
                                <FormControl sx={{ minWidth: 120 }}>
                                    <InputLabel>Permissions</InputLabel>
                                    <Select
                                        value={permissionLevel}
                                        label="Permissions"
                                        onChange={handlePermissionChange}
                                    >
                                        {!isPrivate && (<MenuItem value="read">Read</MenuItem>)}
                                        <MenuItem value="write">Read/Write</MenuItem>
                                    </Select>
                                </FormControl>
                                <div>
                                    <div>Current Shares</div>
                                    <List>
                                        {shares.length > 0 ? (
                                        shares.map((share, index) => (
                                            <ListItem 
                                                key={index}
                                                secondaryAction={
                                                    <IconButton edge="end" aria-label="delete" onClick={() => handleRemoveShare(share.email)}>
                                                        <DeleteIcon />
                                                    </IconButton>
                                                }
                                            >
                                            <ListItemText
                                                primary={share.email}
                                                secondary={`Level: ${share.level}`}
                                            />
                                            </ListItem>
                                        ))
                                        ) : (
                                            <Typography>No shared users found.</Typography>
                                        )}
                                    </List>
                                </div>
                                <Box sx={{ mt: 2, display: 'flex', justifyContent: 'flex-end', gap: 1 }}>
                                    <Button onClick={handleSave}>Share</Button>
                                    <Button onClick={closeModal}>Close</Button>
                                </Box>
                            </Box>
                        </Modal>
                    </div>
                    
                    <Button
                        variant="outlined"
                        sx={{ width: "30%" }}
                        onClick={getCDM}
                        >
                        Download
                    </Button>
                    <Button
                        variant="outlined"
                        sx={{ width: "30%", mt: '1em' }}
                        onClick={handleClickOpen}
                    >
                        Update
                    </Button>
                    <Dialog
                    open={open}
                    onClose={handleClose}
                    >
                    <DialogTitle>Update Model</DialogTitle>
                    <DialogContent>
                        {displayUploadMenu()}
                    </DialogContent>
                    <DialogActions>
                        <Button onClick={handleClose}>Cancel</Button>
                    </DialogActions>
                    </Dialog>
                </Box>
            </Box>

            <Box sx={{ p: 3 }}>
                <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
                    <Tabs value={value} onChange={handleChange} aria-label="basic tabs example">
                        <Tab label="Overview" />
                        <Tab label="Documentation" />
                        <Tab label="Commit Diff" />
                        <Tab label="Fork Info" />
                        {showOwnershipTab && <Tab label="Ownership" />}
                    </Tabs>
                </Box>
                <CustomTabPanel value={value} index={0}>
                    {model.meta ? model.meta.summary : ""}
                </CustomTabPanel>
                <CustomTabPanel value={value} index={1}>
                    {model.meta && model.meta.documentation ? model.meta.documentation.content : ""}
                </CustomTabPanel>
                <CustomTabPanel value={value} index={2}>
                    <FormGroup>
                        {/* Version selector dropdown */}
                        <Box sx={{ mb: 3 }}>
                            <FormControl fullWidth>
                                <InputLabel>Select Diff</InputLabel>
                                <Select
                                    value={selectedVersion || ''}
                                    onChange={handleVersionChange}
                                    label="Select Diff"
                                >
                                    {allCommits.map((commitItem) => (
                                        <MenuItem key={commitItem.version} value={commitItem.version}>
                                            Diff {commitItem.version} - {new Date(commitItem.CreatedAt).toLocaleString()}
                                        </MenuItem>
                                    ))}
                                </Select>
                            </FormControl>
                        </Box>

                        {/* Flex container for cards */}
                        <Box sx={{ display: "flex", gap: 2 }}>
                            <Card sx={{ flex: 1 }}>
                                <CardContent>
                                <h3>Current JSON Model</h3>
                                    <JSONTree
                                    data={selectedVersionModel || model}
                                    shouldExpandNodeInitially={() => true}
                                    />
                                </CardContent>
                            </Card>

                            <Card sx={{ flex: 1 }}>
                                <CardContent>
                                <h3>Previous JSON Model</h3>
                                <JsonDiffViewer 
                                    lastVersionOfModel={selectedVersion === commit.version ? lastVersionOfModel : prevVersionModel} 
                                    commit={selectedCommit || commit} 
                                />
                                </CardContent>
                            </Card>
                        </Box>
                    </FormGroup>
                </CustomTabPanel>
                <CustomTabPanel value={value} index={3}>
                    {CollapsedParentLineage()}
                    {ModelChildren()}
                </CustomTabPanel>
                {showOwnershipTab && (
                    <CustomTabPanel value={value} index={4}>
                        <MemoizedOwnership 
                            tag={model?.addons?.tag} 
                            isOwner={isOwner}
                            pendingTransfer={pendingTransfer}
                            // Pass the setter so the child can update the parent's state
                            onTransferUpdate={setPendingTransfer} 
                            fetchStatus={fetchStatus}
                        />
                    </CustomTabPanel>
                )}
            </Box>
        </Box>
    );
};

export default DownloadPage;