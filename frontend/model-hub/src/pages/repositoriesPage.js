import { useState, useMemo, useEffect } from 'react';
import { useNavigate, useLocation, useSearchParams } from 'react-router-dom';
import {
  Box,
  Button,
  Card,
  CardActionArea,
  Chip,
  Container,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  IconButton,
  InputAdornment,
  Skeleton,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
  Alert,
  CircularProgress,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
} from '@mui/material';
import Grid from '@mui/material/Grid';
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import PublicIcon from '@mui/icons-material/Public';
import LockIcon from '@mui/icons-material/Lock';
import SearchIcon from '@mui/icons-material/Search';
import FolderOffOutlinedIcon from '@mui/icons-material/FolderOffOutlined';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import APIClient from '../util/ApiClient';
import { useUser } from '../context/UserContext';
import { useRepositories } from '../context/RepositoryContext';
import { useNotification } from '../context/NotificationContext';

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

const RepositoriesPage = () => {
  const { user, loading: userLoading } = useUser();
  const {
    repositories, loading: reposLoading,
    addRepository, removeRepository,
    scope, changeScope, refreshRepositories,
  } = useRepositories();
  const { showNotification } = useNotification();
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams, setSearchParams] = useSearchParams();

  const [searchQuery, setSearchQuery] = useState('');
  const [sortBy, setSortBy] = useState('updated');
  const [sortOrder, setSortOrder] = useState('desc');
  const [createOpen, setCreateOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [selectedRepo, setSelectedRepo] = useState(null);
  const [formSlug, setFormSlug] = useState('');
  const [formDescription, setFormDescription] = useState('');
  const [formVisibility, setFormVisibility] = useState('private');
  const [formError, setFormError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [deleteConfirmation, setDeleteConfirmation] = useState('');
  const [publicConfirmOpen, setPublicConfirmOpen] = useState(false);

  const updateSearchInUrl = (nextValue) => {
    const nextParams = new URLSearchParams(searchParams);
    const trimmed = nextValue.trim();
    if (trimmed) {
      nextParams.set('q', trimmed);
    } else {
      nextParams.delete('q');
    }
    setSearchParams(nextParams, { replace: true });
  };

  // Refresh repositories when navigating back to this page
  useEffect(() => {
    refreshRepositories();
  }, [location.pathname, refreshRepositories]);

  // Keep search input synced with URL query param (for navbar-driven search).
  useEffect(() => {
    const queryFromUrl = searchParams.get('q') ?? '';
    setSearchQuery(queryFromUrl);
  }, [searchParams]);

  const filtered = useMemo(() => {
    let result = repositories.filter((r) => {
      const q = searchQuery.toLowerCase();
      return (r.slug || '').toLowerCase().includes(q) ||
             (r.description || '').toLowerCase().includes(q) ||
             (r.owner || '').toLowerCase().includes(q);
    });

    const getRepoDate = (repo, camelKey, snakeKey) => {
      const v = repo?.[camelKey] ?? repo?.[snakeKey];
      if (!v) return 0;
      const t = new Date(v).getTime();
      return Number.isFinite(t) ? t : 0;
    };

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
        aVal = getRepoDate(a, 'createdAt', 'created_at');
        bVal = getRepoDate(b, 'createdAt', 'created_at');
      } else {
        aVal = getRepoDate(a, 'updatedAt', 'updated_at');
        bVal = getRepoDate(b, 'updatedAt', 'updated_at');
      }
      // Use numeric comparison for dates
      return sortOrder === 'asc' ? aVal - bVal : bVal - aVal;
    });

    return result;
  }, [repositories, searchQuery, sortBy, sortOrder]);

  const handleCreateOpen = () => {
    setFormSlug('');
    setFormDescription('');
    setFormVisibility('private');
    setFormError('');
    setCreateOpen(true);
  };

  const handleCreateClose = () => setCreateOpen(false);

  const handleCreateSubmit = async () => {
    if (!formSlug.trim()) {
      setFormError('Repository name is required.');
      return;
    }

    const slug = formSlug.trim();
    if (slug.length > 255) {
      setFormError('Repository name must be 255 characters or less.');
      return;
    }
    if (/\s/.test(slug)) {
      setFormError('Repository name cannot contain spaces.');
      return;
    }
    if (!/^[A-Za-z0-9_-]+$/.test(slug)) {
      setFormError("Repository name may contain only letters, numbers, '-' and '_' (no spaces).");
      return;
    }

    // Show confirmation if making repository public
    if (formVisibility === 'public') {
      setPublicConfirmOpen(true);
      return;
    }

    await performCreate();
  };

  const performCreate = async () => {
    setSubmitting(true);
    setFormError('');
    try {
      const repo = await APIClient.createRepository({
        slug: formSlug.trim(),
        description: formDescription.trim(),
        visibility: formVisibility,
      });
      addRepository(repo);
      setCreateOpen(false);
      setPublicConfirmOpen(false);
      showNotification('Repository created successfully', 'success');
    } catch (err) {
      const msg = err?.message || '';
      if (msg.toLowerCase().includes('invalid repository name')) {
        setFormError("Invalid repository name. Use 1-255 characters with letters/numbers and only '-' or '_' (no spaces).");
      } else {
        setFormError(msg || 'Failed to create repository.');
      }
      setPublicConfirmOpen(false);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteOpen = (e, repo) => {
    e.stopPropagation();
    setSelectedRepo(repo);
    setDeleteConfirmation('');
    setDeleteOpen(true);
  };

  const handleDeleteClose = () => {
    setDeleteOpen(false);
    setSelectedRepo(null);
    setDeleteConfirmation('');
  };

  const handleDeleteConfirm = async () => {
    if (!selectedRepo) return;
    setSubmitting(true);
    try {
      await APIClient.deleteRepository(selectedRepo.id);
      removeRepository(selectedRepo.id);
      setDeleteOpen(false);
      setSelectedRepo(null);
      showNotification('Repository deleted', 'success');
    } catch (err) {
      setFormError(err.message || 'Failed to delete repository.');
    } finally {
      setSubmitting(false);
    }
  };

  if (userLoading) {
    return (
      <Container maxWidth="lg" sx={{ py: 4 }}>
        <Skeleton variant="rectangular" height={48} sx={{ mb: 3, borderRadius: 1 }} />
        <Grid container spacing={2}>
          {[0, 1, 2, 3, 4, 5].map((i) => (
            <Grid key={i} size={{ xs: 12, sm: 6, md: 4 }}>
              <Skeleton variant="rectangular" height={160} sx={{ borderRadius: 1 }} />
            </Grid>
          ))}
        </Grid>
      </Container>
    );
  }

  if (!user) {
    return (
      <Container maxWidth="md" sx={{ py: 8 }}>
        <Card sx={{ p: 6, textAlign: 'center' }}>
          <LockOutlinedIcon sx={{ fontSize: 64, color: 'primary.main', mb: 2 }} />
          <Typography variant="h4" fontWeight="bold" gutterBottom>
            Login Required
          </Typography>
          <Typography variant="body1" color="text.secondary" sx={{ mb: 4 }}>
            You need to be logged in to view your repositories.
          </Typography>
          <Button
            variant="contained"
            size="large"
            onClick={() => { window.location.href = APIClient.getGoogleLoginUrl(); }}
          >
            Login with Google
          </Button>
        </Card>
      </Container>
    );
  }

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      {/* Header */}
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
        <Typography variant="h4" fontWeight="bold">
          My Repositories
        </Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={handleCreateOpen}>
          Create Repository
        </Button>
      </Box>

      {/* Scope Toggle */}
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 2, flexWrap: 'wrap' }}>
        <ToggleButtonGroup
          value={scope}
          exclusive
          onChange={(_, val) => { if (val) changeScope(val); }}
          size="small"
        >
          <ToggleButton value="mine">Owned Repos</ToggleButton>
          <ToggleButton value="shared-with-me">Shared With Me</ToggleButton>
        </ToggleButtonGroup>

        <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', flex: 1, alignItems: 'flex-end' }}>
          {/* Search */}
          <TextField
            size="small"
            placeholder="Search repositories..."
            value={searchQuery}
            onChange={(e) => {
              const nextValue = e.target.value;
              setSearchQuery(nextValue);
              updateSearchInUrl(nextValue);
            }}
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

        {/* Count */}
        <Typography variant="body2" color="text.secondary" sx={{ ml: 'auto' }}>
          {filtered.length} {filtered.length === 1 ? 'repository' : 'repositories'}
        </Typography>
      </Box>

      {/* Loading skeleton */}
      {reposLoading ? (
        <Grid container spacing={2}>
          {[0, 1, 2, 3, 4, 5].map((i) => (
            <Grid key={i} size={{ xs: 12, sm: 6, md: 4 }}>
              <Card sx={{ p: 2.5 }}>
                <Skeleton variant="text" width="60%" height={32} />
                <Skeleton variant="text" width="90%" height={20} sx={{ mt: 1 }} />
                <Skeleton variant="rectangular" width="40%" height={24} sx={{ mt: 2, borderRadius: 1 }} />
                <Skeleton variant="text" width="50%" height={16} sx={{ mt: 1.5 }} />
              </Card>
            </Grid>
          ))}
        </Grid>
      ) : repositories.length === 0 ? (
        <Box sx={{ textAlign: 'center', py: 8, color: 'text.secondary' }}>
          <FolderOffOutlinedIcon sx={{ fontSize: 64, color: 'primary.main', mb: 2 }} />
          {scope === 'mine' ? (
            <>
              <Typography variant="body1">No repositories yet.</Typography>
              <Typography variant="body2" sx={{ mt: 1 }}>
                Create your first repository to start managing models and tags.
              </Typography>
              <Button variant="contained" sx={{ mt: 2 }} onClick={handleCreateOpen}>
                Create a Repository
              </Button>
            </>
          ) : (
            <>
              <Typography variant="body1">No repositories shared with you.</Typography>
              <Typography variant="body2" sx={{ mt: 1 }}>
                Ask others to add you as a collaborator to access their repositories.
              </Typography>
            </>
          )}
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
                        {repo.slug || repo.name}
                      </Typography>
                    </Box>
                    {user && user.username === repo.owner && (
                      <IconButton
                        size="small"
                        onClick={(e) => handleDeleteOpen(e, repo)}
                        sx={{ ml: 1, color: 'text.secondary', '&:hover': { color: 'error.main' } }}
                      >
                        <DeleteOutlineIcon fontSize="small" />
                      </IconButton>
                    )}
                  </Box>

                  {repo.description && (
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 1.5 }} noWrap>
                      {repo.description}
                    </Typography>
                  )}

                  <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.75, mt: 'auto' }}>
                    {repo.visibility && (
                      <Chip
                        icon={repo.visibility === 'public' ? <PublicIcon /> : <LockIcon />}
                        label={repo.visibility}
                        size="small"
                        variant="outlined"
                        sx={{
                          color: repo.visibility === 'public' ? 'success.main' : 'text.secondary',
                          borderColor: repo.visibility === 'public' ? 'success.main' : 'text.secondary',
                          '& .MuiChip-icon': {
                            color: repo.visibility === 'public' ? 'success.main' : 'text.secondary',
                          },
                        }}
                      />
                    )}
                    {(repo.tags || []).map((tag) => (
                      <Chip
                        key={tag.id || tag.name}
                        label={tag.name}
                        size="small"
                        color="primary"
                        sx={{ borderRadius: '12px' }}
                      />
                    ))}
                  </Box>

                  {repo.updatedAt && (
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mt: 1.5 }}>
                      <AccessTimeIcon sx={{ fontSize: 14, color: 'text.secondary' }} />
                      <Typography variant="caption" color="text.secondary">
                        Updated {timeAgo(repo.updatedAt)}
                      </Typography>
                    </Box>
                  )}
                </CardActionArea>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}

      {/* Create Repository Dialog */}
      <Dialog open={createOpen} onClose={handleCreateClose} maxWidth="sm" fullWidth>
        <DialogTitle>Create Repository</DialogTitle>
        <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: '16px !important' }}>
          {formError && <Alert severity="error">{formError}</Alert>}
          <TextField
            label="Repository Name"
            value={formSlug}
            onChange={(e) => setFormSlug(e.target.value)}
            fullWidth
            autoFocus
            required
            placeholder="my-model-repo"
          />
          <TextField
            label="Description"
            value={formDescription}
            onChange={(e) => setFormDescription(e.target.value)}
            fullWidth
            multiline
            rows={2}
            placeholder="Optional description"
          />
          <Box>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>Visibility</Typography>
            <ToggleButtonGroup
              value={formVisibility}
              exclusive
              onChange={(_, val) => { if (val) setFormVisibility(val); }}
              size="small"
            >
              <ToggleButton value="public">
                <PublicIcon fontSize="small" sx={{ mr: 0.5, color: 'success.main' }} /> Public
              </ToggleButton>
              <ToggleButton value="private">
                <LockIcon fontSize="small" sx={{ mr: 0.5, color: 'text.secondary' }} /> Private
              </ToggleButton>
            </ToggleButtonGroup>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCreateClose} disabled={submitting}>Cancel</Button>
          <Button variant="contained" onClick={handleCreateSubmit} disabled={submitting}>
            {submitting ? <CircularProgress size={20} /> : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteOpen} onClose={handleDeleteClose} maxWidth="xs" fullWidth>
        <DialogTitle>Delete Repository</DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you want to delete <strong>{selectedRepo?.slug || selectedRepo?.name}</strong>?
            This action cannot be undone. All tags in this repository will also be deleted.
          </Typography>
          <Typography variant="body2" sx={{ mt: 2, color: 'text.secondary' }}>
            To confirm, type the repository name below:
          </Typography>
          <TextField
            fullWidth
            size="small"
            value={deleteConfirmation}
            onChange={(e) => setDeleteConfirmation(e.target.value)}
            sx={{ mt: 1 }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDeleteClose} disabled={submitting}>Cancel</Button>
          <Button
            variant="contained"
            color="error"
            onClick={handleDeleteConfirm}
            disabled={submitting || deleteConfirmation !== (selectedRepo?.slug || selectedRepo?.name)}
          >
            {submitting ? <CircularProgress size={20} /> : 'Delete'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Public Visibility Confirmation Dialog */}
      <Dialog open={publicConfirmOpen} onClose={() => setPublicConfirmOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Make Repository Public?</DialogTitle>
        <DialogContent>
          <Typography variant="body2" sx={{ mt: 2 }}>
            This repository will be visible to everyone. Anyone can view and download the repository contents.
          </Typography>
          <Typography variant="body2" sx={{ mt: 2, fontWeight: 500 }}>
            Are you sure you want to continue?
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setPublicConfirmOpen(false)} disabled={submitting}>
            Cancel
          </Button>
          <Button variant="contained" color="warning" onClick={performCreate} disabled={submitting}>
            {submitting ? <CircularProgress size={20} /> : 'Make Public'}
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};

export default RepositoriesPage;
