import { createTheme } from '@mui/material/styles';

export const theme = createTheme({
    components: {
        MuiTypography: {
            styleOverrides: {
                fontFamily: 'Noto Sans, sans-serif',
            },
        },
        MuiPaper: {
            defaultProps: {
              color: '#6f8890'
            },
          },
    },
    palette: {
        primary: {
            main: '#63bad6',
        },
        secondary: {
            main: '#9dafb5',
        }
    },
    typography: {
        fontFamily: 'Noto Sans, sans-serif',
    }
});