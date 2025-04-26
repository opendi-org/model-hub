//
// COPYRIGHT OpenDI
//
import React, { useState } from 'react';
import { CardActions, Typography} from '@mui/material';
import Grid from '@mui/material/Grid2';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import Button from '@mui/material/Button';
import { useHistory } from 'react-router-dom';
import { NavLink } from "react-router-dom";

const ModelMinicard = ({id, name, author, summary, version, updatedDate}) => {
    let color = '#ffffff'
    let hoverColor = '#ffffff'

    //For coloring models based off of keywords in their summary - should be replaced by model tags
    
    //const keywords = ["Financial", "Medical", "Business", "Technical"];
    //const foundKeyword = keywords.find(word => summary.includes(word)) || "";
    /**switch(foundKeyword) {
        case 'Financial':
            color = '#6ae48a';
            hoverColor = '#92ebaa'
            break;
        case 'Medical':
            color = '#ff555a';
            hoverColor = '#ff9396'
            break;
        case 'Business':
            color = '#f9ff6b'
            hoverColor = '#fbffa9'
            break;
        case 'Technical':
            color = '#b595ff'
            hoverColor = '#d9c8ff'
    }*/
    return ( 
    <Grid xs={4}>
        <Card sx={{ minWidth: 275, maxWidth: 550}}>
            <CardHeader
                title = {name}
                subheader={author}
                action = {
                    <Typography gutterBottom variant="body2" sx={{ color: 'text.secondary' }}>
                        {version}
                    </Typography>
                }
                sx={{bgcolor: '#63bad6',  '&:hover': {bgcolor: '#34a4c8'}}}
                component={NavLink} 
                to={"/model/" + id}
                style={{ textDecoration: 'none', color: 'inherit' }}
            />
            <CardContent>
                <Typography variant="body2">
                    {summary}
                </Typography>
                <Typography gutterBottom variant="body2" sx={{ color: 'text.secondary' }}>
                    {'Last Updated: ' + updatedDate}
                </Typography>
            </CardContent>
        </Card>
    </Grid>
)};

export default ModelMinicard;