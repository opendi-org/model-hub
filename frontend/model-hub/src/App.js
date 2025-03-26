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
import ModelPage from './pages/downloadPage';
import LoginPage from './pages/login'
import Navbar from './components/Navbar';
import {theme} from './Theme'
import {ThemeProvider} from '@mui/material/styles';
import UserPage from "./pages/user";
import SearchPage from "./pages/search";

function App() {
  return (
    <ThemeProvider theme={theme}>
    <Router>
    <Navbar />
    <Routes>
        <Route exact path="/" element={<Home />} />
        <Route path="/uploadpage" element={<UploadPage />} />
        <Route path="/model/:uuid" element={<ModelPage />} />
        <Route path="/model" element={<ModelPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/user" element={<UserPage />} />
        <Route path="/search" element={<SearchPage />} />
    </Routes>
</Router>
</ThemeProvider>
  );
}

export default App;
