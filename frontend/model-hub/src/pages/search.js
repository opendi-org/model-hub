import { Container, Typography, Box, TextField, InputAdornment } from '@mui/material';
import Grid from '@mui/material/Grid';
import { useEffect, useState, useMemo } from 'react';
import { useTheme } from '@mui/material/styles';
import { useNavigate } from 'react-router-dom';
import APIClient from '../util/ApiClient';
import Card from '@mui/material/Card';
import CardActionArea from '@mui/material/CardActionArea';
import LockIcon from '@mui/icons-material/Lock';
import PublicIcon from '@mui/icons-material/Public';
import SearchIcon from '@mui/icons-material/Search';

const ExplorePage = () => {
    const [repositories, setRepositories] = useState([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const theme = useTheme();
    const navigate = useNavigate();

    useEffect(() => {
        APIClient.getRepositories('all')
            .then(data => {
                setRepositories(Array.isArray(data) ? data : data.repositories || []);
            })
            .catch(error => {
                console.error('There was a problem fetching repositories:', error);
                setRepositories([]);
            })
            .finally(() => setLoading(false));
    }, []);

    const filtered = useMemo(() => {
        const q = searchQuery.toLowerCase();
        return repositories.filter((r) => {
            return (r.slug || '').toLowerCase().includes(q) ||
                   (r.description || '').toLowerCase().includes(q) ||
                   (r.owner || '').toLowerCase().includes(q);
        });
    }, [repositories, searchQuery]);

    return (
        <Box sx={{ minHeight: '100vh', backgroundColor: theme.palette.background.default }}>
            <Container maxWidth="xl" sx={{ py: 4, px: { xs: 2, md: 4 } }}>
                <Box sx={{ mb: 3, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <Typography variant="h5" fontWeight={600} color="text.primary">
                        Explore Public Repositories
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                        {repositories.length} {repositories.length === 1 ? 'repository' : 'repositories'}
                    </Typography>
                </Box>

                {/* Search Bar */}
                <Box sx={{ mb: 3 }}>
                    <TextField
                        size="small"
                        placeholder="Filter repositories..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        slotProps={{
                            input: {
                                startAdornment: (
                                    <InputAdornment position="start">
                                        <SearchIcon fontSize="small" color="action" />
                                    </InputAdornment>
                                ),
                            },
                        }}
                        sx={{ minWidth: 240 }}
                    />
                </Box>

                {loading ? (
                    <Box sx={{ textAlign: 'center', py: 8 }}>
                        <Typography variant="body2" color="text.secondary">Loading...</Typography>
                    </Box>
                ) : repositories.length === 0 ? (
                    <Box
                        sx={{
                            textAlign: 'center',
                            py: 8,
                            color: 'text.secondary',
                        }}
                    >
                        <Typography variant="body1">No public repositories yet.</Typography>
                        <Typography variant="body2" sx={{ mt: 1 }}>
                            Check back later or create your own repository and make it public.
                        </Typography>
                    </Box>
                ) : filtered.length === 0 ? (
                    <Box sx={{ textAlign: 'center', py: 8, color: 'text.secondary' }}>
                        <SearchIcon sx={{ fontSize: 48, mb: 1 }} />
                        <Typography variant="body1">No repositories match "{searchQuery}"</Typography>
                    </Box>
                ) : (
                    <Grid container spacing={2}>
                        {filtered.map((repo) => (
                            <Grid key={repo.id} size={{ xs: 12, sm: 6, md: 4 }}>
                                <Card
                                    sx={{
                                        height: '100%',
                                        display: 'flex',
                                        flexDirection: 'column',
                                    }}
                                >
                                    <CardActionArea
                                        onClick={() => navigate(`/repositories/${repo.id}`)}
                                        sx={{ flexGrow: 1, p: 2.5, display: 'flex', flexDirection: 'column', alignItems: 'flex-start' }}
                                    >
                                        <Box sx={{ display: 'flex', alignItems: 'center', width: '100%', mb: 0.5 }}>
                                            <Box sx={{ flexGrow: 1, minWidth: 0 }}>
                                                <Typography variant="caption" color="text.secondary" noWrap>
                                                    {repo.owner}
                                                </Typography>
                                                <Typography variant="h6" fontWeight={600} noWrap>
                                                    {repo.slug}
                                                </Typography>
                                            </Box>
                                            {repo.visibility === 'public' ? (
                                                <PublicIcon sx={{ fontSize: 20, color: 'success.main', ml: 1 }} />
                                            ) : (
                                                <LockIcon sx={{ fontSize: 20, color: 'warning.main', ml: 1 }} />
                                            )}
                                        </Box>
                                        <Typography variant="body2" color="text.secondary" sx={{ mb: 1, display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden' }}>
                                            {repo.description || 'No description'}
                                        </Typography>
                                        <Typography variant="caption" color="text.secondary" sx={{ mt: 'auto' }}>
                                            Updated {new Date(repo.updated_at).toLocaleDateString()}
                                        </Typography>
                                    </CardActionArea>
                                </Card>
                            </Grid>
                        ))}
                    </Grid>
                )}
            </Container>
        </Box>
    );
};

export default ExplorePage;