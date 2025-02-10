import { NavLink } from "react-router-dom";
import "./uploadPage.css";
const UploadPage = () => {
    return (
        <section class="profile-assets">
            <div class="left-nav-outer-container">
                <div class="account-tabs">
                    <NavLink className="profile-tab" to="" activeStyle>Profile</NavLink>
                    <NavLink className="assets-tab" to="" activeStyle>Assets</NavLink>
                    <NavLink className="settings-tab" to="" activeStyle>Settings</NavLink>
                </div>
                <div class="logout-container">
                    <NavLink className="logout-tab" to="" activeStyle>Logout</NavLink>
                </div>
            </div>
            <div class="asset-contents">
                <h1 class="asset-title"> My Assets</h1>
                <div class="upload-container">
                    <button> Upload </button>
                </div>
                <div class="asset-detail-container">
                    <table class="asset-detail-table"> 
                        <tr>
                            <th> Name </th>
                            <th> Size </th>
                            <th> Last Modified </th>
                            <th> Owner </th>
                            <th> Visibility </th>
                            <th> Actions </th>
                        </tr>
                        <tr>
                            <td> Model X </td>
                            <td> 1.5 MB </td>
                            <td> 2/8/2025 </td>
                            <td> User X </td>
                            <td> Private </td>
                            <td> Select </td>
                        </tr>
                        <tr>
                            <td> Model y </td>
                            <td> 1.0 MB </td>
                            <td> 2/8/2025 </td>
                            <td> User Y </td>
                            <td> Public </td>
                            <td> Select </td>
                        </tr>
                    </table>
                </div>
            </div>





            {/* <div>
                <h1>
                    Click the button below to upload a model to the OpenDI Model Hub
                </h1>
                <img src={opendiIcon} alt="OpenDI Icon" />
            </div> */}
        </section >
    );
};

export default UploadPage;
