import { createTheme } from '@mui/material/styles';

const dockerBlue = '#086DD7';
const navDarkLight = '#1E2130';
const navDarkDark = '#0D1117';

export const getTheme = (mode) =>
    createTheme({
        palette: {
            mode,
            primary: {
                main: dockerBlue,
                contrastText: '#FFFFFF',
            },
            secondary: {
                main: mode === 'light' ? '#667085' : '#9CA3AF',
                contrastText: '#FFFFFF',
            },
            background: {
                default: mode === 'light' ? '#F5F7FA' : '#111827',
                paper: mode === 'light' ? '#FFFFFF' : '#1F2937',
            },
            text: {
                primary: mode === 'light' ? '#1D2939' : '#F9FAFB',
                secondary: mode === 'light' ? '#667085' : '#9CA3AF',
            },
            divider: mode === 'light' ? '#D0D5DD' : '#374151',
            navbar: {
                background: mode === 'light' ? navDarkLight : navDarkDark,
                text: '#FFFFFF',
                border: mode === 'light' ? 'rgba(255,255,255,0.15)' : 'rgba(255,255,255,0.1)',
            },
        },
        typography: {
            fontFamily: '"Inter", "Segoe UI", "Roboto", "Helvetica Neue", sans-serif',
            h4: { fontWeight: 700 },
            h5: { fontWeight: 600 },
            h6: { fontWeight: 600 },
            subtitle1: { fontWeight: 400 },
            body2: { fontSize: '0.875rem' },
        },
        shape: {
            borderRadius: 8,
        },
        components: {
            MuiButton: {
                styleOverrides: {
                    root: {
                        borderRadius: 6,
                        textTransform: 'none',
                        fontWeight: 600,
                        fontSize: '0.875rem',
                    },
                    containedPrimary: {
                        backgroundColor: dockerBlue,
                        '&:hover': {
                            backgroundColor: '#0558AE',
                        },
                    },
                    outlinedPrimary: {
                        borderColor: dockerBlue,
                        color: dockerBlue,
                        '&:hover': {
                            borderColor: '#0558AE',
                            backgroundColor: 'rgba(8,109,215,0.06)',
                        },
                    },
                },
            },
            MuiCard: {
                styleOverrides: {
                    root: ({ theme }) => ({
                        border: `1px solid ${theme.palette.divider}`,
                        boxShadow: 'none',
                        transition: 'box-shadow 0.2s ease, transform 0.15s ease',
                        '&:hover': {
                            boxShadow: mode === 'light'
                                ? '0 4px 12px rgba(0,0,0,0.10)'
                                : '0 4px 12px rgba(0,0,0,0.40)',
                            transform: 'translateY(-1px)',
                        },
                    }),
                },
            },
            MuiCardHeader: {
                styleOverrides: {
                    root: ({ theme }) => ({
                        backgroundColor: theme.palette.background.paper,
                        borderBottom: `1px solid ${theme.palette.divider}`,
                    }),
                    title: {
                        fontWeight: 600,
                        fontSize: '1rem',
                    },
                    subheader: ({ theme }) => ({
                        color: theme.palette.text.secondary,
                        fontSize: '0.8rem',
                    }),
                },
            },
            MuiAppBar: {
                styleOverrides: {
                    root: ({ theme }) => ({
                        backgroundColor: theme.palette.navbar.background,
                        color: theme.palette.navbar.text,
                        boxShadow: 'none',
                        borderBottom: `1px solid ${
                            mode === 'light' ? 'rgba(255,255,255,0.08)' : 'rgba(255,255,255,0.05)'
                        }`,
                    }),
                },
            },
            MuiTextField: {
                styleOverrides: {
                    root: {
                        '& .MuiOutlinedInput-root': {
                            borderRadius: 8,
                        },
                    },
                },
            },
            MuiChip: {
                styleOverrides: {
                    root: {
                        borderRadius: 4,
                        fontSize: '0.75rem',
                        fontWeight: 500,
                    },
                },
            },
            MuiPaper: {
                styleOverrides: {
                    root: ({ theme }) => ({
                        backgroundImage: 'none',
                    }),
                },
            },
            MuiDivider: {
                styleOverrides: {
                    root: ({ theme }) => ({
                        borderColor: theme.palette.divider,
                    }),
                },
            },
        },
    });

export const theme = getTheme('light');
