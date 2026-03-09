import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Alert,
  Box,
  Button,
  Card,
  Chip,
  CircularProgress,
  Container,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Divider,
  IconButton,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from '@mui/material';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import PublicIcon from '@mui/icons-material/Public';
import LockIcon from '@mui/icons-material/Lock';
import { useDropzone } from 'react-dropzone';
import APIClient from '../util/ApiClient';
import { useUser } from '../context/UserContext';

const RepositoryDetailsPage = () => {
  const { repositoryId } = useParams();
  const navigate = useNavigate();
  const { user } = useUser();

  const [repo, setRepo] = useState(null);
  const [tags, setTags] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Add-tag dialog state
  const [addOpen, setAddOpen] = useState(false);
  const [tagName, setTagName] = useState('');
  const [tagFile, setTagFile] = useState(null);
  const [tagFileData, setTagFileData] = useState(null);
  const [addError, setAddError] = useState('');
  const [addSubmitting, setAddSubmitting] = useState(false);

  // Delete-tag dialog state
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [selectedTag, setSelectedTag] = useState(null);
  const [deleteSubmitting, setDeleteSubmitting] = useState(false);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const [repoData, tagsData] = await Promise.all([
        APIClient.getRepositoryById(repositoryId),
        APIClient.getRepositoryTags(repositoryId),
      ]);
      setRepo(repoData);
      setTags(Array.isArray(tagsData) ? tagsData : tagsData.tags ?? []);
    } catch (err) {
      setError(err.message || 'Failed to load repository.');
    } finally {
      setLoading(false);
    }
  }, [repositoryId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // --- Add Tag ---
  const handleAddOpen = () => {
    setTagName('');
    setTagFile(null);
    setTagFileData(null);
    setAddError('');
    setAddOpen(true);
  };

  const handleAddClose = () => setAddOpen(false);

  const onDrop = useCallback(async (acceptedFiles) => {
    const file = acceptedFiles[0];
    if (!file) return;
    try {
      const text = await file.text();
      const json = JSON.parse(text);
      setTagFile(file);
      setTagFileData(json);
      setAddError('');
    } catch {
      setAddError('Invalid JSON file.');
      setTagFile(null);
      setTagFileData(null);
    }
  }, []);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: { 'application/json': ['.json'] },
    multiple: false,
    noClick: false,
  });

  const handleAddSubmit = async () => {
    if (!tagName.trim()) {
      setAddError('Tag name is required.');
      return;
    }
    if (!tagFileData) {
      setAddError('Please upload a CDM JSON file.');
      return;
    }
    setAddSubmitting(true);
    setAddError('');
    try {
      if (user) {
        tagFileData.meta = tagFileData.meta || {};
        tagFileData.meta.creator = tagFileData.meta.creator || {};
        tagFileData.meta.creator.email = user.email;
      }
      await APIClient.createOrUpdateTag(repositoryId, tagName.trim(), tagFileData);
      setAddOpen(false);
      await fetchData();
    } catch (err) {
      setAddError(err.message || 'Failed to create tag.');
    } finally {
      setAddSubmitting(false);
    }
  };

  // --- Delete Tag ---
  const handleDeleteOpen = (tag) => {
    setSelectedTag(tag);
    setDeleteOpen(true);
  };

  const handleDeleteClose = () => {
    setDeleteOpen(false);
    setSelectedTag(null);
  };

  const handleDeleteConfirm = async () => {
    if (!selectedTag) return;
    setDeleteSubmitting(true);
    try {
      await APIClient.deleteTag(repositoryId, selectedTag.name);
      setTags((prev) => prev.filter((t) => t.id !== selectedTag.id));
      setDeleteOpen(false);
      setSelectedTag(null);
    } catch (err) {
      setAddError(err.message || 'Failed to delete tag.');
    } finally {
      setDeleteSubmitting(false);
    }
  };

  if (loading) {
    return (
      <Container maxWidth="md" sx={{ py: 8, display: 'flex', justifyContent: 'center' }}>
        <CircularProgress />
      </Container>
    );
  }

  if (error) {
    return (
      <Container maxWidth="md" sx={{ py: 8 }}>
        <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>
        <Button startIcon={<ArrowBackIcon />} onClick={() => navigate('/repositories')}>
          Back to Repositories
        </Button>
      </Container>
    );
  }

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      {/* Header */}
      <Button startIcon={<ArrowBackIcon />} onClick={() => navigate('/repositories')} sx={{ mb: 2 }}>
        Back to Repositories
      </Button>

      <Card sx={{ p: 3, mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1.5, mb: 1 }}>
          <Typography variant="h4" fontWeight="bold">
            {repo?.slug || repo?.name}
          </Typography>
          {repo?.visibility && (
            <Chip
              icon={repo.visibility === 'public' ? <PublicIcon /> : <LockIcon />}
              label={repo.visibility}
              size="small"
              variant="outlined"
            />
          )}
        </Box>
        {repo?.description && (
          <Typography variant="body1" color="text.secondary" sx={{ mb: 1 }}>
            {repo.description}
          </Typography>
        )}
        {repo?.owner?.username && (
          <Typography variant="body2" color="text.secondary">
            Owner: {repo.owner.username}
          </Typography>
        )}
      </Card>

      <Divider sx={{ mb: 3 }} />

      {/* Tags Section */}
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
        <Typography variant="h5" fontWeight={600}>
          Tags
        </Typography>
        <Button variant="contained" startIcon={<AddIcon />} onClick={handleAddOpen}>
          Add Tag
        </Button>
      </Box>

      {tags.length === 0 ? (
        <Box sx={{ textAlign: 'center', py: 6, color: 'text.secondary' }}>
          <Typography variant="body1">No tags yet.</Typography>
          <Typography variant="body2" sx={{ mt: 1 }}>
            Add a tag by uploading a CDM model file.
          </Typography>
        </Box>
      ) : (
        <TableContainer component={Card}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell><strong>Tag</strong></TableCell>
                <TableCell><strong>Model UUID</strong></TableCell>
                <TableCell><strong>Size</strong></TableCell>
                <TableCell><strong>Updated</strong></TableCell>
                <TableCell align="right"><strong>Actions</strong></TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {tags.map((tag) => (
                <TableRow key={tag.id || tag.name} hover>
                  <TableCell>
                    <Chip label={tag.name} color="primary" size="small" sx={{ borderRadius: '12px' }} />
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" sx={{ fontFamily: 'monospace' }} noWrap>
                      {tag.modelUUID || tag.model_uuid || '—'}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    {formatBytes(tag.sizeBytes ?? tag.size_bytes)}
                  </TableCell>
                  <TableCell>
                    {tag.updatedAt || tag.updated_at
                      ? new Date(tag.updatedAt || tag.updated_at).toLocaleDateString()
                      : '—'}
                  </TableCell>
                  <TableCell align="right">
                    <IconButton
                      size="small"
                      onClick={() => handleDeleteOpen(tag)}
                      sx={{ color: 'text.secondary', '&:hover': { color: 'error.main' } }}
                    >
                      <DeleteOutlineIcon fontSize="small" />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {/* Add Tag Dialog */}
      <Dialog open={addOpen} onClose={handleAddClose} maxWidth="sm" fullWidth>
        <DialogTitle>Add Tag</DialogTitle>
        <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: '16px !important' }}>
          {addError && <Alert severity="error">{addError}</Alert>}
          <TextField
            label="Tag Name"
            value={tagName}
            onChange={(e) => setTagName(e.target.value)}
            fullWidth
            autoFocus
            required
            placeholder="e.g. v1.0, latest"
          />
          <Box
            {...getRootProps()}
            sx={{
              border: '2px dashed',
              borderColor: isDragActive ? 'primary.main' : 'grey.400',
              borderRadius: 2,
              p: 4,
              textAlign: 'center',
              cursor: 'pointer',
              backgroundColor: isDragActive ? 'action.hover' : 'background.paper',
              transition: 'all 0.2s ease-in-out',
              '&:hover': {
                backgroundColor: 'action.hover',
                borderColor: 'primary.main',
              },
            }}
          >
            <input {...getInputProps()} />
            {tagFile ? (
              <Typography variant="body1" color="primary">
                {tagFile.name}
              </Typography>
            ) : (
              <>
                <Typography variant="body1" gutterBottom>
                  {isDragActive ? 'Drop the JSON file here' : 'Drag & drop a CDM JSON file here'}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  or click to select a file
                </Typography>
              </>
            )}
            <Typography variant="caption" display="block" sx={{ mt: 1 }} color="text.secondary">
              Only .json files are accepted
            </Typography>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleAddClose} disabled={addSubmitting}>Cancel</Button>
          <Button variant="contained" onClick={handleAddSubmit} disabled={addSubmitting}>
            {addSubmitting ? <CircularProgress size={20} /> : 'Upload & Create Tag'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Delete Tag Confirmation */}
      <Dialog open={deleteOpen} onClose={handleDeleteClose} maxWidth="xs" fullWidth>
        <DialogTitle>Delete Tag</DialogTitle>
        <DialogContent>
          <Typography>
            Are you sure you want to delete tag <strong>{selectedTag?.name}</strong>?
            This action cannot be undone.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDeleteClose} disabled={deleteSubmitting}>Cancel</Button>
          <Button variant="contained" color="error" onClick={handleDeleteConfirm} disabled={deleteSubmitting}>
            {deleteSubmitting ? <CircularProgress size={20} /> : 'Delete'}
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};

function formatBytes(bytes) {
  if (bytes == null) return '—';
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

export default RepositoryDetailsPage;
