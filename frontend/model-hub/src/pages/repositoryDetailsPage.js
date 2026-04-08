import { useState, useEffect, useCallback, useMemo } from 'react';
import { useParams, useNavigate, Link as RouterLink } from 'react-router-dom';
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
  FormControlLabel,
  Checkbox,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TableSortLabel,
  Select,
  MenuItem,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Tooltip,
  Typography,
  Link as MuiLink,
} from '@mui/material';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import EditOutlinedIcon from '@mui/icons-material/EditOutlined';
import PublicIcon from '@mui/icons-material/Public';
import LockIcon from '@mui/icons-material/Lock';
import ContentCopyOutlinedIcon from '@mui/icons-material/ContentCopyOutlined';
import DownloadOutlinedIcon from '@mui/icons-material/DownloadOutlined';
import LabelOffOutlinedIcon from '@mui/icons-material/LabelOffOutlined';
import PersonAddIcon from '@mui/icons-material/PersonAdd';
import { useDropzone } from 'react-dropzone';
import APIClient from '../util/ApiClient';
import { useUser } from '../context/UserContext';
import { useRepositories } from '../context/RepositoryContext';
import { useNotification } from '../context/NotificationContext';

const RepositoryDetailsPage = () => {
  const { repositoryId, owner, slug } = useParams();
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
  const [addMode, setAddMode] = useState('upload'); // 'upload' | 'retag'
  const [tagName, setTagName] = useState('');
  const [tagFile, setTagFile] = useState(null);
  const [tagFileData, setTagFileData] = useState(null);
  const [sourceTagName, setSourceTagName] = useState('');
  const [addOverwrite, setAddOverwrite] = useState(false);
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

  // Fork repository dialog state
  const [forkOpen, setForkOpen] = useState(false);
  const [forkSlug, setForkSlug] = useState('');
  const [forkDescription, setForkDescription] = useState('');
  const [forkTags, setForkTags] = useState([]);
  const [forkError, setForkError] = useState('');
  const [forkSubmitting, setForkSubmitting] = useState(false);

  // Manage collaborators dialog state
  const [collaboratorsOpen, setCollaboratorsOpen] = useState(false);
  const [collaborators, setCollaborators] = useState([]);
  const [newCollabUsername, setNewCollabUsername] = useState('');
  const [newCollabRole, setNewCollabRole] = useState('read');
  const [collabError, setCollabError] = useState('');
  const [collabSubmitting, setCollabSubmitting] = useState(false);
  const [collabLoading, setCollabLoading] = useState(false);

  // Transfer ownership dialog state
  const [transferOpen, setTransferOpen] = useState(false);
  const [transferUsername, setTransferUsername] = useState('');
  const [transferError, setTransferError] = useState('');
  const [transferSubmitting, setTransferSubmitting] = useState(false);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      let repoData;
      if (repositoryId) {
        repoData = await APIClient.getRepositoryById(repositoryId);
      } else {
        repoData = await APIClient.getRepositoryByOwnerSlug(owner, slug);
      }
      setRepo(repoData);
      setTags(repoData.tags ?? []);
    } catch (err) {
      setError(err.message || 'Failed to load repository.');
    } finally {
      setLoading(false);
    }
  }, [repositoryId, owner, slug]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const activeRepoId = repo?.id ?? repositoryId;
  const backTarget = user ? '/repositories' : '/';
  const trimmedTagName = tagName.trim();
  const tagNameAlreadyExists = useMemo(
    () => trimmedTagName !== '' && tags.some((t) => t.name === trimmedTagName),
    [tags, trimmedTagName]
  );
  const canEditTags = useMemo(() => {
    if (!user || !repo) return false;
    if (user.username === repo.owner) return true;
    const myCollab = (repo.collaborators || []).find((c) => c.username === user.username);
    return myCollab?.role === 'write' || myCollab?.role === 'owner';
  }, [user, repo]);
  const canEditRepo = !!user && !!repo && user.username === repo.owner;

  // If the user changes the name to a non-existing tag, drop overwrite.
  useEffect(() => {
    if (!tagNameAlreadyExists && addOverwrite) {
      setAddOverwrite(false);
    }
  }, [tagNameAlreadyExists, addOverwrite]);

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

  const handleCopyPermalink = () => {
    if (!repo?.id) return;
    const permalink = `${window.location.origin}/repo/${repo.id}`;
    navigator.clipboard.writeText(permalink).then(() => {
      showNotification('Permalink copied to clipboard', 'info');
    });
  };

  const handleDownloadTagModel = async (tagName) => {
    try {
      const model = await APIClient.getTagModel(activeRepoId, tagName);
      const json = JSON.stringify(model, null, 2);
      const blob = new Blob([json], { type: 'application/json;charset=utf-8' });
      const url = URL.createObjectURL(blob);
      const safeRepo = (repo?.slug || 'repository').replace(/[^a-zA-Z0-9._-]/g, '_');
      const safeTag = tagName.replace(/[^a-zA-Z0-9._-]/g, '_');
      const link = document.createElement('a');
      link.href = url;
      link.download = `${safeRepo} ${safeTag}.json`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
      showNotification(`Downloaded ${tagName}`, 'success');
    } catch (err) {
      showNotification(err.message || 'Failed to download tag model.', 'error');
    }
  };

  // --- Add Tag ---
  const handleAddOpen = () => {
    setAddMode('upload');
    setTagName('');
    setSourceTagName('');
    setTagFile(null);
    setTagFileData(null);
    setAddOverwrite(false);
    setAddError('');
    setAddOpen(true);
  };

  const handleAddClose = () => setAddOpen(false);

  const onDrop = useCallback(async (acceptedFiles) => {
    if (addMode !== 'upload') return;
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
  }, [addMode]);

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
    if (tagNameAlreadyExists && !addOverwrite) {
      setAddError('Tag already exists. Enable overwrite to replace it.');
      return;
    }
    if (addMode === 'upload') {
      if (!tagFileData) {
        setAddError('Please upload a CDM JSON file.');
        return;
      }
    }
    if (addMode === 'retag') {
      if (!sourceTagName) {
        setAddError('Please select a source tag.');
        return;
      }
      if (sourceTagName === trimmedTagName) {
        setAddError('Source tag must be different from the new tag name.');
        return;
      }
    }
    setAddSubmitting(true);
    setAddError('');
    try {
      const payload =
        addMode === 'upload'
          ? tagFileData
          : { sourceTag: sourceTagName };
      await APIClient.createOrUpdateTag(activeRepoId, trimmedTagName, payload, { overwrite: addOverwrite });
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
      await APIClient.deleteTag(activeRepoId, selectedTag.name);
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
      const updated = await APIClient.updateRepository(activeRepoId, {
        slug: editSlug.trim(),
        description: editDescription.trim(),
        visibility: editVisibility,
      });
      setRepo((prev) => ({ ...prev, ...updated }));
      updateRepoInContext(Number(activeRepoId), updated);
      setEditOpen(false);
      showNotification('Repository updated', 'success');

      navigate(`/repo/${activeRepoId}`);
    } catch (err) {
      setEditError(err.message || 'Failed to update repository.');
    } finally {
      setEditSubmitting(false);
    }
  };

  // --- Fork Repository ---
  const handleForkOpen = () => {
    setForkSlug('');
    setForkDescription('');
    setForkTags([]);
    setForkError('');
    setForkOpen(true);
  };

  const handleForkClose = () => setForkOpen(false);

  const handleToggleForkTag = (tagName) => {
    setForkTags((prev) =>
      prev.includes(tagName)
        ? prev.filter((t) => t !== tagName)
        : [...prev, tagName]
    );
  };

  const handleForkSubmit = async () => {
    if (!forkSlug.trim()) {
      setForkError('Repository name is required.');
      return;
    }
    setForkSubmitting(true);
    setForkError('');
    try {
      const createdRepo = await APIClient.forkRepository(repo.owner, repo.slug, {
        slug: forkSlug.trim(),
        description: forkDescription.trim(),
        tags: forkTags,
      });
      setForkOpen(false);
      showNotification('Repository forked successfully', 'success');
      if (createdRepo?.id) {
        navigate(`/repo/${createdRepo.id}`);
      } else {
        navigate('/repositories');
      }
    } catch (err) {
      setForkError(err.message || 'Failed to fork repository.');
    } finally {
      setForkSubmitting(false);
    }
  };

  // --- Manage Collaborators ---
  const handleCollaboratorsOpen = async () => {
    setCollabLoading(true);
    setCollabError('');
    try {
      const data = await APIClient.listCollaborators(repo.owner, repo.slug);
      setCollaborators(data.collaborators || []);
      setCollaboratorsOpen(true);
    } catch (err) {
      setCollabError(err.message || 'Failed to load collaborators.');
    } finally {
      setCollabLoading(false);
    }
  };

  const handleCollaboratorsClose = () => setCollaboratorsOpen(false);

  const handleAddCollaborator = async () => {
    if (!newCollabUsername.trim()) {
      setCollabError('Username is required.');
      return;
    }
    setCollabSubmitting(true);
    setCollabError('');
    try {
      const result = await APIClient.addCollaborator(
        repo.owner,
        repo.slug,
        newCollabUsername.trim(),
        newCollabRole
      );
      setCollaborators((prev) => {
        const existing = prev.findIndex((c) => c.username === newCollabUsername.trim());
        if (existing >= 0) {
          const updated = [...prev];
          updated[existing] = result;
          return updated;
        }
        return [...prev, result];
      });
      setNewCollabUsername('');
      setNewCollabRole('read');
      showNotification('Collaborator added', 'success');
    } catch (err) {
      setCollabError(err.message || 'Failed to add collaborator.');
    } finally {
      setCollabSubmitting(false);
    }
  };

  const handleRemoveCollaborator = async (username) => {
    if (!window.confirm(`Remove ${username} from this repository?`)) return;
    try {
      await APIClient.removeCollaborator(repo.owner, repo.slug, username);
      setCollaborators((prev) => prev.filter((c) => c.username !== username));
      showNotification('Collaborator removed', 'success');
    } catch (err) {
      showNotification(err.message || 'Failed to remove collaborator.', 'error');
    }
  };

  // --- Transfer Ownership ---
  const handleTransferOpen = () => {
    setTransferUsername('');
    setTransferError('');
    setTransferOpen(true);
  };

  const handleTransferClose = () => setTransferOpen(false);

  const handleTransferSubmit = async () => {
    if (!transferUsername.trim()) {
      setTransferError('Username is required.');
      return;
    }
    if (transferUsername.trim() === repo.owner) {
      setTransferError('Cannot transfer to the current owner.');
      return;
    }
    setTransferSubmitting(true);
    setTransferError('');
    try {
      await APIClient.transferRepositoryOwnership(repo.owner, repo.slug, transferUsername.trim());
      setTransferOpen(false);
      showNotification('Repository ownership transferred', 'success');
      navigate('/repositories');
    } catch (err) {
      setTransferError(err.message || 'Failed to transfer repository.');
    } finally {
      setTransferSubmitting(false);
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
        <Button startIcon={<ArrowBackIcon />} onClick={() => navigate(backTarget)}>
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
      <Button startIcon={<ArrowBackIcon />} onClick={() => navigate(backTarget)} sx={{ mb: 2 }}>
        Back to Repositories
      </Button>

      <Card sx={{ p: 3, mb: 3 }}>
        <Box sx={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 3 }}>
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
                  sx={{
                    color: repo.visibility === 'public' ? 'success.main' : 'text.secondary',
                    borderColor: repo.visibility === 'public' ? 'success.main' : 'text.secondary',
                    '& .MuiChip-icon': {
                      color: repo.visibility === 'public' ? 'success.main' : 'text.secondary',
                    },
                  }}
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
                Owner:{' '}
                <MuiLink
                  component={RouterLink}
                  to={`/repositories/${encodeURIComponent(repo.owner)}`}
                  underline="hover"
                >
                  {repo.owner}
                </MuiLink>
              </Typography>
            )}
          </Box>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
            <Tooltip title="Copy permalink">
              <IconButton onClick={handleCopyPermalink} sx={{ color: 'text.secondary' }}>
                <ContentCopyOutlinedIcon />
              </IconButton>
            </Tooltip>
            {canEditRepo && (
              <Tooltip title="Edit repository">
                <IconButton onClick={handleEditOpen} sx={{ color: 'text.secondary' }}>
                  <EditOutlinedIcon />
                </IconButton>
              </Tooltip>
            )}
          </Box>
        </Box>
        <Box sx={{ display: 'flex', gap: 1, alignItems: 'center', mt: 2 }}>
          {user && user.username === repo?.owner && (
            <>
              <Button size="small" onClick={handleForkOpen} variant="outlined">
                Fork
              </Button>
              <Button size="small" onClick={handleCollaboratorsOpen} variant="outlined">
                Collaborators
              </Button>
              <Button size="small" onClick={handleTransferOpen} variant="outlined" color="primary">
                Transfer
              </Button>
            </>
          )}
          {user && user.username !== repo?.owner && (
            <Button size="small" onClick={handleForkOpen} variant="outlined">
              Fork
            </Button>
          )}
        </Box>
      </Card>

      <Divider sx={{ mb: 3 }} />

      {/* Tags Section */}
      <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
        <Typography variant="h5" fontWeight={600}>
          Tags
        </Typography>
        <Box sx={{ display: 'flex', gap: 1 }}>
          <Tooltip title="Compare tags coming soon">
            <span>
              <Button variant="outlined" disabled>
                Compare tags
              </Button>
            </span>
          </Tooltip>
          {canEditTags && (
              <Button variant="contained" startIcon={<AddIcon />} onClick={handleAddOpen}>
                Add Tag
              </Button>
            )}
        </Box>
        
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
                    {tag.createdBy ? (
                      <MuiLink
                        component={RouterLink}
                        to={`/repositories/${encodeURIComponent(tag.createdBy)}`}
                        underline="hover"
                      >
                        {tag.createdBy}
                      </MuiLink>
                    ) : '—'}
                  </TableCell>
                  <TableCell align="right">
                    <Tooltip title="Download model">
                      <IconButton
                        size="small"
                        onClick={() => handleDownloadTagModel(tag.name)}
                        sx={{ color: 'text.secondary', mr: 0.5, '&:hover': { color: 'primary.main' } }}
                      >
                        <DownloadOutlinedIcon fontSize="small" />
                      </IconButton>
                    </Tooltip>
                    {canEditTags && (
                      <IconButton
                        size="small"
                        onClick={() => handleDeleteOpen(tag)}
                        sx={{ color: 'text.secondary', '&:hover': { color: 'error.main' } }}
                      >
                        <DeleteOutlineIcon fontSize="small" />
                      </IconButton>
                    )}
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
          <ToggleButtonGroup
            value={addMode}
            exclusive
            onChange={(_, next) => {
              if (!next) return;
              setAddMode(next);
              setAddError('');
              // Reset inputs for clarity
              setTagFile(null);
              setTagFileData(null);
            }}
          >
            <ToggleButton value="upload" sx={{ textTransform: 'none' }}>
              Upload new
            </ToggleButton>
            <ToggleButton value="retag" sx={{ textTransform: 'none' }}>
              Use existing tag
            </ToggleButton>
          </ToggleButtonGroup>
          {addMode === 'retag' && (
            <TextField
              select
              label="Source tag"
              value={sourceTagName}
              onChange={(e) => setSourceTagName(e.target.value)}
              fullWidth
              required
              disabled={tags.length === 0}
              helperText={
                tags.length === 0
                  ? 'No tags available to retag.'
                  : 'The new tag will point to the same model as the selected source tag (one-time copy).'
              }
            >
              <MenuItem value="" disabled>
                Select a tag
              </MenuItem>
              {tags.map((t) => (
                <MenuItem key={t.name} value={t.name}>
                  {t.name}
                </MenuItem>
              ))}
            </TextField>
          )}
{addMode === 'upload' && (
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
            )}
            {tagNameAlreadyExists && (
            <>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={addOverwrite}
                    onChange={(e) => setAddOverwrite(e.target.checked)}
                  />
                }
                label="Overwrite existing tag (replaces the model this tag points to)"
              />
              {addOverwrite && (
                <Alert severity="warning">
                  This will overwrite the existing tag <strong>{trimmedTagName}</strong>. Existing references to this
                  tag will now point to the new model.
                </Alert>
              )}
            </>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleAddClose} disabled={addSubmitting}>Cancel</Button>
          <Button variant="contained" onClick={handleAddSubmit} disabled={addSubmitting}>
            {addSubmitting ? <CircularProgress size={20} /> : addMode === 'upload' ? 'Upload & Create Tag' : 'Create Tag From Existing'}
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
                <PublicIcon fontSize="small" sx={{ mr: 0.5, color: 'success.main' }} /> Public
              </ToggleButton>
              <ToggleButton value="private">
                <LockIcon fontSize="small" sx={{ mr: 0.5, color: 'text.secondary' }} /> Private
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

      {/* Fork Repository Dialog */}
      <Dialog open={forkOpen} onClose={handleForkClose} maxWidth="sm" fullWidth>
        <DialogTitle>Fork Repository</DialogTitle>
        <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: '16px !important' }}>
          {forkError && <Alert severity="error">{forkError}</Alert>}
          <TextField
            label="New Repository Name"
            value={forkSlug}
            onChange={(e) => setForkSlug(e.target.value)}
            fullWidth
            autoFocus
            required
            placeholder="e.g. my-fork"
          />
          <TextField
            label="Description"
            value={forkDescription}
            onChange={(e) => setForkDescription(e.target.value)}
            fullWidth
            multiline
            rows={2}
          />
          <Box>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>Copy Tags (optional)</Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
              Select which tags to copy from the source repository:
            </Typography>
            {tags.length === 0 ? (
              <Typography variant="body2" color="text.secondary">No tags available</Typography>
            ) : (
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                {tags.map((tag) => (
                  <FormControlLabel
                  key={tag.name}
                  control={
                    <Checkbox
                      size="small"
                      checked={forkTags.includes(tag.name)}
                      onChange={() => handleToggleForkTag(tag.name)}
                    />
                  }
                  label={
                    <Typography variant="body2">
                      {tag.name}
                    </Typography>
                  }
                  sx={{ m: 0 }}
                />
              ))}
            </Box>
          )}
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleForkClose} disabled={forkSubmitting}>Cancel</Button>
        <Button variant="contained" onClick={handleForkSubmit} disabled={forkSubmitting}>
          {forkSubmitting ? <CircularProgress size={20} /> : 'Fork'}
        </Button>
      </DialogActions>
    </Dialog>

      {/* Manage Collaborators Dialog */}
      <Dialog open={collaboratorsOpen} onClose={handleCollaboratorsClose} maxWidth="sm" fullWidth>
        <DialogTitle>Manage Collaborators</DialogTitle>
        <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: '16px !important', maxHeight: '60vh', overflow: 'auto' }}>
          {collabError && <Alert severity="error">{collabError}</Alert>}
          
          {/* Add Collaborator Section */}
          <Box sx={{ borderBottom: 1, borderColor: 'divider', pb: 2 }}>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>Add Collaborator</Typography>
            <Box sx={{ display: 'flex', gap: 1, alignItems: 'flex-end' }}>
              <TextField
                label="Username"
                value={newCollabUsername}
                onChange={(e) => setNewCollabUsername(e.target.value)}
                size="small"
                sx={{ flex: 1 }}
                placeholder="Enter username"
              />
              <ToggleButtonGroup
                value={newCollabRole}
                exclusive
                onChange={(_, val) => { if (val) setNewCollabRole(val); }}
                size="small"
              >
                <ToggleButton value="read">Read</ToggleButton>
                <ToggleButton value="write">Write</ToggleButton>
              </ToggleButtonGroup>
              <Button
                variant="contained"
                size="small"
                onClick={handleAddCollaborator}
                disabled={collabSubmitting}
                startIcon={<PersonAddIcon />}
              >
                Add
              </Button>
            </Box>
          </Box>

          {/* Collaborators List */}
          <Box>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>Current Collaborators</Typography>
            {collabLoading ? (
              <CircularProgress size={24} />
            ) : collaborators.length === 0 ? (
              <Typography variant="body2" color="text.secondary">No collaborators yet</Typography>
            ) : (
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                {collaborators.map((collab) => (
                  <Box
                    key={collab.username}
                    sx={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      p: 1,
                      backgroundColor: 'action.hover',
                      borderRadius: 1,
                    }}
                  >
                    <Box>
                      <Typography variant="body2" fontWeight={500}>{collab.username}</Typography>
                      <Typography variant="caption" color="text.secondary">
                        {collab.role}
                      </Typography>
                    </Box>
                    <Button
                      size="small"
                      color="error"
                      onClick={() => handleRemoveCollaborator(collab.username)}
                    >
                      Remove
                    </Button>
                  </Box>
                ))}
              </Box>
            )}
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCollaboratorsClose}>Close</Button>
        </DialogActions>
      </Dialog>

      {/* Transfer Ownership Dialog */}
      <Dialog open={transferOpen} onClose={handleTransferClose} maxWidth="sm" fullWidth>
        <DialogTitle>Transfer Repository Ownership</DialogTitle>
        <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: '16px !important' }}>
          {transferError && <Alert severity="error">{transferError}</Alert>}
          <Alert severity="warning">
            This action is permanent. After transfer, you will lose ownership of this repository.
          </Alert>
          <TextField
            label="New Owner Username"
            value={transferUsername}
            onChange={(e) => setTransferUsername(e.target.value)}
            fullWidth
            autoFocus
            required
            placeholder="Enter username of new owner"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleTransferClose} disabled={transferSubmitting}>Cancel</Button>
          <Button variant="contained" color="error" onClick={handleTransferSubmit} disabled={transferSubmitting}>
            {transferSubmitting ? <CircularProgress size={20} /> : 'Transfer'}
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
