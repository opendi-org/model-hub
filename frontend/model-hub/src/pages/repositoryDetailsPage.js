import { useState, useEffect, useCallback, useMemo } from 'react';
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
  Skeleton,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TableSortLabel,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Tooltip,
  Typography,
} from '@mui/material';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import EditOutlinedIcon from '@mui/icons-material/EditOutlined';
import PublicIcon from '@mui/icons-material/Public';
import LockIcon from '@mui/icons-material/Lock';
import ContentCopyOutlinedIcon from '@mui/icons-material/ContentCopyOutlined';
import LabelOffOutlinedIcon from '@mui/icons-material/LabelOffOutlined';
import { useDropzone } from 'react-dropzone';
import APIClient from '../util/ApiClient';
import { useUser } from '../context/UserContext';
import { useRepositories } from '../context/RepositoryContext';
import { useNotification } from '../context/NotificationContext';

const RepositoryDetailsPage = () => {
  const { repositoryId } = useParams();
  const navigate = useNavigate();
  const { user } = useUser();
  const { updateRepository: updateRepoInContext } = useRepositories();
  const { showNotification } = useNotification();

  const [repo, setRepo] = useState(null);
  const [tags, setTags] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Sort state
  const [sortField, setSortField] = useState('name');
  const [sortDir, setSortDir] = useState('asc');

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

  // Edit repo dialog state
  const [editOpen, setEditOpen] = useState(false);
  const [editSlug, setEditSlug] = useState('');
  const [editDescription, setEditDescription] = useState('');
  const [editVisibility, setEditVisibility] = useState('private');
  const [editError, setEditError] = useState('');
  const [editSubmitting, setEditSubmitting] = useState(false);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const repoData = await APIClient.getRepositoryById(repositoryId);
      setRepo(repoData);
      setTags(repoData.tags ?? []);
    } catch (err) {
      setError(err.message || 'Failed to load repository.');
    } finally {
      setLoading(false);
    }
  }, [repositoryId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // --- Sorting ---
  const handleSort = (field) => {
    if (sortField === field) {
      setSortDir((prev) => (prev === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortField(field);
      setSortDir('asc');
    }
  };

  const sortedTags = useMemo(() => {
    const sorted = [...tags];
    sorted.sort((a, b) => {
      let aVal = a[sortField];
      let bVal = b[sortField];
      if (sortField === 'size') {
        aVal = aVal ?? 0;
        bVal = bVal ?? 0;
        return sortDir === 'asc' ? aVal - bVal : bVal - aVal;
      }
      if (sortField === 'updatedAt') {
        aVal = aVal ? new Date(aVal).getTime() : 0;
        bVal = bVal ? new Date(bVal).getTime() : 0;
        return sortDir === 'asc' ? aVal - bVal : bVal - aVal;
      }
      aVal = (aVal || '').toLowerCase();
      bVal = (bVal || '').toLowerCase();
      if (aVal < bVal) return sortDir === 'asc' ? -1 : 1;
      if (aVal > bVal) return sortDir === 'asc' ? 1 : -1;
      return 0;
    });
    return sorted;
  }, [tags, sortField, sortDir]);

  // --- Copy digest ---
  const handleCopyDigest = (digest) => {
    navigator.clipboard.writeText(digest).then(() => {
      showNotification('Digest copied to clipboard', 'info');
    });
  };

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
      showNotification('Tag created successfully', 'success');
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
      setTags((prev) => prev.filter((t) => t.name !== selectedTag.name));
      setDeleteOpen(false);
      setSelectedTag(null);
      showNotification('Tag deleted', 'success');
    } catch (err) {
      showNotification(err.message || 'Failed to delete tag.', 'error');
    } finally {
      setDeleteSubmitting(false);
    }
  };

  // --- Edit Repo ---
  const handleEditOpen = () => {
    setEditSlug(repo?.slug || '');
    setEditDescription(repo?.description || '');
    setEditVisibility(repo?.visibility || 'private');
    setEditError('');
    setEditOpen(true);
  };

  const handleEditClose = () => setEditOpen(false);

  const handleEditSubmit = async () => {
    if (!editSlug.trim()) {
      setEditError('Repository name is required.');
      return;
    }
    setEditSubmitting(true);
    setEditError('');
    try {
      const updated = await APIClient.updateRepository(repositoryId, {
        slug: editSlug.trim(),
        description: editDescription.trim(),
        visibility: editVisibility,
      });
      setRepo((prev) => ({ ...prev, ...updated }));
      updateRepoInContext(Number(repositoryId), updated);
      setEditOpen(false);
      showNotification('Repository updated', 'success');
    } catch (err) {
      setEditError(err.message || 'Failed to update repository.');
    } finally {
      setEditSubmitting(false);
    }
  };

  // --- Loading skeleton ---
  if (loading) {
    return (
      <Container maxWidth="lg" sx={{ py: 4 }}>
        <Skeleton variant="text" width={180} height={36} sx={{ mb: 2 }} />
        <Card sx={{ p: 3, mb: 3 }}>
          <Skeleton variant="text" width="40%" height={40} />
          <Skeleton variant="text" width="70%" height={20} sx={{ mt: 1 }} />
          <Skeleton variant="text" width="30%" height={16} sx={{ mt: 1 }} />
        </Card>
        <Divider sx={{ mb: 3 }} />
        <Skeleton variant="text" width={100} height={32} sx={{ mb: 2 }} />
        {[0, 1, 2].map((i) => (
          <Skeleton key={i} variant="rectangular" height={48} sx={{ mb: 1, borderRadius: 1 }} />
        ))}
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

  const sortableColumns = [
    { id: 'name', label: 'Tag' },
    { id: 'digest', label: 'Digest', sortable: false },
    { id: 'size', label: 'Size' },
    { id: 'updatedAt', label: 'Updated' },
    { id: 'createdBy', label: 'Created By' },
  ];

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      {/* Header */}
      <Button startIcon={<ArrowBackIcon />} onClick={() => navigate('/repositories')} sx={{ mb: 2 }}>
        Back to Repositories
      </Button>

      <Card sx={{ p: 3, mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between' }}>
          <Box>
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
            {repo?.owner && (
              <Typography variant="body2" color="text.secondary">
                Owner: {repo.owner}
              </Typography>
            )}
          </Box>
          <Tooltip title="Edit repository">
            <IconButton onClick={handleEditOpen} sx={{ color: 'text.secondary' }}>
              <EditOutlinedIcon />
            </IconButton>
          </Tooltip>
        </Box>
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
          <LabelOffOutlinedIcon sx={{ fontSize: 64, color: 'primary.main', mb: 2 }} />
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
                {sortableColumns.map((col) => (
                  <TableCell key={col.id}>
                    {col.sortable === false ? (
                      <strong>{col.label}</strong>
                    ) : (
                      <TableSortLabel
                        active={sortField === col.id}
                        direction={sortField === col.id ? sortDir : 'asc'}
                        onClick={() => handleSort(col.id)}
                      >
                        <strong>{col.label}</strong>
                      </TableSortLabel>
                    )}
                  </TableCell>
                ))}
                <TableCell align="right"><strong>Actions</strong></TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {sortedTags.map((tag) => (
                <TableRow key={tag.name} hover>
                  <TableCell>
                    <Chip label={tag.name} color="primary" size="small" sx={{ borderRadius: '12px' }} />
                  </TableCell>
                  <TableCell>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                      <Typography variant="body2" sx={{ fontFamily: 'monospace' }} noWrap>
                        {tag.digest || '—'}
                      </Typography>
                      {tag.digest && (
                        <Tooltip title="Copy digest">
                          <IconButton size="small" onClick={() => handleCopyDigest(tag.digest)}>
                            <ContentCopyOutlinedIcon sx={{ fontSize: 14 }} />
                          </IconButton>
                        </Tooltip>
                      )}
                    </Box>
                  </TableCell>
                  <TableCell>
                    {formatBytes(tag.size)}
                  </TableCell>
                  <TableCell>
                    {tag.updatedAt
                      ? new Date(tag.updatedAt).toLocaleDateString()
                      : '—'}
                  </TableCell>
                  <TableCell>
                    {tag.createdBy || '—'}
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
              borderColor: isDragActive ? 'primary.main' : 'divider',
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

      {/* Edit Repository Dialog */}
      <Dialog open={editOpen} onClose={handleEditClose} maxWidth="sm" fullWidth>
        <DialogTitle>Edit Repository</DialogTitle>
        <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: '16px !important' }}>
          {editError && <Alert severity="error">{editError}</Alert>}
          <TextField
            label="Repository Name"
            value={editSlug}
            onChange={(e) => setEditSlug(e.target.value)}
            fullWidth
            autoFocus
            required
          />
          <TextField
            label="Description"
            value={editDescription}
            onChange={(e) => setEditDescription(e.target.value)}
            fullWidth
            multiline
            rows={2}
          />
          <Box>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>Visibility</Typography>
            <ToggleButtonGroup
              value={editVisibility}
              exclusive
              onChange={(_, val) => { if (val) setEditVisibility(val); }}
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
          <Button onClick={handleEditClose} disabled={editSubmitting}>Cancel</Button>
          <Button variant="contained" onClick={handleEditSubmit} disabled={editSubmitting}>
            {editSubmitting ? <CircularProgress size={20} /> : 'Save'}
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
