//
// COPYRIGHT OpenDI
//

import { NavLink, useNavigate } from "react-router-dom";
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
import Tooltip from '@mui/material/Tooltip';
import DarkModeIcon from '@mui/icons-material/DarkMode';
import LightModeIcon from '@mui/icons-material/LightMode';
import { useUser } from '../context/UserContext';
import { useColorMode } from '../App';

const Search = styled('div')(({ theme }) => ({
    position: 'relative',
    borderRadius: 20,
    backgroundColor: alpha('#ffffff', 0.08),
    border: '1px solid rgba(255,255,255,0.2)',
    '&:hover': {
        backgroundColor: alpha('#ffffff', 0.12),
        borderColor: 'rgba(255,255,255,0.35)',
    },
    '&:focus-within': {
        backgroundColor: alpha('#ffffff', 0.14),
        borderColor: 'rgba(255,255,255,0.5)',
    },
    marginRight: theme.spacing(2),
    marginLeft: theme.spacing(2),
    width: '100%',
    [theme.breakpoints.up('sm')]: {
        width: 'auto',
    },
    transition: 'background-color 0.2s, border-color 0.2s',
}));

const SearchIconWrapper = styled('div')(({ theme }) => ({
    padding: theme.spacing(0, 2),
    height: '100%',
    position: 'absolute',
    pointerEvents: 'none',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    color: 'rgba(255,255,255,0.6)',
}));

const StyledInputBase = styled(InputBase)(() => ({
    color: '#ffffff',
    '& .MuiInputBase-input': {
        padding: '8px 12px 8px 0',
        paddingLeft: `calc(1em + 32px)`,
        width: '100%',
        '&::placeholder': {
            color: 'rgba(255,255,255,0.5)',
            opacity: 1,
        },
        '@media (min-width:900px)': {
            width: '36ch',
        },
    },
}));

const NavButton = styled(Button)(() => ({
    color: 'rgba(255,255,255,0.85)',
    fontWeight: 500,
    fontSize: '0.875rem',
    textTransform: 'none',
    padding: '6px 12px',
    borderRadius: 6,
    '&:hover': {
        color: '#ffffff',
        backgroundColor: 'rgba(255,255,255,0.1)',
    },
    '&.active': {
        color: '#ffffff',
        backgroundColor: 'rgba(255,255,255,0.12)',
    },
}));

