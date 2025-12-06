# **OpenDI Model Hub \- Deployment Guide**

The following instructions describe how to deploy the OpenDI Model Hub using Oracle Cloud Infrastructure (OCI). These steps assume that you already have an OCI account, access to the Model Hub container images, and the necessary credentials for database connections and Google OAuth authentication, which is now required for user login.

**1\. Set Up OCI Container Repositories**  
Begin by navigating to *Developer Services → Container Registry* in the OCI dashboard. Create two private repositories—one dedicated to the backend API container and another for the frontend web application. Make sure the repositories are accessible within your tenancy. Next, generate an Auth Token by visiting *User Settings → Auth Tokens*. This token will serve as your Docker password when pushing container images. Keep this token stored safely, as you will need it each time you authenticate for image uploads.

**2\. Build, Tag, and Push Containers to OCI**  
Ensure your images are up to date by running docker compose build or the appropriate build command for your environment. Log into the OCI registry using docker login with your tenancy namespace and Auth Token. Tag both the backend and frontend images according to the required OCI format, which includes your region key, tenancy namespace, and repository name. Push each image to its respective repository with docker push. After pushing, verify that the images appear inside OCI Container Registry. If a conflict or outdated tag error occurs, simply re-tag and push again.

**3\. Create and Configure a Compute Instance**  
Navigate to *Compute → Instances* and create a new virtual machine. Choose an Ubuntu image and allocate at least 4 GB of RAM (8 GB provides smoother performance). During creation, upload or generate an SSH key for remote access. Once the instance is deployed, obtain its public IP address and connect via SSH to perform the remainder of the setup. This VM will host and run both the backend and frontend containers.

**4\. Open Required Network Ports and Security Rules**  
To make the Model Hub accessible externally, open the VM’s subnet configuration under its attached Virtual Cloud Network. Modify the security list to allow inbound traffic on port **8080** (backend API) and port **3000** (frontend UI). Set the allowed source to 0.0.0.0/0 unless tighter access control is desired. These firewall rules enable users and developers to interact with the deployed frontend interface and API endpoints.

**5\. Install Docker and Prepare the Environment**  
Install Docker on the VM using the official Ubuntu installation instructions or Docker’s convenience script. This tutorial from the official Docker page works well: [https://docs.docker.com/engine/install/ubuntu/](https://docs.docker.com/engine/install/ubuntu/). After installation, verify that Docker is working by running basic commands. At this stage, you should also prepare any environment variables required by the backend, including database credentials, JWT signing secrets, and Google OAuth client information. The backend relies on these values to correctly authenticate users and communicate with the Model Hub database.

**6\. Pull and Run the Model Hub Containers**  
Locate your images inside OCI Registry and use the “Copy Pull Command” option to pull each one onto your VM. Once both containers are present, you may either run them directly via individual Docker commands or use a simple docker-compose file to manage them together. Ensure that ports 8080 and 3000 are properly mapped and that the backend container can reach your MySQL instance. After verifying your configuration, run docker compose up \-d (or equivalent) to start the Model Hub services in detached mode.

**7\. Verify Deployment and Test Core Functionality**  
After deployment, access the frontend through your VM’s public IP on port 3000\. You should be able to log in using Google OAuth, which will redirect you through the authenticated flow and return you to the Model Hub interface. From there, confirm that models can be viewed and that privacy settings, sharing controls, and ownership transfer behave as expected. Additionally, the CLI should be able to authenticate using its token and interact with the backend via push, pull, commit, and lineage commands.
