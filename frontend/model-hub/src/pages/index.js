//
// COPYRIGHT OpenDI
//

import { Button, Container, Typography, Box, TextField, InputAdornment, FormControl, InputLabel, Select, MenuItem } from '@mui/material';
import Grid from '@mui/material/Grid';
import { useEffect, useState, useMemo } from 'react';
import { useTheme } from '@mui/material/styles';
import { useNavigate, useParams, useLocation } from 'react-router-dom';
import APIClient from '../util/ApiClient';
import Card from '@mui/material/Card';
import CardActionArea from '@mui/material/CardActionArea';
import LockIcon from '@mui/icons-material/Lock';
import PublicIcon from '@mui/icons-material/Public';
import SearchIcon from '@mui/icons-material/Search';
import AccessTimeIcon from '@mui/icons-material/AccessTime';

function timeAgo(dateString) {
  if (!dateString) return '';
  const seconds = Math.floor((Date.now() - new Date(dateString).getTime()) / 1000);
  if (seconds < 60) return 'just now';
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}d ago`;
  const months = Math.floor(days / 30);
  if (months < 12) return `${months}mo ago`;
  return `${Math.floor(months / 12)}y ago`;
}

const Home = () => {
    const { owner } = useParams();
    const [repositories, setRepositories] = useState([]);
    const [loading, setLoading] = useState(true);
    const [searchQuery, setSearchQuery] = useState('');
    const [sortBy, setSortBy] = useState('updated');
    const [sortOrder, setSortOrder] = useState('desc');
    const theme = useTheme();
    const navigate = useNavigate();
    const location = useLocation();

    useEffect(() => {
        setLoading(true);
        APIClient.globalSearch('', 'public', owner || '')
            .then(data => {
                const allRepos = Array.isArray(data) ? data : data.repositories || [];
                // Defensive guard: Home should only display public repositories.
                setRepositories(allRepos.filter((repo) => repo?.visibility === 'public'));
            })
            .catch(error => {
                console.error('There was a problem fetching repositories:', error);
                setRepositories([]);
            })
            .finally(() => setLoading(false));
    }, [owner, location.pathname]);

    const filtered = useMemo(() => {
        let result = repositories;

        // Filter by search query
        if (searchQuery.trim()) {
            const q = searchQuery.toLowerCase();
            result = result.filter((r) => {
                return (r.slug || '').toLowerCase().includes(q) ||
                       (r.description || '').toLowerCase().includes(q) ||
                       (r.owner || '').toLowerCase().includes(q);
            });
        }

        // Sort
        result = [...result].sort((a, b) => {
            let aVal, bVal;
            if (sortBy === 'name') {
                aVal = (a.slug || '').toLowerCase();
                bVal = (b.slug || '').toLowerCase();
                // Use string comparison for names
                const comparison = aVal.localeCompare(bVal);
                return sortOrder === 'asc' ? comparison : -comparison;
            } else if (sortBy === 'created') {
        const getRepoDate = (repo, camelKey, snakeKey) => {
          const v = repo?.[camelKey] ?? repo?.[snakeKey];
          if (!v) return 0;
          const t = new Date(v).getTime();
          return Number.isFinite(t) ? t : 0;
        };
        aVal = getRepoDate(a, 'createdAt', 'created_at');
        bVal = getRepoDate(b, 'createdAt', 'created_at');
            } else {
        const getRepoDate = (repo, camelKey, snakeKey) => {
          const v = repo?.[camelKey] ?? repo?.[snakeKey];
          if (!v) return 0;
          const t = new Date(v).getTime();
          return Number.isFinite(t) ? t : 0;
        };
        aVal = getRepoDate(a, 'updatedAt', 'updated_at');
        bVal = getRepoDate(b, 'updatedAt', 'updated_at');
            }
            // Use numeric comparison for dates
            return sortOrder === 'asc' ? aVal - bVal : bVal - aVal;
        });

        return result;
    }, [repositories, searchQuery, sortBy, sortOrder]);

    return (
        <Box sx={{ minHeight: '100vh', backgroundColor: theme.palette.background.default }}>
            {/* Hero section */}
            <Box
                sx={{
                    background: theme.palette.mode === 'light'
                        ? 'linear-gradient(135deg, #1E2130 0%, #0D2B55 100%)'
                        : 'linear-gradient(135deg, #0D1117 0%, #0D2244 100%)',
                    color: '#ffffff',
                    py: { xs: 4, md: 5 },
                    px: 3,
                    textAlign: 'center',
                }}
            >
                <Container maxWidth="md">
                    <Typography
                        variant="h5"
                        fontWeight={700}
                        gutterBottom
                        sx={{ letterSpacing: '-0.02em', mb: 1.25 }}
                    >
                        Get started with OpenDI
                    </Typography>
                    <Typography
                        variant="body1"
                        sx={{
                            color: 'rgba(255,255,255,0.75)',
                            maxWidth: 680,
                            mx: 'auto',
                            lineHeight: 1.6,
                            mb: 2,
                        }}
                    >
                        The purpose of the OpenDI initiative is to foster a vibrant and healthy
                        ecosystem for decision intelligence (DI), which supports innovative DI
                        research, a healthy vendor market, and — ultimately — better decisions
                        in many domains worldwide.
                    </Typography>
                    <Button
                        variant="contained"
                        size="medium"
                        href="https://opendi.org"
                        target="_blank"
                        rel="noopener noreferrer"
                        sx={{
                            backgroundColor: '#086DD7',
                            '&:hover': { backgroundColor: '#0558AE' },
                            px: 3,
                            py: 0.9,
                            fontWeight: 700,
                            fontSize: '0.9rem',
                        }}
                    >
                        Start Here
                    </Button>
                </Container>
            </Box>

            {/* Models grid */}
            <Container maxWidth="xl" sx={{ py: 4, px: { xs: 2, md: 4 } }}>
                <Box sx={{ mb: 3, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                    <Typography variant="h6" fontWeight={600} color="text.primary">
                        {owner ? `${owner}'s Repositories` : 'Model Repositories'}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                        {filtered.length} {filtered.length === 1 ? 'repository' : 'repositories'}
                    </Typography>
                </Box>

                {/* Search & Filters */}
                <Box sx={{ mb: 3, display: 'flex', gap: 2, flexWrap: 'wrap', alignItems: 'flex-end' }}>
                    <TextField
                        size="small"
                        placeholder="Search repositories..."
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
                    <FormControl size="small" sx={{ minWidth: 120 }}>
                        <InputLabel>Sort By</InputLabel>
                        <Select
                            value={sortBy}
                            label="Sort By"
                            onChange={(e) => setSortBy(e.target.value)}
                        >
                            <MenuItem value="updated">Updated</MenuItem>
                            <MenuItem value="created">Created</MenuItem>
                            <MenuItem value="name">Name</MenuItem>
                        </Select>
                    </FormControl>
                    <FormControl size="small" sx={{ minWidth: 100 }}>
                        <InputLabel>Order</InputLabel>
                        <Select
                            value={sortOrder}
                            label="Order"
                            onChange={(e) => setSortOrder(e.target.value)}
                        >
                            <MenuItem value="desc">Descending</MenuItem>
                            <MenuItem value="asc">Ascending</MenuItem>
                        </Select>
                    </FormControl>
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
                        <Typography variant="body1">No repositories found.</Typography>
                        <Typography variant="body2" sx={{ mt: 1 }}>
                            Create your own repository to get started.
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
                                        onClick={() => navigate(`/repositories/${repo.owner}/${repo.slug}`, { state: { from: location.pathname } })}
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
                                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mt: 'auto' }}>
                                            <AccessTimeIcon sx={{ fontSize: 14, color: 'text.secondary' }} />
                                            <Typography variant="caption" color="text.secondary">
                                                Updated {timeAgo(repo.updatedAt)}
                                            </Typography>
                                        </Box>
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

export default Home;
