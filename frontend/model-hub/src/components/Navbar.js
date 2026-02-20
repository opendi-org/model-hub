//
// COPYRIGHT OpenDI
//

import { NavLink, useLocation, useNavigate } from "react-router-dom";
import opendiIcon from '../opendi-icon.png';
import * as React from 'react';
import { styled, alpha } from '@mui/material/styles';
import AppBar from '@mui/material/AppBar';
import Box from '@mui/material/Box';
import Toolbar from '@mui/material/Toolbar';
import Typography from '@mui/material/Typography';
import InputBase from '@mui/material/InputBase';
import MenuItem from '@mui/material/MenuItem';
import Menu from '@mui/material/Menu';
import SearchIcon from '@mui/icons-material/Search';
import Button from '@mui/material/Button';
import IconButton from '@mui/material/IconButton';
import Avatar from '@mui/material/Avatar';
import Divider from '@mui/material/Divider';
import Paper from '@mui/material/Paper';
import List from '@mui/material/List';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import CircularProgress from '@mui/material/CircularProgress';
import ClickAwayListener from '@mui/material/ClickAwayListener';
import { useUser } from '../context/UserContext';
import API_URL from '../config';

const Search = styled('div')(({ theme }) => ({
    position: 'relative',
    borderRadius: theme.shape.borderRadius,
    backgroundColor: alpha(theme.palette.common.white, 0.15),
    border: '1px solid #ccc',
    '&:hover': {
        backgroundColor: alpha(theme.palette.common.white, 0.25),
    },
    marginRight: theme.spacing(2),
    marginLeft: 0,
    width: '100%',
    [theme.breakpoints.up('sm')]: {
        marginLeft: theme.spacing(3),
        width: 'auto',
    },
}));

const SearchIconWrapper = styled('div')(({ theme }) => ({
    padding: theme.spacing(0, 2),
    height: '100%',
    position: 'absolute',
    pointerEvents: 'none',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
}));

const StyledInputBase = styled(InputBase)(({ theme }) => ({
    color: 'inherit',
    '& .MuiInputBase-input': {
        padding: theme.spacing(1, 1, 1, 0),
        paddingLeft: `calc(1em + ${theme.spacing(4)})`,
        transition: theme.transitions.create('width'),
        width: '100%',
        [theme.breakpoints.up('md')]: {
            width: '40ch',
        },
    },
}));

