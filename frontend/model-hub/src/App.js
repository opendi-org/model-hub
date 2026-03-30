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
import UploadPage from "./pages/uploadPage";
import ModelPage from './pages/modelPage';
import CliDownloadPage from './pages/downloadPage';
import AuthCallback from './pages/AuthCallback';
import LoginPage from './pages/login'
import Navbar from './components/Navbar';
import { getTheme } from './Theme';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import UserPage from "./pages/user";
import SearchPage from "./pages/search";
import NotFoundPage from "./pages/NotFound";
import { UserProvider } from './context/UserContext';

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
        <UserProvider>
          <Router>
            <Navbar />
            <Routes>
              <Route exact path="/" element={<Home />} />
              <Route path="/upload" element={<UploadPage />} />
              <Route path="/cli-download" element={<CliDownloadPage />} />
              <Route path="/model/:uuid" element={<ModelPage />} />
              <Route path="/model" element={<ModelPage />} />
              <Route path="/login" element={<LoginPage />} />
              <Route path="/auth/callback" element={<AuthCallback />} />
              <Route path="/user" element={<UserPage />} />
              <Route path="/search" element={<SearchPage />} />
              <Route path="*" element={<NotFoundPage />} />
            </Routes>
          </Router>
        </UserProvider>
      </ThemeProvider>
    </ColorModeContext.Provider>
  );
}

export default App;