export default function Navbar() {
    const { user, logout } = useUser();
    const { mode, toggleColorMode } = useColorMode();
    const [anchorEl, setAnchorEl] = React.useState(null);
    const [imageError, setImageError] = React.useState(false);
    const [searchValue, setSearchValue] = React.useState('');
    const navigate = useNavigate();

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

    const handleSearchSubmit = () => {
        if (searchValue.trim()) {
            navigate(`/search?term=${encodeURIComponent(searchValue.trim())}`);
        }
    };

    const getInitials = () => {
        if (!user) return 'U';
        if (user.username) return user.username[0].toUpperCase();
        if (user.email) return user.email[0].toUpperCase();
        return 'U';
    };

    return (
        <Box sx={{ flexGrow: 1 }}>
            <AppBar position="static">
                <Toolbar sx={{ gap: 0.5, minHeight: '60px !important' }}>
                    {/* Logo */}
                    <NavLink to="/" style={{ display: 'flex', alignItems: 'center', textDecoration: 'none', gap: 8, flexShrink: 0 }}>
                        <img
                            src={opendiIcon}
                            alt="OpenDI Logo"
                            style={{ height: 36 }}
                        />
                        <Typography
                            variant="h6"
                            sx={{
                                display: { xs: 'none', sm: 'block' },
                                color: '#ffffff',
                                fontWeight: 700,
                                fontSize: '1rem',
                                letterSpacing: '-0.01em',
                            }}
                        >
                            OpenDI
                        </Typography>
                    </NavLink>

                    {/* Search */}
                    <Search>
                        <SearchIconWrapper>
                            <SearchIcon fontSize="small" />
                        </SearchIconWrapper>
                        <StyledInputBase
                            placeholder="Search models..."
                            inputProps={{ 'aria-label': 'search' }}
                            value={searchValue}
                            onChange={(e) => setSearchValue(e.target.value)}
                            onKeyDown={(e) => {
                                if (e.key === 'Enter') handleSearchSubmit();
                            }}
                        />
                    </Search>

                    <Box sx={{ flexGrow: 1 }} />

                    {/* Nav links */}
                    <Box sx={{ display: { xs: 'none', md: 'flex' }, gap: 0.5, alignItems: 'center' }}>
                        <NavButton component={NavLink} to="/search">Explore</NavButton>
                        <NavButton component={NavLink} to="/cli-download">Download</NavButton>
                        <NavButton component={NavLink} to="/upload">Upload</NavButton>
                        <NavButton href="https://opendi.org" target="_blank" rel="noopener noreferrer">About</NavButton>
                    </Box>

                    {/* Dark/Light toggle */}
                    <Tooltip title={mode === 'light' ? 'Switch to dark mode' : 'Switch to light mode'}>
                        <IconButton
                            onClick={toggleColorMode}
                            size="small"
                            sx={{
                                color: 'rgba(255,255,255,0.75)',
                                ml: 1,
                                '&:hover': { color: '#ffffff', backgroundColor: 'rgba(255,255,255,0.1)' },
                            }}
                        >
                            {mode === 'light' ? <DarkModeIcon fontSize="small" /> : <LightModeIcon fontSize="small" />}
                        </IconButton>
                    </Tooltip>

                    {/* User section */}
                    <Box sx={{ ml: 1 }}>
                        {user ? (
                            <>
                                <Tooltip title={user.username || user.email}>
                                    <IconButton onClick={handleMenuOpen} sx={{ p: 0.5 }}>
                                        <Avatar
                                            alt={user.username}
                                            src={!imageError && user.picture ? user.picture : undefined}
                                            imgProps={{
                                                onError: () => setImageError(true),
                                                referrerPolicy: 'no-referrer',
                                            }}
                                            sx={{
                                                width: 32,
                                                height: 32,
                                                bgcolor: imageError || !user.picture ? '#086DD7' : undefined,
                                                border: '2px solid rgba(255,255,255,0.3)',
                                                fontSize: '0.875rem',
                                                fontWeight: 600,
                                            }}
                                        >
                                            {(imageError || !user.picture) && getInitials()}
                                        </Avatar>
                                    </IconButton>
                                </Tooltip>
                                <Menu
                                    anchorEl={anchorEl}
                                    open={Boolean(anchorEl)}
                                    onClose={handleMenuClose}
                                    anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
                                    transformOrigin={{ vertical: 'top', horizontal: 'right' }}
                                    PaperProps={{
                                        sx: {
                                            mt: 0.5,
                                            minWidth: 200,
                                            boxShadow: '0 8px 24px rgba(0,0,0,0.15)',
                                        },
                                    }}
                                >
                                    <MenuItem disabled sx={{ opacity: '1 !important' }}>
                                        <Box>
                                            <Typography variant="body2" fontWeight={600}>
                                                {user.username}
                                            </Typography>
                                            <Typography variant="caption" color="text.secondary">
                                                {user.email}
                                            </Typography>
                                        </Box>
                                    </MenuItem>
                                    <Divider />
                                    <MenuItem component={NavLink} to="/user" onClick={handleMenuClose}>
                                        My Profile
                                    </MenuItem>
                                    <MenuItem component={NavLink} to="/UploadPage" onClick={handleMenuClose}>
                                        Upload Model
                                    </MenuItem>
                                    <Divider />
                                    <MenuItem onClick={handleLogout} sx={{ color: 'error.main' }}>
                                        Sign out
                                    </MenuItem>
                                </Menu>
                            </>
                        ) : (
                            <Box sx={{ display: 'flex', gap: 1 }}>
                                <Button
                                    variant="outlined"
                                    size="small"
                                    component={NavLink}
                                    to="/login"
                                    sx={{
                                        color: '#ffffff',
                                        borderColor: 'rgba(255,255,255,0.4)',
                                        '&:hover': {
                                            borderColor: '#ffffff',
                                            backgroundColor: 'rgba(255,255,255,0.1)',
                                        },
                                        textTransform: 'none',
                                        fontWeight: 600,
                                        fontSize: '0.875rem',
                                    }}
                                >
                                    Sign in
                                </Button>
                                <Button
                                    variant="contained"
                                    size="small"
                                    component={NavLink}
                                    to="/login"
                                    sx={{
                                        backgroundColor: '#086DD7',
                                        '&:hover': { backgroundColor: '#0558AE' },
                                        textTransform: 'none',
                                        fontWeight: 600,
                                        fontSize: '0.875rem',
                                    }}
                                >
                                    Sign up
                                </Button>
                            </Box>
                        )}
                    </Box>
                </Toolbar>
            </AppBar>
        </Box>
    );
}
