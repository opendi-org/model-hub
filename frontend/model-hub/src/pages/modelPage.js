//
// COPYRIGHT OpenDI
//

import * as React from 'react';
import { useEffect } from 'react';
import { useState } from 'react';
import opendiIcon from '../opendi-icon.png';
import APIClient from '../util/ApiClient';
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
            const data = await APIClient.createTransfer(tag, newOwnerEmail);
            setStatus('success');
            setMessage(`Ownership transfer request sent successfully to ${newOwnerEmail}.`);
            onTransferUpdate(data);
            setNewOwnerEmail('');
        } catch (error) {
            console.error('Error during ownership transfer:', error);
            setStatus('error');
            setMessage(error.message || 'An error occurred during the API call.');
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

const ModelPage = () => {
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
                const data = await APIClient.getTransfer(model.addons.tag);
                setPendingTransfer(data);
                setFetchStatus('loaded');
            } catch (error) {
                if (error.message && error.message.includes('404')) {
                    setPendingTransfer(null);
                    setFetchStatus('loaded');
                } else {
                    console.error('Error fetching transfer:', error);
                    setFetchStatus('failed');
                }
            }
        }; 

        fetchTransfer();
    }, [model]); // Re-run effect if model or token changes

    const currentUserID = user ? user.id : null;
    
    // setHeader("Authorization", `Bearer ${jwtToken}`);

    useEffect(() => {
        APIClient.getModelByUUID(uuid)
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
        if (!model?.addons?.tag) return;
        APIClient.getLatestCommitByTag(model.addons.tag)
            .then(data => {
                setCommit(data);
                if (data.version > 0 && !selectedVersion) {
                    setSelectedVersion(data.version);
                }
            })
            .catch(() => setCommit({ version: 0 }));
    }, [model, selectedVersion]);

    // Fetch all commits for the model
    useEffect(() => {
        if (!model?.addons?.tag) return;
        APIClient.getCommitsByTag(model.addons.tag)
            .then(data => setAllCommits(data))
            .catch(() => setAllCommits([]));
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
                const data = await APIClient.getModelVersionByTagAndCommit(model.addons.tag, commit.version - 1);
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
                const data = await APIClient.getModelVersionByTagAndCommit(model.addons.tag, selectedVersion);
                setSelectedVersionModel(data);
                try {
                    const prevData = await APIClient.getModelVersionByTagAndCommit(model.addons.tag, selectedVersion - 1);
                    setPrevVersionModel(prevData);
                } catch {
                    if (selectedVersion === 1) console.log("Fetching version 0 (original state)");
                    setPrevVersionModel("No previous version");
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
            const result = await APIClient.updateModelWithFile(file);
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

        const tag = model?.addons?.tag;
        if (!tag) return;
        try {
            data = await APIClient.getModelPrivacy(tag);
        } catch (err) {
            setError(err.message);
            setIsPrivate(previousIsPrivate);
            return;
        }

        let currentShares = data.shares || [];
        console.log(newIsPublic, currentShares);
        if (newIsPublic) {
            currentShares = currentShares.filter(share => share.level === 'write');
        }

        try {
            await APIClient.updateModelPrivacy(tag, { isPublic: newIsPublic, shares: currentShares });
            setError(null);
        } catch (error) {
            setIsPrivate(previousIsPrivate);
            console.error("An error occurred during the API call:", error);
            setError(error.message || "Failed to update privacy.");
        }
    };


    const [modalIsOpen, setModalIsOpen] = useState(false);
    const [shares, setShares] = useState([]); // [{email: "...", level: "..."}, ...]
    const [searchTerm, setSearchTerm] = useState('');
    // eslint-disable-next-line no-unused-vars
    const [error, setError] = useState(null);
    const privacyTag = model?.addons?.tag;

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
        if (!privacyTag) return;
        let data;
        try {
            data = await APIClient.getModelPrivacy(privacyTag);
        } catch (err) {
            setError(err.message);
            return;
        }
        const newEmail = searchTerm;
        if (!newEmail) {
            setError("Email cannot be empty.");
            return;
        }

        const newShare = { email: newEmail, level: permissionLevel };
        const existingShares = (data.shares || []).filter(share => share.email !== newShare.email);
        const bodyData = { isPublic: isPrivate, shares: [...existingShares, newShare] };

        try {
            await APIClient.updateModelPrivacy(privacyTag, bodyData);
            console.log("Share settings updated successfully.");
            closeModal();
        } catch (error) {
            console.error("Network or other error:", error);
            setError(error.message || "A network error occurred.");
        }
    };

    const handleRemoveShare = async (emailToRemove) => {
        const newSharesList = shares.filter(share => share.email !== emailToRemove);
        const bodyData = { isPublic: isPrivate, shares: newSharesList };

        try {
            await APIClient.updateModelPrivacy(privacyTag, bodyData);
            console.log("Share removed successfully.");
            setShares(newSharesList);
        } catch (error) {
            console.error("API Error removing share:", error);
            setError(error.message || "Failed to remove share.");
        }
    };

    // Fetch data from the API when the modal is opened
    useEffect(() => {
        if (modalIsOpen && privacyTag) {
            const fetchShares = async () => {
                setError(null);
                try {
                    const data = await APIClient.getModelPrivacy(privacyTag);
                    setShares(data.shares || []);
                } catch (err) {
                    setError(err.message);
                }
            };
            fetchShares();
        }
    }, [modalIsOpen, privacyTag]);

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

        // eslint-disable-next-line react-hooks/exhaustive-deps
        React.useEffect(() => {
            if (!model?.addons?.tag) return;
            APIClient.getModelLineage(model.addons.tag)
                .then(data => setLineage(data))
                .catch(error => console.error('Error fetching lineage:', error));
        }, []);

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

        // eslint-disable-next-line react-hooks/exhaustive-deps
        React.useEffect(() => {
            if (!model?.addons?.tag) return;
            APIClient.getModelChildren(model.addons.tag)
                .then(data => setChildren(data))
                .catch(error => console.error('Error fetching children:', error));
        }, []);

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
                            href={`/model/${child?.meta.UUID}`}
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
                    alt="OpenDI Logo – Synergies, Accessibility, Standards"
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

export default ModelPage;