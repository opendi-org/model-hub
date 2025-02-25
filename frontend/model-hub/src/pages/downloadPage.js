//
// COPYRIGHT OpenDI
//

import { NavLink } from "react-router-dom";
import * as React from 'react';
import { useEffect } from 'react';
import { useState } from 'react';
import opendiIcon from '../opendi-icon.png';
import {
    Box,
    Button,
    Tabs,
    Tab,
    Typography,
    Link,
    Stack,
    Breadcrumbs,
    Chip
} from "@mui/material";
import { useParams } from "react-router-dom";
import { useDropzone } from "react-dropzone";
import { useCallback } from "react";

const DownloadPage = () => {

    const cdm = {
        creator: 'No CDM loaded'
    };

    const { uuid } = useParams();
    // console.log( uuid );

    const [model, setModel] = useState({})
    useEffect(() => {
        fetch(`http://localhost:8080/v0/models/${uuid}`)
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

    async function getCDM() {
        // console.log("Creator: " + cdm.creator);
        try {
            const json = model;
            cdm.creator = json.meta.creator;
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
                    <Typography variant="subtitle1" sx={{ pb: 2 }}> By: {model.meta ? model.meta.creator : ""} </Typography>

                    <Stack direction="row" spacing={1} sx={{ pb: 8 }}>
                        <Chip label="Tag 1" />
                        <Chip label="Tag 2" />
                        <Chip label="Tag 3" />
                        <Chip label="Tag 4" />
                    </Stack>

                    <Button
                        variant="outlined"
                        sx={{ width: "30%" }}
                        component={NavLink}
                        to=""
                        onClick={getCDM}
                    >
                        Download
                    </Button>
                </Box>
            </Box>

            <Box sx={{ p: 3 }}>
                <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
                    <Tabs value={value} onChange={handleChange} aria-label="basic tabs example">
                        <Tab label="Overview" />
                        <Tab label="Documentation" />
                        <Tab label="Item Three" />
                    </Tabs>
                </Box>
                <CustomTabPanel value={value} index={0}>
                    {model.meta ? model.meta.summary : ""}
                </CustomTabPanel>
                <CustomTabPanel value={value} index={1}>
                {model.meta && model.meta.documentation ? model.meta.documentation.content : ""}
                </CustomTabPanel>
                <CustomTabPanel value={value} index={2}>
                    Item Three
                </CustomTabPanel>
            </Box>
        </Box>
    );
};

export default DownloadPage;
