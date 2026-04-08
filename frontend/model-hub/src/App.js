//
// COPYRIGHT OpenDI
//

import React, { useMemo, useState, createContext, useContext } from 'react';
import {
  BrowserRouter as Router,
  Routes,
  Route,
} from "react-router-dom";
import Home from "./pages";
import ModelPage from './pages/modelPage';
import CliDownloadPage from './pages/cliPage';
import AuthCallback from './pages/AuthCallback';
import LoginPage from './pages/login'
import SigninPage from './pages/signin';
import SignupPage from './pages/signup';
import Navbar from './components/Navbar';
import { getTheme } from './Theme';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import UserPage from "./pages/user";
import NotFoundPage from "./pages/NotFound";
import RepositoriesPage from "./pages/repositoriesPage";
import RepositoryDetailsPage from "./pages/repositoryDetailsPage";
import { UserProvider } from './context/UserContext';
import { RepositoryProvider } from './context/RepositoryContext';
import { NotificationProvider } from './context/NotificationContext';

export const ColorModeContext = createContext({ toggleColorMode: () => {}, mode: 'light' });

export const useColorMode = () => useContext(ColorModeContext);

function App() {
  const [mode, setMode] = useState(() => localStorage.getItem('colorMode') ?? 'light');

  const colorMode = useMemo(() => ({
    mode,
    toggleColorMode: () => {
      setMode((prev) => {
        const next = prev === 'light' ? 'dark' : 'light';
        localStorage.setItem('colorMode', next);
        return next;
      });
    },
  }), [mode]);

  const theme = useMemo(() => getTheme(mode), [mode]);

  return (
    <ColorModeContext.Provider value={colorMode}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <NotificationProvider>
          <UserProvider>
            <RepositoryProvider>
              <Router>
              <Navbar />
              <Routes>
                <Route exact path="/" element={<Home />} />
                <Route path="/repositories" element={<RepositoriesPage />} />
                <Route path="/repositories/:owner/:slug" element={<RepositoryDetailsPage />} />
                <Route path="/repositories/:repositoryId" element={<RepositoryDetailsPage />} />
                <Route path="/cli-tool" element={<CliDownloadPage />} />
                <Route path="/model/:uuid" element={<ModelPage />} />
                <Route path="/model" element={<ModelPage />} />
                <Route path="/login" element={<LoginPage />} />
                <Route path="/signin" element={<SigninPage />} />
                <Route path="/signup" element={<SignupPage />} />
                <Route path="/auth/callback" element={<AuthCallback />} />
                <Route path="/user" element={<UserPage />} />
              </Routes>
              </Router>
            </RepositoryProvider>
          </UserProvider>
        </NotificationProvider>
      </ThemeProvider>
    </ColorModeContext.Provider>
  );
}

export default App;
