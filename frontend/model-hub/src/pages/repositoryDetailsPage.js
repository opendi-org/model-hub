import { useState, useEffect, useCallback, useMemo } from 'react';
import { useParams, useNavigate, Link as RouterLink, useLocation } from 'react-router-dom';
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
  const location = useLocation();
  const { user } = useUser();
  const { updateRepository: updateRepoInContext, refreshRepositories } = useRepositories();
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
  // When set, we are using the Add Tag dialog to overwrite an existing tag.
  // In this mode the tag name is locked and the overwrite checkbox is hidden.
  const [tagEditTarget, setTagEditTarget] = useState(null);

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
  const [forkVisibility, setForkVisibility] = useState('private');
  const [forkTags, setForkTags] = useState([]);
  const [forkError, setForkError] = useState('');
  const [forkSubmitting, setForkSubmitting] = useState(false);
  const [forkPublicConfirmOpen, setForkPublicConfirmOpen] = useState(false);

  // Manage collaborators dialog state
  const [collaboratorsOpen, setCollaboratorsOpen] = useState(false);
  const [collaborators, setCollaborators] = useState([]);
  const [newCollabUsername, setNewCollabUsername] = useState('');
  const [newCollabRole, setNewCollabRole] = useState('read');
  const [collabError, setCollabError] = useState('');
  const [collabSubmitting, setCollabSubmitting] = useState(false);
  const [collabLoading, setCollabLoading] = useState(false);
  const [removeCollabConfirmOpen, setRemoveCollabConfirmOpen] = useState(false);
  const [removeCollabUsername, setRemoveCollabUsername] = useState('');

  // Edit collaborator role state
  const [editCollabSubmitting, setEditCollabSubmitting] = useState(false);

  // Transfer ownership dialog state
  const [transferOpen, setTransferOpen] = useState(false);
  const [transferUsername, setTransferUsername] = useState('');
  const [transferError, setTransferError] = useState('');
  const [transferSubmitting, setTransferSubmitting] = useState(false);
  const [transferConfirmOpen, setTransferConfirmOpen] = useState(false);
  const [transferPreviousOwnerAccess, setTransferPreviousOwnerAccess] = useState('admin');

  // View lineage dialog state
  const [lineageOpen, setLineageOpen] = useState(false);
  const [lineageData, setLineageData] = useState(null);
  const [lineageLoading, setLineageLoading] = useState(false);
  const [lineageError, setLineageError] = useState('');

  // Public visibility confirmation dialog state
  const [editPublicConfirmOpen, setEditPublicConfirmOpen] = useState(false);

  // Compare tags dialog state
  const [compareOpen, setCompareOpen] = useState(false);
  const [compareLeft, setCompareLeft] = useState('');
  const [compareRight, setCompareRight] = useState('');
  const [compareResult, setCompareResult] = useState(null); // null | { lines: string[], identical: bool }
  const [compareError, setCompareError] = useState('');
  const [compareLoading, setCompareLoading] = useState(false);

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
  const trimmedTagName = tagName.trim();
  const tagNameAlreadyExists = useMemo(
    () => trimmedTagName !== '' && tags.some((t) => t.name === trimmedTagName),
    [tags, trimmedTagName]
  );
  const canEditTags = useMemo(() => {
    if (!user || !repo) return false;
    if (user.username === repo.owner) return true;
    const myCollab = (repo.collaborators || []).find((c) => c.username === user.username);
    return myCollab?.role === 'write' || myCollab?.role === 'admin' || myCollab?.role === 'owner';
  }, [user, repo]);
  const canEditRepo = !!user && !!repo && user.username === repo.owner;

  // Get current user's collaborator role (for permission checks in collaborators dialog)
  const currentUserCollabRole = useMemo(() => {
    if (!user || !collaborators) return null;
    const collab = collaborators.find((c) => c.username === user.username);
    return collab?.role || null;
  }, [user, collaborators]);

  // Check if current user is a collaborator on this repo
  const isCurrentUserCollaborator = useMemo(() => {
    if (!user || !repo) return false;
    if (user.username === repo.owner) return true;
    return (repo.collaborators || []).some((c) => c.username === user.username);
  }, [user, repo]);

  const getRepoSlugValidationError = (value) => {
    const slug = (value || '').trim();
    if (!slug) return 'Repository name is required.';
    if (slug.length > 255) return 'Repository name must be 255 characters or less.';
    if (/\s/.test(slug)) return 'Repository name cannot contain spaces.';
    if (!/^[A-Za-z0-9_-]+$/.test(slug)) {
      return "Repository name may contain only letters, numbers, '-' and '_' (no spaces).";
    }
    return '';
  };

  const getTagNameValidationError = (value) => {
    const name = (value || '').trim();
    if (!name) return 'Tag name is required.';
    if (name.length > 255) return 'Tag name must be 255 characters or less.';
    if (/\s/.test(name)) return 'Tag name cannot contain spaces.';
    if (!/^[A-Za-z0-9._-]+$/.test(name)) {
      return "Tag name may contain only letters, numbers, '.', '-', and '_' (no spaces).";
    }
    return '';
  };

  // If the user changes the name to a non-existing tag, drop overwrite.
  useEffect(() => {
    if (!tagNameAlreadyExists && addOverwrite && !tagEditTarget) {
      setAddOverwrite(false);
    }
  }, [tagNameAlreadyExists, addOverwrite, tagEditTarget]);

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
    const permalink = `${window.location.origin}/repo/${encodeURIComponent(repo.id)}`;
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
    setTagEditTarget(null);
    setTagName('');
    setSourceTagName('');
    setTagFile(null);
    setTagFileData(null);
    setAddOverwrite(false);
    setAddError('');
    setAddOpen(true);
  };

  const handleEditTagOpen = (tag) => {
    const targetName = tag?.name;
    if (!targetName) return;

    setAddMode('upload');
    setTagEditTarget(targetName);
    setTagName(targetName);
    setSourceTagName('');
    setTagFile(null);
    setTagFileData(null);
    setAddOverwrite(true); // forced overwrite (no checkbox needed)
    setAddError('');
    setAddOpen(true);
  };

  const handleAddClose = () => {
    setAddOpen(false);
    setTagEditTarget(null);
  };

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
    const validationError = getTagNameValidationError(tagName);
    if (validationError) {
      setAddError(validationError);
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
      const msg = err?.message || '';
      if (msg.toLowerCase().includes('invalid tag name')) {
        setAddError("Invalid tag name. Use 1-255 characters with letters/numbers and only '-', '_', and '.' (no spaces).");
      } else if (msg.toLowerCase().includes('tag already exists')) {
        setAddError('Tag already exists. Enable overwrite to replace it.');
      } else {
        setAddError(msg || 'Failed to create tag.');
      }
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
    const validationError = getRepoSlugValidationError(editSlug);
    if (validationError) {
      setEditError(validationError);
      return;
    }

    // Show confirmation if changing to public
    if (editVisibility === 'public' && repo.visibility !== 'public') {
      setEditPublicConfirmOpen(true);
      return;
    }

    await performEdit();
  };

  const performEdit = async () => {
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
      await refreshRepositories();
      setEditOpen(false);
      setEditPublicConfirmOpen(false);
      showNotification('Repository updated', 'success');

      // Always replace history entry on edit to avoid "back button shows same page" issue
      navigate(`/repositories/${updated.owner}/${updated.slug}`, {
        state: { from: location?.state?.from },
        replace: true,
      });
    } catch (err) { 
      const msg = err?.message || '';
      if (msg.toLowerCase().includes('invalid repository name') || msg.toLowerCase().includes('invalid slug')) {
        setEditError("Invalid repository name. Use 1-255 characters with letters/numbers and only '-' or '_' (no spaces).");
      } else {
        setEditError(msg || 'Failed to update repository.');
      }
      setEditPublicConfirmOpen(false);
    } finally {
      setEditSubmitting(false);
    }
  };

  // --- Fork Repository ---
  const handleForkOpen = () => {
    setForkSlug('');
    setForkDescription('');
    setForkVisibility('private');
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
    const validationError = getRepoSlugValidationError(forkSlug);
    if (validationError) {
      setForkError(validationError);
      return;
    }
    // Show confirmation if making repository public
    if (forkVisibility === 'public') {
      setForkPublicConfirmOpen(true);
      return;
    }
    setForkSubmitting(true);
    setForkError('');
    try {
      const createdRepo = await APIClient.forkRepository(repo.owner, repo.slug, {
        slug: forkSlug.trim(),
        description: forkDescription.trim(),
        visibility: forkVisibility,
        tags: forkTags,
      });
      setForkOpen(false);
      showNotification('Repository forked successfully', 'success');
      // Navigate to the newly created fork
      navigate(`/repositories/${createdRepo.owner}/${createdRepo.slug}`, { state: { from: location.pathname } });
    } catch (err) {
      const msg = err?.message || '';
      if (msg.toLowerCase().includes('invalid repository name') || msg.toLowerCase().includes('invalid slug')) {
        setForkError("Invalid repository name. Use 1-255 characters with letters/numbers and only '-' or '_' (no spaces).");
      } else {
        setForkError(msg || 'Failed to fork repository.');
      }
    } finally {
      setForkSubmitting(false);
    }
  };

  const handleForkPublicConfirm = async () => {
    setForkPublicConfirmOpen(false);
    setForkSubmitting(true);
    setForkError('');
    try {
      const createdRepo = await APIClient.forkRepository(repo.owner, repo.slug, {
        slug: forkSlug.trim(),
        description: forkDescription.trim(),
        visibility: forkVisibility,
        tags: forkTags,
      });
      setForkOpen(false);
      showNotification('Repository forked successfully', 'success');
      // Navigate to the newly created fork
      navigate(`/repositories/${createdRepo.owner}/${createdRepo.slug}`, { state: { from: location.pathname } });
    } catch (err) {
      const msg = err?.message || '';
      if (msg.toLowerCase().includes('invalid repository name') || msg.toLowerCase().includes('invalid slug')) {
        setForkError("Invalid repository name. Use 1-255 characters with letters/numbers and only '-' or '_' (no spaces).");
      } else {
        setForkError(msg || 'Failed to fork repository.');
      }
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
      // Also refresh the main repo data to ensure collaborators list is up-to-date
      const repoData = await APIClient.getRepositoryByOwnerSlug(repo.owner, repo.slug);
      setRepo(repoData);
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
    if (newCollabUsername.trim() === repo?.owner) {
      setCollabError('Cannot add the repository owner as a collaborator.');
      return;
    }
    if (newCollabUsername.trim() === user?.username) {
      setCollabError('Cannot share repository with yourself.');
      return;
    }
    setCollabSubmitting(true);
    setCollabError('');
    try {
      await APIClient.addCollaborator(
        repo.owner,
        repo.slug,
        newCollabUsername.trim(),
        newCollabRole
      );
      setNewCollabUsername('');
      setNewCollabRole('read');
      showNotification('Collaborator added', 'success');
      // Refresh collaborators from server to ensure consistency
      await refreshCollaborators();
    } catch (err) {
      setCollabError(err.message || 'Failed to add collaborator.');
    } finally {
      setCollabSubmitting(false);
    }
  };

  const handleRemoveCollaborator = async (username) => {
    setRemoveCollabUsername(username);
    setRemoveCollabConfirmOpen(true);
  };

  const handleRemoveCollaboratorConfirm = async () => {
    setRemoveCollabConfirmOpen(false);
    const isRemovingSelf = user && user.username === removeCollabUsername;
    
    try {
      setCollabSubmitting(true);
      await APIClient.removeCollaborator(repo.owner, repo.slug, removeCollabUsername);
      const successMsg = isRemovingSelf 
        ? 'You have left the repository' 
        : 'Collaborator removed';
      showNotification(successMsg, 'success');
      if (isRemovingSelf) {
        navigate('/repositories');
      } else {
        // Refresh collaborators from server to ensure consistency for other users viewing
        await refreshCollaborators();
      }
    } catch (err) {
      showNotification(err.message || 'Failed to remove collaborator.', 'error');
    } finally {
      setCollabSubmitting(false);
      setRemoveCollabUsername('');
    }
  };

  const handleRemoveCollaboratorCancel = () => {
    setRemoveCollabConfirmOpen(false);
    setRemoveCollabUsername('');
  };

  // Helper to refresh collaborators from server
  const refreshCollaborators = async () => {
    try {
      const data = await APIClient.listCollaborators(repo.owner, repo.slug);
      setCollaborators(data.collaborators || []);
      // Also update repo's collaborators
      setRepo((prev) => ({
        ...prev,
        collaborators: data.collaborators || [],
      }));
    } catch (err) {
      console.error('Failed to refresh collaborators:', err);
    }
  };

  const handleUpdateCollaboratorRole = async (username, newRole) => {
    setEditCollabSubmitting(true);
    try {
      await APIClient.addCollaborator(repo.owner, repo.slug, username, newRole);
      showNotification(`${username}'s role updated to ${newRole}`, 'success');
      // Refresh collaborators from server to ensure consistency
      await refreshCollaborators();
    } catch (err) {
      showNotification(err.message || 'Failed to update collaborator role.', 'error');
    } finally {
      setEditCollabSubmitting(false);
    }
  };

  // --- Transfer Ownership ---
  const handleTransferOpen = () => {
    setTransferUsername('');
    setTransferError('');
    setTransferPreviousOwnerAccess('admin');
    setTransferOpen(true);
  };

  const handleTransferClose = () => setTransferOpen(false);

  const handleTransferSubmit = () => {
    if (!transferUsername.trim()) {
      setTransferError('Username is required.');
      return;
    }
    if (transferUsername.trim() === repo.owner) {
      setTransferError('Cannot transfer to the current owner.');
      return;
    }
    setTransferError('');
    // Show confirmation dialog instead of immediately transferring
    setTransferConfirmOpen(true);
  };

  const handleTransferConfirm = async () => {
    setTransferConfirmOpen(false);
    setTransferSubmitting(true);
    setTransferError('');
    try {
      // Transfer with previous owner access level - backend handles adding collaborator atomically
      await APIClient.transferRepositoryOwnership(
        repo.owner, 
        repo.slug, 
        transferUsername.trim(),
        transferPreviousOwnerAccess
      );
      
      setTransferOpen(false);
      showNotification('Repository ownership transferred', 'success');
      await refreshRepositories();
      if (transferPreviousOwnerAccess === 'none') {
        navigate('/repositories');
      } else {
        navigate(-1);
      }
    } catch (err) {
      setTransferError(err.message || 'Failed to transfer repository.');
      setTransferSubmitting(false);
    }
  };

  const handleTransferConfirmClose = () => setTransferConfirmOpen(false);

  // --- View Lineage ---
  const handleLineageOpen = async () => {
    setLineageOpen(true);
    setLineageLoading(true);
    setLineageError('');
    setLineageData(null);
    try {
      const data = await APIClient.getRepositoryLineage(repo.owner, repo.slug);
      setLineageData(data);
    } catch (err) {
      setLineageError(err.message || 'Failed to load lineage.');
    } finally {
      setLineageLoading(false);
    }
  };

  const handleLineageClose = () => setLineageOpen(false);

  // --- Compare tags ---
  const handleCompareOpen = () => {
    setCompareLeft(tags.length > 0 ? tags[0].name : '');
    setCompareRight(tags.length > 1 ? tags[1].name : '');
    setCompareResult(null);
    setCompareError('');
    setCompareOpen(true);
  };

  const handleCompareClose = () => {
    setCompareOpen(false);
    setCompareResult(null);
    setCompareError('');
  };

  const handleCompareRun = async () => {
    if (!compareLeft || !compareRight) {
      setCompareError('Please select two tags.');
      return;
    }
    setCompareLoading(true);
    setCompareError('');
    setCompareResult(null);
    try {
      const [leftModel, rightModel] = await Promise.all([
        APIClient.getTagModel(activeRepoId, compareLeft),
        APIClient.getTagModel(activeRepoId, compareRight),
      ]);
      const sortedReplacer = (_, val) =>
        val && typeof val === 'object' && !Array.isArray(val)
          ? Object.fromEntries(Object.entries(val).sort(([a], [b]) => a.localeCompare(b)))
          : val;
      const leftStr = JSON.stringify(leftModel, sortedReplacer, 2) + '\n';
      const rightStr = JSON.stringify(rightModel, sortedReplacer, 2) + '\n';
      if (leftStr === rightStr) {
        setCompareResult({ identical: true, lines: [] });
      } else {
        const leftLines = leftStr.split('\n');
        const rightLines = rightStr.split('\n');

        // Build LCS-based diff
        const CONTEXT = 3;
        const m = leftLines.length, n = rightLines.length;
        // dp[i][j] = LCS length of leftLines[0..i-1] vs rightLines[0..j-1]
        const dp = Array.from({ length: m + 1 }, () => new Int32Array(n + 1));
        for (let i = 1; i <= m; i++)
          for (let j = 1; j <= n; j++)
            dp[i][j] = leftLines[i-1] === rightLines[j-1]
              ? dp[i-1][j-1] + 1
              : Math.max(dp[i-1][j], dp[i][j-1]);

        // Backtrack to get edit script
        const edits = [];
        let i = m, j = n;
        while (i > 0 || j > 0) {
          if (i > 0 && j > 0 && leftLines[i-1] === rightLines[j-1]) {
            edits.push({ type: 'context', text: leftLines[i-1] });
            i--; j--;
          } else if (j > 0 && (i === 0 || dp[i][j-1] >= dp[i-1][j])) {
            edits.push({ type: 'add', text: rightLines[j-1] });
            j--;
          } else {
            edits.push({ type: 'remove', text: leftLines[i-1] });
            i--;
          }
        }
        edits.reverse();

        // Collapse context: only keep CONTEXT lines around changes
        const changed = edits.map((e) => e.type !== 'context');
        const visible = edits.map((_, idx) => {
          if (changed[idx]) return true;
          for (let d = 1; d <= CONTEXT; d++) {
            if (changed[idx - d] || changed[idx + d]) return true;
          }
          return false;
        });

        const lines = [];
        for (let idx = 0; idx < edits.length; idx++) {
          if (!visible[idx]) {
            if (idx === 0 || visible[idx - 1]) lines.push({ type: 'separator', text: '...' });
          } else {
            lines.push(edits[idx]);
          }
        }
        setCompareResult({ identical: false, lines });
      }
    } catch (err) {
      setCompareError(err.message || 'Failed to load models for comparison.');
    } finally {
      setCompareLoading(false);
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
        <Button startIcon={<ArrowBackIcon />} onClick={() => navigate(-1)}>
          Back
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
      <Button startIcon={<ArrowBackIcon />} onClick={() => navigate(-1)} sx={{ mb: 2 }}>
        Back
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
          {/* View Lineage - available to all users */}
          <Button size="small" onClick={handleLineageOpen} variant="outlined">
            View Lineage
          </Button>

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
            <>
              <Button size="small" onClick={handleForkOpen} variant="outlined">
                Fork
              </Button>
              {isCurrentUserCollaborator && (
                <Button size="small" onClick={handleCollaboratorsOpen} variant="outlined">
                  Collaborators
                </Button>
              )}
            </>
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
          <Button
            variant="outlined"
            onClick={handleCompareOpen}
            disabled={tags.length < 2}
          >
            Compare tags
          </Button>
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
                    <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'flex-end', gap: 0.5 }}>
                      {canEditTags && (
                        <Tooltip title="Edit tag">
                          <IconButton
                            size="small"
                            onClick={() => handleEditTagOpen(tag)}
                            sx={{ color: 'text.secondary', '&:hover': { color: 'warning.main' } }}
                          >
                            <EditOutlinedIcon fontSize="small" />
                          </IconButton>
                        </Tooltip>
                      )}
                      {canEditTags && (
                        <IconButton
                          size="small"
                          onClick={() => handleDeleteOpen(tag)}
                          sx={{ color: 'text.secondary', '&:hover': { color: 'error.main' } }}
                        >
                          <DeleteOutlineIcon fontSize="small" />
                        </IconButton>
                      )}
                      <Tooltip title="Download model">
                        <IconButton
                          size="small"
                          onClick={() => handleDownloadTagModel(tag.name)}
                          sx={{ color: 'primary.main', '&:hover': { color: 'primary.main' } }}
                        >
                          <DownloadOutlinedIcon fontSize="small" />
                        </IconButton>
                      </Tooltip>
                    </Box>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {/* Add Tag Dialog */}
      <Dialog open={addOpen} onClose={handleAddClose} maxWidth="sm" fullWidth>
        <DialogTitle>{tagEditTarget ? 'Edit Tag' : 'Add Tag'}</DialogTitle>
        <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: '16px !important' }}>
          {addError && <Alert severity="error">{addError}</Alert>}
          <TextField
            label="Tag Name"
            value={tagName}
            onChange={(e) => setTagName(e.target.value)}
            fullWidth
            autoFocus
            required
            disabled={!!tagEditTarget}
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
            {tagNameAlreadyExists && !tagEditTarget && (
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
            {tagNameAlreadyExists && tagEditTarget && (
              <Alert severity="warning">
                This will overwrite <strong>{trimmedTagName}</strong>.
              </Alert>
            )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleAddClose} disabled={addSubmitting}>Cancel</Button>
          <Button variant="contained" onClick={handleAddSubmit} disabled={addSubmitting}>
            {addSubmitting ? <CircularProgress size={20} /> : tagEditTarget ? 'Update Tag' : (addMode === 'upload' ? 'Upload & Create Tag' : 'Create Tag From Existing')}
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

      {/* Public Visibility Confirmation Dialog */}
      <Dialog open={editPublicConfirmOpen} onClose={() => setEditPublicConfirmOpen(false)} maxWidth="sm" fullWidth>
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
          <Button onClick={() => setEditPublicConfirmOpen(false)} disabled={editSubmitting}>
            Cancel
          </Button>
          <Button variant="contained" color="warning" onClick={performEdit} disabled={editSubmitting}>
            {editSubmitting ? <CircularProgress size={20} /> : 'Make Public'}
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
        <Box>
          <Typography variant="subtitle2" sx={{ mb: 1 }}>Visibility</Typography>
          <ToggleButtonGroup
            value={forkVisibility}
            exclusive
            onChange={(_, val) => { if (val) setForkVisibility(val); }}
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
        <Button onClick={handleForkClose} disabled={forkSubmitting}>Cancel</Button>
        <Button variant="contained" onClick={handleForkSubmit} disabled={forkSubmitting}>
          {forkSubmitting ? <CircularProgress size={20} /> : 'Fork'}
        </Button>
      </DialogActions>
    </Dialog>

      {/* Fork Public Visibility Confirmation Dialog */}
      <Dialog open={forkPublicConfirmOpen} onClose={() => setForkPublicConfirmOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Make Forked Repository Public?</DialogTitle>
        <DialogContent>
          <Typography variant="body2" sx={{ mt: 2 }}>
            The forked repository will be visible to everyone. Anyone can view and download the repository contents.
          </Typography>
          <Typography variant="body2" sx={{ mt: 2, fontWeight: 500 }}>
            Are you sure you want to continue?
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setForkPublicConfirmOpen(false)} disabled={forkSubmitting}>
            Cancel
          </Button>
          <Button variant="contained" color="warning" onClick={handleForkPublicConfirm} disabled={forkSubmitting}>
            {forkSubmitting ? <CircularProgress size={20} /> : 'Make Public'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Manage Collaborators Dialog */}
      <Dialog open={collaboratorsOpen} onClose={handleCollaboratorsClose} maxWidth="sm" fullWidth>
        <DialogTitle>
          {canEditRepo || currentUserCollabRole === 'admin' ? 'Manage Collaborators' : 'Collaborators'}
        </DialogTitle>
        <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: '16px !important', maxHeight: '60vh', overflow: 'auto' }}>
          {collabError && <Alert severity="error">{collabError}</Alert>}
          
          {/* Add Collaborator Section - Only visible to owners and admin-level collaborators */}
          {(canEditRepo || currentUserCollabRole === 'admin') && (
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
                  {/* Only owners can grant admin role */}
                  {canEditRepo && <ToggleButton value="admin">Admin</ToggleButton>}
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
          )}

          {/* Collaborators List */}
          <Box>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>Current Collaborators</Typography>
            {collabLoading ? (
              <CircularProgress size={24} />
            ) : collaborators.length === 0 ? (
              <Typography variant="body2" color="text.secondary">No collaborators yet</Typography>
            ) : (
              <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                {collaborators.map((collab) => {
                  const isCurrentUser = user && user.username === collab.username;
                  const isOwner = collab.username === repo?.owner;
                  const isAdmin = collab.role === 'admin';
                  // Can remove if: removing self, owner removing anyone, or admin removing non-owner non-admins
                  const canRemove = isCurrentUser || canEditRepo || (currentUserCollabRole === 'admin' && !isOwner && !isAdmin);
                  const canEditRole = (canEditRepo || currentUserCollabRole === 'admin') && !isCurrentUser && !isOwner && (!isAdmin || canEditRepo);
                  const buttonLabel = isCurrentUser ? 'Leave' : 'Remove';
                  const buttonTooltip = isCurrentUser 
                    ? 'Remove yourself as a collaborator' 
                    : isOwner
                    ? 'Cannot remove the repository owner'
                    : !canEditRepo && isAdmin
                    ? 'Only the owner can remove admins'
                    : 'Remove this collaborator';

                  return (
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
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <Typography variant="body2" fontWeight={500}>
                          {collab.username}
                        </Typography>
                        {isCurrentUser && <Typography variant="caption" sx={{ ml: 0.5 }}>(You)</Typography>}
                        {isOwner && <Typography variant="caption" sx={{ ml: 0.5, fontWeight: 500 }}>(Owner)</Typography>}
                      </Box>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        {canEditRole ? (
                          <Select
                            value={collab.role}
                            onChange={(e) => handleUpdateCollaboratorRole(collab.username, e.target.value)}
                            size="small"
                            disabled={editCollabSubmitting}
                            sx={{ minWidth: '90px', fontSize: '0.875rem' }}
                          >
                            <MenuItem value="read">Read</MenuItem>
                            <MenuItem value="write">Write</MenuItem>
                            {canEditRepo && <MenuItem value="admin">Admin</MenuItem>}
                          </Select>
                        ) : (
                          <Typography variant="caption" color="text.secondary">
                            {collab.role}
                          </Typography>
                        )}
                        {canRemove && (
                          <Tooltip title={buttonTooltip}>
                            <span>
                              <Button
                                size="small"
                                color="error"
                                onClick={() => handleRemoveCollaborator(collab.username)}
                                disabled={collabSubmitting}
                              >
                                {buttonLabel}
                              </Button>
                            </span>
                          </Tooltip>
                        )}
                        {!canRemove && (isOwner || !canEditRole) && (
                          <Tooltip title={buttonTooltip}>
                            <span>
                              <Button
                                size="small"
                                color="error"
                                disabled
                              >
                                Remove
                              </Button>
                            </span>
                          </Tooltip>
                        )}
                      </Box>
                    </Box>
                  );
                })}
              </Box>
            )}
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCollaboratorsClose}>Close</Button>
        </DialogActions>
      </Dialog>

      {/* Remove Collaborator Confirmation Dialog */}
      <Dialog open={removeCollabConfirmOpen} onClose={handleRemoveCollaboratorCancel}>
        <DialogTitle>
          {user && user.username === removeCollabUsername ? 'Leave Repository?' : 'Remove Collaborator?'}
        </DialogTitle>
        <DialogContent>
          <Typography>
            {user && user.username === removeCollabUsername
              ? 'Are you sure you want to leave this repository as a collaborator? You will lose access to it.'
              : `Are you sure you want to remove ${removeCollabUsername} as a collaborator? They will lose access to this repository.`}
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleRemoveCollaboratorCancel} disabled={collabSubmitting}>
            Cancel
          </Button>
          <Button
            onClick={handleRemoveCollaboratorConfirm}
            variant="contained"
            color="error"
            disabled={collabSubmitting}
          >
            {user && user.username === removeCollabUsername ? 'Leave' : 'Remove'}
          </Button>
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
          <Box>
            <Typography variant="subtitle2" fontWeight={600} sx={{ mb: 1 }}>
              Keep Access Rights (Optional)
            </Typography>
            <ToggleButtonGroup
              value={transferPreviousOwnerAccess}
              exclusive
              onChange={(e, newValue) => {
                if (newValue !== null) {
                  setTransferPreviousOwnerAccess(newValue);
                }
              }}
              fullWidth
            >
              <ToggleButton value="none" aria-label="no access">
                No Access
              </ToggleButton>
              <ToggleButton value="read" aria-label="read access">
                Read
              </ToggleButton>
              <ToggleButton value="write" aria-label="write access">
                Write
              </ToggleButton>
              <ToggleButton value="admin" aria-label="admin access">
                Admin
              </ToggleButton>
            </ToggleButtonGroup>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleTransferClose} disabled={transferSubmitting}>Cancel</Button>
          <Button variant="contained" color="error" onClick={handleTransferSubmit} disabled={transferSubmitting}>
            {transferSubmitting ? <CircularProgress size={20} /> : 'Transfer'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Transfer Confirmation Dialog */}
      <Dialog open={transferConfirmOpen} onClose={handleTransferConfirmClose} maxWidth="xs" fullWidth>
        <DialogTitle>Confirm Transfer</DialogTitle>
        <DialogContent sx={{ pt: '16px !important' }}>
          <Alert severity="error" sx={{ mb: 2 }}>
            Are you sure you want to transfer ownership of this repository to <strong>{transferUsername}</strong>? This action cannot be undone.
          </Alert>
          <Typography variant="body2" color="text.secondary">
            After confirmation, you will lose all ownership rights to this repository.
          </Typography>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleTransferConfirmClose} disabled={transferSubmitting}>Cancel</Button>
          <Button variant="contained" color="error" onClick={handleTransferConfirm} disabled={transferSubmitting}>
            {transferSubmitting ? <CircularProgress size={20} /> : 'Confirm Transfer'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* View Lineage Dialog */}
      <Dialog open={lineageOpen} onClose={handleLineageClose} maxWidth="sm" fullWidth>
        <DialogTitle>Repository Lineage</DialogTitle>
        <DialogContent sx={{ pt: '16px !important' }}>
          {lineageLoading && (
            <Box sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
              <CircularProgress />
            </Box>
          )}
          {lineageError && <Alert severity="error">{lineageError}</Alert>}
          {lineageData && !lineageLoading && (
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              {/* Parent Lineage */}
              <Box>
                <Typography variant="subtitle2" fontWeight={600} sx={{ mb: 1 }}>
                  Parent Lineage
                </Typography>
                {!lineageData.parent ? (
                  <Typography variant="body2" color="text.secondary">
                    No parent lineage visible.
                  </Typography>
                ) : (
                  <Box sx={{ pl: 2, borderLeft: '2px solid', borderColor: 'divider' }}>
                    {lineageData.ancestors && lineageData.ancestors.length > 0 ? (
                      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0 }}>
                        {[...lineageData.ancestors].reverse().map((ancestor, idx) => (
                          <Box key={idx}>
                            <MuiLink
                              component="button"
                              type="button"
                              variant="body2"
                              onClick={() => {
                                handleLineageClose();
                                navigate(`/repositories/${ancestor.owner}/${ancestor.slug}`, { state: { from: location.pathname } });
                              }}
                              sx={{ cursor: 'pointer', textDecoration: 'none', display: 'block' }}
                            >
                              {ancestor.owner}/{ancestor.slug}
                            </MuiLink>
                            {idx < lineageData.ancestors.length - 1 && (
                              <Typography variant="h6" color="text.secondary" sx={{ display: 'block', py: 0.25, lineHeight: 1 }}>
                                ↓
                              </Typography>
                            )}
                          </Box>
                        ))}
                      </Box>
                    ) : (
                      <MuiLink
                        component="button"
                        type="button"
                        variant="body2"
                        onClick={() => {
                          handleLineageClose();
                          navigate(`/repositories/${lineageData.parent.owner}/${lineageData.parent.slug}`, { state: { from: location.pathname } });
                        }}
                        sx={{ cursor: 'pointer', textDecoration: 'none' }}
                      >
                        {lineageData.parent.owner}/{lineageData.parent.slug}
                      </MuiLink>
                    )}
                  </Box>
                )}
              </Box>

              <Divider />

              {/* Child Forks */}
              <Box>
                <Typography variant="subtitle2" fontWeight={600} sx={{ mb: 1 }}>
                  Child Forks
                </Typography>
                {!lineageData.children || lineageData.children.length === 0 ? (
                  <Typography variant="body2" color="text.secondary">
                    No child repositories visible.
                  </Typography>
                ) : (
                  <Box sx={{ pl: 2, borderLeft: '2px solid', borderColor: 'divider' }}>
                    <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
                      {lineageData.children.map((child, idx) => (
                        <Box key={idx}>
                          <MuiLink
                            component="button"
                            type="button"
                            variant="body2"
                            onClick={() => {
                              handleLineageClose();
                              navigate(`/repositories/${child.owner}/${child.slug}`, { state: { from: location.pathname } });
                            }}
                            sx={{ cursor: 'pointer', textDecoration: 'none' }}
                          >
                            {child.owner}/{child.slug}
                          </MuiLink>
                        </Box>
                      ))}
                    </Box>
                  </Box>
                )}
              </Box>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleLineageClose}>Close</Button>
        </DialogActions>
      </Dialog>

      {/* Compare Tags Dialog */}
      <Dialog open={compareOpen} onClose={handleCompareClose} maxWidth="md" fullWidth>
        <DialogTitle>Compare Tags</DialogTitle>
        <DialogContent sx={{ display: 'flex', flexDirection: 'column', gap: 2, pt: '16px !important' }}>
          <Box sx={{ display: 'flex', gap: 2, alignItems: 'center' }}>
            <Select
              value={compareLeft}
              onChange={(e) => { setCompareLeft(e.target.value); setCompareResult(null); }}
              fullWidth
              size="small"
              displayEmpty
            >
              {tags.map((t) => (
                <MenuItem key={t.name} value={t.name}>{t.name}</MenuItem>
              ))}
            </Select>
            <Typography sx={{ flexShrink: 0 }}>vs</Typography>
            <Select
              value={compareRight}
              onChange={(e) => { setCompareRight(e.target.value); setCompareResult(null); }}
              fullWidth
              size="small"
              displayEmpty
            >
              {tags.map((t) => (
                <MenuItem key={t.name} value={t.name}>{t.name}</MenuItem>
              ))}
            </Select>
          </Box>
          {compareError && <Alert severity="error">{compareError}</Alert>}
          {compareLoading && <Box sx={{ display: 'flex', justifyContent: 'center', py: 2 }}><CircularProgress /></Box>}
          {compareResult && compareResult.identical && (
            <Alert severity="success">No differences — both tags are identical.</Alert>
          )}
          {compareResult && !compareResult.identical && (
            <Box
              sx={{
                fontFamily: 'monospace',
                fontSize: '0.75rem',
                overflowX: 'auto',
                maxHeight: '50vh',
                overflowY: 'auto',
                border: '1px solid',
                borderColor: 'divider',
                borderRadius: 1,
                p: 1,
              }}
            >
              {compareResult.lines.map((line, idx) => (
                <Box
                  key={idx}
                  sx={{
                    whiteSpace: 'pre',
                    backgroundColor:
                      line.type === 'add' ? '#e6ffed' :
                      line.type === 'remove' ? '#ffebe9' :
                      line.type === 'separator' ? '#f6f8fa' : 'transparent',
                    color:
                      line.type === 'add' ? '#1a7f37' :
                      line.type === 'remove' ? '#cf222e' :
                      line.type === 'separator' ? '#57606a' : 'inherit',
                    px: 1,
                    borderLeft: line.type === 'add' ? '3px solid #2da44e' :
                                line.type === 'remove' ? '3px solid #f85149' : '3px solid transparent',
                    fontStyle: line.type === 'separator' ? 'italic' : 'normal',
                  }}
                >
                  {line.type === 'add' ? '+ ' : line.type === 'remove' ? '- ' : line.type === 'separator' ? '  ' : '  '}{line.text}
                </Box>
              ))}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCompareClose}>Close</Button>
          <Button
            variant="contained"
            onClick={handleCompareRun}
            disabled={compareLoading || !compareLeft || !compareRight || compareLeft === compareRight}
          >
            {compareLoading ? <CircularProgress size={20} /> : 'Compare'}
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