export default function Navbar() {
    const { user, logout } = useUser();
    const [anchorEl, setAnchorEl] = React.useState(null);
    const [imageError, setImageError] = React.useState(false);
    const location = useLocation();
    const navigate = useNavigate();

    const [navSearchTerm, setNavSearchTerm] = React.useState('');
    const [navResults, setNavResults] = React.useState([]);
    const [dropdownOpen, setDropdownOpen] = React.useState(false);
    const [loading, setLoading] = React.useState(false);

    React.useEffect(() => {
        if (!navSearchTerm.trim()) {
            setNavResults([]);
            setDropdownOpen(false);
            return;
        }
        const timer = setTimeout(() => {
            setLoading(true);
            fetch(`${API_URL}/v0/models/search/model/${navSearchTerm}`)
                .then(res => res.ok ? res.json() : [])
                .then(data => {
                    setNavResults(Array.isArray(data) ? data : []);
                    setDropdownOpen(true);
                })
                .catch(() => setNavResults([]))
                .finally(() => setLoading(false));
        }, 300);
        return () => clearTimeout(timer);
    }, [navSearchTerm]);

    const handleResultClick = (uuid) => {
        setNavSearchTerm('');
        setNavResults([]);
        setDropdownOpen(false);
        navigate(`/model/${uuid}`);
    };

    const handleSearchKeyDown = (e) => {
        if (e.key === 'Enter' && navSearchTerm.trim()) {
            setDropdownOpen(false);
            navigate(`/search?term=${navSearchTerm}`);
        }
    };

    const handleClickAway = () => {
        setDropdownOpen(false);
    };

    const activeStyle = (path) => ({
        fontWeight: location.pathname === path ? 'bold' : 'normal',
        borderBottom: location.pathname === path ? '2px solid black' : 'none',
        borderRadius: 0,
    });

    const handleMenuOpen = (event) => {
        setAnchorEl(event.currentTarget);
    };

    const handleMenuClose = () => {
        setAnchorEl(null);
    };

    const handleLogout = () => {
        logout();
        handleMenuClose();
    };

    // Helper function to get initials for avatar
    const getInitials = () => {
        if (!user) return 'U';
        if (user.username) return user.username[0].toUpperCase();
        if (user.email) return user.email[0].toUpperCase();
        return 'U';
    };

    return (
        <Box sx={{ flexGrow: 1 }}>
            <AppBar position="static" sx={{ backgroundColor: 'white', color: 'black' }}>
                <Toolbar>
                    <NavLink to="/" style={{ textDecoration: 'none' }}>
                        <img
                            src={opendiIcon}
                            alt="OpenDI Logo"
                            style={{ height: 40, marginRight: 10 }}
                        />
                    </NavLink>
                    <Typography
                        variant="h6"
                        noWrap
                        component={NavLink}
                        to="/"
                        sx={{ display: { xs: 'none', sm: 'block' }, textDecoration: 'none', color: 'inherit' }}
                    >
                        OpenDI
                    </Typography>
                    <ClickAwayListener onClickAway={handleClickAway}>
                        <Box sx={{ position: 'relative' }}>
                            <Search>
                                <SearchIconWrapper>
                                    {loading ? (
                                        <CircularProgress size={18} color="inherit" />
                                    ) : (
                                        <SearchIcon />
                                    )}
                                </SearchIconWrapper>
                                <StyledInputBase
                                    placeholder="Search…"
                                    inputProps={{ 'aria-label': 'search' }}
                                    sx={{ width: '25em' }}
                                    value={navSearchTerm}
                                    onChange={(e) => setNavSearchTerm(e.target.value)}
                                    onKeyDown={handleSearchKeyDown}
                                    onFocus={() => navResults.length > 0 && setDropdownOpen(true)}
                                />
                            </Search>
                            {dropdownOpen && (
                                <Paper
                                    sx={{
                                        position: 'absolute',
                                        top: '100%',
                                        left: 0,
                                        right: 0,
                                        zIndex: 1300,
                                        maxHeight: 360,
                                        overflowY: 'auto',
                                        mt: 0.5,
                                        boxShadow: 3,
                                    }}
                                >
                                    <List dense disablePadding>
                                        {navResults.length === 0 ? (
                                            <ListItemButton disabled>
                                                <ListItemText primary="No results found" />
                                            </ListItemButton>
                                        ) : (
                                            navResults.map((result) => (
                                                <ListItemButton
                                                    key={result.meta.UUID}
                                                    onClick={() => handleResultClick(result.meta.UUID)}
                                                    divider
                                                >
                                                    <ListItemText
                                                        primary={result.meta.name}
                                                        secondary={result.meta.creator?.username}
                                                    />
                                                </ListItemButton>
                                            ))
                                        )}
                                    </List>
                                </Paper>
                            )}
                        </Box>
                    </ClickAwayListener>
                    <Box sx={{ flexGrow: 1 }} />

                    <Box sx={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                        <Button color="inherit" component={NavLink} to="/search" sx={activeStyle('/search')}>Search</Button>
                        <Button color="inherit" component={NavLink} to="/cli-download" sx={activeStyle('/cli-download')}>Download</Button>
                        <Button color="inherit" component={NavLink} to="/upload" sx={activeStyle('/upload')}>Upload</Button>
                        <Button color="inherit" href="https://opendi.org" target="_blank">About</Button>
                        
                        {user ? (
                            <>
                                <IconButton onClick={handleMenuOpen} sx={{ p: 0, ml: 1 }}>
                                    <Avatar 
                                        alt={user.username}
                                        src={!imageError && user.picture ? user.picture : undefined}
                                        imgProps={{
                                            onError: () => {
                                                console.log('Avatar image failed to load, using fallback');
                                                setImageError(true);
                                            },
                                            referrerPolicy: 'no-referrer',
                                        }}
                                        sx={{ 
                                            width: 32, 
                                            height: 32,
                                            bgcolor: imageError || !user.picture ? '#1976d2' : undefined,
                                        }}
                                    >
                                        {(imageError || !user.picture) && getInitials()}
                                    </Avatar>
                                </IconButton>
                                <Menu
                                    anchorEl={anchorEl}
                                    open={Boolean(anchorEl)}
                                    onClose={handleMenuClose}
                                    anchorOrigin={{
                                        vertical: 'bottom',
                                        horizontal: 'right',
                                    }}
                                    transformOrigin={{
                                        vertical: 'top',
                                        horizontal: 'right',
                                    }}
                                >
                                    <MenuItem disabled>
                                        <Typography variant="body2" color="text.secondary">
                                            {user.email}
                                        </Typography>
                                    </MenuItem>
                                    <MenuItem component={NavLink} to="/user" onClick={handleMenuClose}>
                                        Profile
                                    </MenuItem>
                                    <Divider />
                                    <MenuItem onClick={handleLogout}>Logout</MenuItem>
                                </Menu>
                            </>
                        ) : (
                            <Button color="inherit" component={NavLink} to="/login" sx={activeStyle('/login')}>
                                Login
                            </Button>
                        )}
                    </Box>
                </Toolbar>
            </AppBar>
        </Box>
    );
}