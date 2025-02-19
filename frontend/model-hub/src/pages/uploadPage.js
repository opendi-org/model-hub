//
// COPYRIGHT OpenDI
//

import { NavLink } from "react-router-dom";
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
    Paper,
    Typography
} from "@mui/material";

import { useDropzone } from "react-dropzone";
import { useCallback } from "react";

const UploadPage = () => {

    //drag and drop functionality
    const onDrop = useCallback(async (acceptedFiles) => {
        console.log(acceptedFiles);

        const file = acceptedFiles[0];

        try {
            // const fileText = await file.text();
            const response = await fetch("http://localhost:8080/v0/models", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: file
            });

            if (!response.ok) {
                throw new Error(`Upload failed: ${response.statusText}`);
            }

            const result = await response.json();
            console.log("Upload success:", result);
        } catch (error) {
            console.error("Error uploading file:", error);
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
