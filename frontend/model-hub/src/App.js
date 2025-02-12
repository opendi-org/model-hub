
import {
  BrowserRouter as Router,
  Routes,
  Route,
} from "react-router-dom";
import Home from "./pages";
import UploadPage from "./pages/uploadPage";
import DownloadPage from './pages/downloadPage';
import Navbar from './components/Navbar';
import {theme} from './Theme'
import {ThemeProvider} from '@mui/material/styles';

function App() {
  return (
    <ThemeProvider theme={theme}>
    <Router>
    <Navbar />
    <Routes>
        <Route exact path="/" element={<Home />} />
        <Route path="/uploadpage" element={<UploadPage />} />
        <Route path="/downloadpage" element={<DownloadPage />} />
    </Routes>
</Router>
</ThemeProvider>
  );
}

export default App;
