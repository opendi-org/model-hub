//
// COPYRIGHT OpenDI
//

import { NavLink } from "react-router-dom";
import { useEffect } from 'react';
import { useState } from 'react';
import {
    Box,
    Button,
    Card,
    Tabs,
    Tab,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    Alert,
    Typography
} from "@mui/material";
import API_URL from '../config';
import { useDropzone } from "react-dropzone";
import { useCallback } from "react";

const UploadPage = () => {
    const [uploadStatus, setUploadStatus] = useState(null);
    const [errorMessage, setErrorMessage] = useState("");

    //drag and drop functionality
    const onDrop = useCallback(async (acceptedFiles) => {
        console.log(acceptedFiles);

        const file = acceptedFiles[0];

        try {
            // const fileText = await file.text();

            const fileText = await file.text(); // Read file contents as text

            let fileData;
            try {
                fileData = JSON.parse(fileText); // Parse the text as JSON
            } catch (parseError) {
                throw new Error("Invalid JSON file format.");
            }

            // Add metadata
            fileData.id = null
            fileData.meta = fileData.meta || {}; // Ensure `meta` exists
            fileData.meta.creator = fileData.meta.creator || {}; 
            fileData.meta.creator.email = sessionStorage.getItem('email');
    


            const response = await fetch(`${API_URL}/v0/models`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify(fileData) //Even though file is a File object, the Fetch API automatically converts it into a binary stream when used as the body. Now, we upload the parsed json object instead
            });

            if (!response.ok) {
                throw new Error(`Upload failed: ${response.statusText}`);
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
    }, []);

    const { getRootProps, getInputProps, isDragActive } = useDropzone({ onDrop });


    return (
        <Box sx={{ display: "flex", flexDirection: "row", gap: 2, p: 3, height: 'calc(100vh - 112px)' }}>
            {/* Left Navigation */}
            <Card sx={{ p: 2, minWidth: 200, display: 'flex', flexDirection: 'column', height: 'calc(100vh - 144px)' }}>
                <Tabs orientation="vertical">
                    <Tab label="Profile" component={NavLink} to="" />
                    <Tab label="Assets" component={NavLink} to="" />
                    <Tab label="Settings" component={NavLink} to="" />
                </Tabs>

                {/* This Box will take up the available space */}
                <Box sx={{ flexGrow: 1 }} />

                {/* Button is aligned to the bottom */}
                <Box sx={{ mt: 2 }}>
                    <Button variant="contained" color="error" fullWidth component={NavLink} to="">
                        Logout
                    </Button>
                </Box>
            </Card>

            {/* Main Content */}
            <Card sx={{ flex: 1, p: 3, display: 'flex', flexDirection: 'column', height: 'calc(100vh - 160px)' }}>
                <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", mb: 2 }}>
                    <Typography variant="h5" fontWeight="bold">My Assets</Typography>

                </Box>

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
                    //getRootProps() is a function that returns props that need to be applied to the root element of the dropzone (in this case, the Box component).
                    //{getRootProps} expands to a series of parameters
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

                {/* Table for Asset Details */}
                <TableContainer sx={{ marginTop: '2em' }}>
                    <Table stickyHeader>
                        <TableHead>
                            <TableRow>
                                <TableCell><Typography fontWeight="bold">Name</Typography></TableCell>
                                <TableCell><Typography fontWeight="bold">Size</Typography></TableCell>
                                <TableCell><Typography fontWeight="bold">Last Modified</Typography></TableCell>
                                <TableCell><Typography fontWeight="bold">Owner</Typography></TableCell>
                                <TableCell><Typography fontWeight="bold">Visibility</Typography></TableCell>
                                <TableCell><Typography fontWeight="bold">Actions</Typography></TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody sx={{ maxHeight: '100%' }}>
                            {[{ name: "Model X", size: "1.5 MB", date: "2/8/2025", owner: "User X", visibility: "Private" },
                            { name: "Model Y", size: "1.0 MB", date: "2/8/2025", owner: "User Y", visibility: "Public" },
                            { name: "Model X", size: "1.5 MB", date: "2/8/2025", owner: "User X", visibility: "Private" },
                            { name: "Model Y", size: "1.0 MB", date: "2/8/2025", owner: "User Y", visibility: "Public" },
                            { name: "Model X", size: "1.5 MB", date: "2/8/2025", owner: "User X", visibility: "Private" },
                            { name: "Model Y", size: "1.0 MB", date: "2/8/2025", owner: "User Y", visibility: "Public" },
                            { name: "Model X", size: "1.5 MB", date: "2/8/2025", owner: "User X", visibility: "Private" },
                            { name: "Model Y", size: "1.0 MB", date: "2/8/2025", owner: "User Y", visibility: "Public" },
                            { name: "Model X", size: "1.5 MB", date: "2/8/2025", owner: "User X", visibility: "Private" },
                            { name: "Model Y", size: "1.0 MB", date: "2/8/2025", owner: "User Y", visibility: "Public" },
                            { name: "Model X", size: "1.5 MB", date: "2/8/2025", owner: "User X", visibility: "Private" },
                            { name: "Model Y", size: "1.0 MB", date: "2/8/2025", owner: "User Y", visibility: "Public" },
                            { name: "Model X", size: "1.5 MB", date: "2/8/2025", owner: "User X", visibility: "Private" },
                            { name: "Model Y", size: "1.0 MB", date: "2/8/2025", owner: "User Y", visibility: "Public" },
                            ]
                                .map((asset, index) => (
                                    <TableRow key={index}>
                                        <TableCell>{asset.name}</TableCell>
                                        <TableCell>{asset.size}</TableCell>
                                        <TableCell>{asset.date}</TableCell>
                                        <TableCell>{asset.owner}</TableCell>
                                        <TableCell>{asset.visibility}</TableCell>
                                        <TableCell>
                                            <Button variant="outlined">Select</Button>
                                        </TableCell>
                                    </TableRow>
                                ))}
                        </TableBody>
                    </Table>
                </TableContainer>
            </Card>
        </Box>
    );
};

export default UploadPage;
