
import {
  BrowserRouter as Router,
  Routes,
  Route,
} from "react-router-dom";
import Home from "./pages";
import UploadPage from "./pages/uploadPage";
import DownloadPage from './pages/downloadPage';
import Navbar from './components/Navbar';

function App() {
  return (
    <Router>
      <Navbar />
      <Routes>
          <Route exact path="/" element={<Home />} />
          <Route path="/uploadpage" element={<UploadPage />} />
          <Route path="/downloadpage" element={<DownloadPage />} />
      </Routes>
    </Router>
  );
}

export default App;
