//
// COPYRIGHT OpenDI
//

import { NavLink } from "react-router-dom";
import * as React from 'react';
import { useEffect } from 'react';
import { useState } from 'react';
import JsonPatchViewer from "../components/JsonPatchViewer";
import opendiIcon from '../opendi-icon.png';
import API_URL from '../config';
import { useMemo } from 'react';
import {
    Box,
    Button,
    Tabs,
    Tab,
    Typography,
    Link,
    Stack,
    Breadcrumbs,
    Chip,
    Checkbox,
    FormControlLabel,
    FormGroup,
    Card,
    CardContent,
    TextField
} from "@mui/material";
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import { useParams } from "react-router-dom";
import { useDropzone } from "react-dropzone";
import { useCallback } from "react";

const DownloadPage = () => {
    const [uploadStatus, setUploadStatus] = useState(null);
    const [errorMessage, setErrorMessage] = useState("");
    const [open, setOpen] = React.useState(false);

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

    If uuid changes, useEffect runs again (because uuid is in the dependency array).

    */

    useEffect(() => {
        fetch(`${API_URL}/v0/models/${uuid}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(data => {
                // console.log(data);
                setModel(data);
            })
            .catch(error => console.error('There was a problem with the fetch operation:', error));
    }, [uuid]);

    const [commit, setCommit] = useState({})
    useEffect(() => {
        fetch(`${API_URL}/v0/commits/${uuid}`)
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
            })
            .catch(error => console.error('There was a problem with the fetch operation:', error));
    }, [uuid]);


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
                const response = await fetch(`${API_URL}/v0/models/modelVersion/${uuid}/${commit.version - 1}`);
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

        } catch (error) {
            console.error("Error uploading file:", error);
            setUploadStatus("error");
            setErrorMessage(error.message || "Upload failed.");
        }
    }, []);

    const { getRootProps, getInputProps, isDragActive } = useDropzone({ onDrop });

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

    function CollapsedParentLineage() {
        const [lineage, setLineage] = React.useState(null);

        React.useEffect(() => {
            async function fetchLineage() {
                try {
                    const res = await fetch(`${API_URL}/v0/models/lineage/${uuid}`);
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
        }, [uuid]);

        if (!lineage) {
            return;
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
                            href={`/model/${parent.meta?.uuid}`}
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
                    const res = await fetch(`${API_URL}/v0/models/children/${uuid}`);
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
        }, [uuid]);

        if (!children || children.length == 0) {
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
                            href={`/model/${child.meta?.uuid}`}
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
                    <Typography variant="h4" sx={{ pb: 1 }}>   {model.meta ? model.meta.name : ""} </Typography>
                    <Typography variant="subtitle1" sx={{ pb: 2 }}> By: {model && model.meta && model.meta.creator ? model.meta.creator.username : ""} </Typography>

                    <Stack direction="row" spacing={1} sx={{ pb: 8 }}>
                        <Chip label="Tag 1" />
                        <Chip label="Tag 2" />
                        <Chip label="Tag 3" />
                        <Chip label="Tag 4" />
                    </Stack>

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
                        slotProps={{
                            paper: {
                                component: 'form',
                                onSubmit: (event) => {
                                    event.preventDefault();
                                    const formData = new FormData(event.currentTarget);
                                    const formJson = Object.fromEntries(formData.entries());
                                    const email = formJson.email;
                                    console.log(email);
                                    handleClose();
                                },
                            },
                        }}
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

                        {/* Flex container for cards */}
                        <Box sx={{ display: "flex", gap: 2 }}>
                            <Card sx={{ flex: 1 }}>
                                <CardContent>
                                    <h3>Current JSON Model</h3>
                                    <pre>{JSON.stringify(model, null, 2)}</pre>
                                </CardContent>
                            </Card>

                            <Card sx={{ flex: 1 }}>
                                <CardContent>
                                <JsonPatchViewer lastVersionOfModel={lastVersionOfModel} commit={commit} />

                                </CardContent>
                            </Card>
                        </Box>
                    </FormGroup>
                </CustomTabPanel>
                <CustomTabPanel value={value} index={3}>
                    {CollapsedParentLineage()}
                    {ModelChildren()}
                </CustomTabPanel>
            </Box>
        </Box>
    );
};

export default DownloadPage;
