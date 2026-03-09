import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  Button,
  Card,
  CardActionArea,
  Chip,
  Container,
  CircularProgress,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  IconButton,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
  Alert,
} from '@mui/material';
import Grid from '@mui/material/Grid';
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import PublicIcon from '@mui/icons-material/Public';
import LockIcon from '@mui/icons-material/Lock';
import APIClient from '../util/ApiClient';
import { useUser } from '../context/UserContext';
import { useRepositories } from '../context/RepositoryContext';

const RepositoriesPage = () => {
  const { user, loading: userLoading } = useUser();
  const { repositories, loading: reposLoading, addRepository, removeRepository } = useRepositories();
  const navigate = useNavigate();

  const [createOpen, setCreateOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [selectedRepo, setSelectedRepo] = useState(null);
  const [formSlug, setFormSlug] = useState('');
  const [formDescription, setFormDescription] = useState('');
  const [formVisibility, setFormVisibility] = useState('private');
  const [formError, setFormError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleCreateOpen = () => {
    setFormSlug('');
    setFormDescription('');
    setFormVisibility('private');
    setFormError('');
    setCreateOpen(true);
  };

  const handleCreateClose = () => {
    setCreateOpen(false);
  };

  const handleCreateSubmit = async () => {
    if (!formSlug.trim()) {
      setFormError('Repository name is required.');
      return;
    }
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
    } catch (err) {
      setFormError(err.message || 'Failed to create repository.');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteOpen = (e, repo) => {
    e.stopPropagation();
    setSelectedRepo(repo);
    setDeleteOpen(true);
  };

  const handleDeleteClose = () => {
    setDeleteOpen(false);
    setSelectedRepo(null);
  };

  const handleDeleteConfirm = async () => {
    if (!selectedRepo) return;
    setSubmitting(true);
    try {
      await APIClient.deleteRepository(selectedRepo.id);
      removeRepository(selectedRepo.id);
      setDeleteOpen(false);
      setSelectedRepo(null);
    } catch (err) {
      setFormError(err.message || 'Failed to delete repository.');
    } finally {
      setSubmitting(false);
    }
  };

  if (userLoading || reposLoading) {
    return (
      <Container maxWidth="md" sx={{ py: 8, display: 'flex', justifyContent: 'center' }}>
        <CircularProgress />
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
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 3 }}>
        <Typography variant="h4" fontWeight="bold">
          My Repositories
        </Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={handleCreateOpen}>
          Create Repository
        </Button>
      </Box>

      {repositories.length === 0 ? (
        <Box sx={{ textAlign: 'center', py: 8, color: 'text.secondary' }}>
          <Typography variant="body1">No repositories yet.</Typography>
          <Typography variant="body2" sx={{ mt: 1 }}>
            Create your first repository to start managing models and tags.
          </Typography>
          <Button variant="contained" sx={{ mt: 2 }} onClick={handleCreateOpen}>
            Create a Repository
          </Button>
        </Box>
      ) : (
        <Grid container spacing={2}>
          {repositories.map((repo) => (
            <Grid key={repo.id} size={{ xs: 12, sm: 6, md: 4 }}>
              <Card
                sx={{
                  height: '100%',
                  display: 'flex',
                  flexDirection: 'column',
                  transition: 'box-shadow 0.2s',
                  '&:hover': { boxShadow: 6 },
                }}
              >
                <CardActionArea
                  onClick={() => navigate(`/repositories/${repo.id}`)}
                  sx={{ flexGrow: 1, p: 2.5, display: 'flex', flexDirection: 'column', alignItems: 'flex-start' }}
                >
                  <Box sx={{ display: 'flex', alignItems: 'center', width: '100%', mb: 1 }}>
                    <Typography variant="h6" fontWeight={600} noWrap sx={{ flexGrow: 1 }}>
                      {repo.slug || repo.name}
                    </Typography>
                    <IconButton
                      size="small"
                      onClick={(e) => handleDeleteOpen(e, repo)}
                      sx={{ ml: 1, color: 'text.secondary', '&:hover': { color: 'error.main' } }}
                    >
                      <DeleteOutlineIcon fontSize="small" />
                    </IconButton>
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
                <PublicIcon fontSize="small" sx={{ mr: 0.5 }} /> Public
              </ToggleButton>
              <ToggleButton value="private">
                <LockIcon fontSize="small" sx={{ mr: 0.5 }} /> Private
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
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDeleteClose} disabled={submitting}>Cancel</Button>
          <Button variant="contained" color="error" onClick={handleDeleteConfirm} disabled={submitting}>
            {submitting ? <CircularProgress size={20} /> : 'Delete'}
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};

export default RepositoriesPage;
