//
// COPYRIGHT OpenDI
//
import React from 'react';
import { Typography, Chip, Box } from '@mui/material';
import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import { NavLink } from "react-router-dom";
import { useTheme } from '@mui/material/styles';
import StorageIcon from '@mui/icons-material/Storage';

const ModelMinicard = ({ id, name, author, summary, version, updatedDate }) => {
    const theme = useTheme();

    return (
        <Grid xs={4}>
            <Card
                sx={{
                    minWidth: 275,
                    maxWidth: 550,
                    backgroundColor: theme.palette.background.paper,
                    cursor: 'pointer',
                }}
                component={NavLink}
                to={`/model/${id}`}
                style={{ textDecoration: 'none', display: 'block' }}
            >
                <CardHeader
                    avatar={
                        <Box
                            sx={{
                                width: 40,
                                height: 40,
                                borderRadius: 1,
                                backgroundColor: theme.palette.mode === 'light'
                                    ? 'rgba(8,109,215,0.08)'
                                    : 'rgba(8,109,215,0.2)',
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'center',
                                flexShrink: 0,
                            }}
                        >
                            <StorageIcon sx={{ color: theme.palette.primary.main, fontSize: 20 }} />
                        </Box>
                    }
                    title={
                        <Typography
                            variant="body1"
                            fontWeight={600}
                            color="text.primary"
                            sx={{
                                '&:hover': { color: theme.palette.primary.main },
                                transition: 'color 0.15s',
                            }}
                        >
                            {name}
                        </Typography>
                    }
                    subheader={
                        <Typography variant="caption" color="text.secondary">
                            by {author}
                        </Typography>
                    }
                    action={
                        version && (
                            <Chip
                                label={`v${version}`}
                                size="small"
                                sx={{
                                    mt: 0.5,
                                    mr: 0.5,
                                    backgroundColor: theme.palette.mode === 'light'
                                        ? 'rgba(8,109,215,0.08)'
                                        : 'rgba(8,109,215,0.2)',
                                    color: theme.palette.primary.main,
                                    fontWeight: 600,
                                    border: 'none',
                                }}
                            />
                        )
                    }
                />
                <CardContent sx={{ pb: '12px !important' }}>
                    <Typography
                        variant="body2"
                        color="text.secondary"
                        sx={{
                            display: '-webkit-box',
                            WebkitLineClamp: 2,
                            WebkitBoxOrient: 'vertical',
                            overflow: 'hidden',
                            lineHeight: 1.5,
                            mb: 1.5,
                            minHeight: '3em',
                        }}
                    >
                        {summary || 'No description provided.'}
                    </Typography>
                    {updatedDate && (
                        <Typography variant="caption" color="text.disabled">
                            Updated {updatedDate}
                        </Typography>
                    )}
                </CardContent>
            </Card>
        </Grid>
    );
};

export default ModelMinicard;
