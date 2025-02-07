import { NavLink } from "react-router-dom";
const Navbar = () => {
    return (
        <div>
            <NavLink to="/UploadPage" activeStyle>
                Upload models
            </NavLink>
            <NavLink to="/DownloadPage" activeStyle>
                Download models
            </NavLink>
            <NavLink to="/" activeStyle>
                Home
            </NavLink>
        </div>
    );
};

export default Navbar;
