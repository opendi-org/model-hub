import opendiIcon from '../opendi-icon.png';
const cdm = {
    creator : 'No CDM loaded'
};

async function getCDM() {
    const url = "http://localhost:8080/models";
    console.log(cdm.creator);
    try {
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(`Response status: ${response.status}`);
      }
  
      const json = await response.json();
      cdm.creator = json.meta.creator;
      document.getElementById("cdm").innerHTML = cdm.creator;
    } catch (error) {
      console.error(error.message);
    }
  }
  

const DownloadPage = () => {
    return (
        <div>
            <h1>
                Click the button below to download a model from the OpenDI Model Hub
            </h1>
            <img src={opendiIcon} alt="OpenDI Icon" />
            <button onClick={getCDM}>Get CDM</button>
            <p id="cdm">{cdm.creator}</p>
        </div>
    );
};

export default DownloadPage;
