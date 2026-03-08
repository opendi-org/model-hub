//
// COPYRIGHT OpenDI
//

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
import {theme} from './Theme'
import {ThemeProvider} from '@mui/material/styles';
import UserPage from "./pages/user";
import SearchPage from "./pages/search";
import { UserProvider } from './context/UserContext';


function App() {
  return (
    <ThemeProvider theme={theme}>
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
          </Routes>
        </Router>
      </UserProvider>
    </ThemeProvider>
  );
}

export default App;
