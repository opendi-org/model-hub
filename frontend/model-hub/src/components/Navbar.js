import { NavLink } from "react-router-dom";
import "./Navbar.css";
import opendiIcon from '../opendi-icon.png';

const Navbar = () => {
    return (
        <section class="navbar">
            <div class="navbar_right">
                <NavLink className="navbar_brand" to="/">
                    <div class="navbar_logo">
                        <img src={opendiIcon} alt="OpenDI Logo" />
                    </div>
                    <div class="navbar_title"> OpenDI </div>
                </NavLink>

                <form class="search_bar">
                    <input type="text" value="Search OpenDI" />
                </form>
            </div>
            <div class="navbar_option_items">
                <nav class="nav_page_items">
                    <NavLink to="" activeStyle>Search</NavLink>
                    <NavLink to="/DownloadPage" activeStyle>Download</NavLink>
                    <NavLink to="/UploadPage" activeStyle>Upload</NavLink>
                    <NavLink to="" activeStyle>Popular</NavLink>
                    <NavLink to="" activeStyle>About</NavLink>
                </nav>

                <div class="nav_account_items">
                    <NavLink className="nav_login" to="" activeStyle>Login</NavLink>
                    <NavLink className="nav_signup" to="" activeStyle>Sign Up</NavLink>
                </div>
            </div>
        </section>
    );
};

export default Navbar;
