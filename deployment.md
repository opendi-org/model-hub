# Deployment Instructions for OpenDI Model Hub
If you already have the up-to-date containers pushed to the container registry, skip to Step X. Make sure you have an Oracle Cloud account setup

### Step 1: Setup Container Repository
Click on the menu in the upper left corner of Oracle Cloud Infrastructure, and go to Developer Services -> Container Regsitry. Click "Create repository", 
and make one repository for the API container, and one for the frontend container. Make sure the repositories are public.

You will also need an auth token to access these repositories. Go to User -> User Settings -> Auth Tokens, and generate a token. Make sure to save this somewhere!

### Step 2: Push Containers to Registry
Make sure your containers are all up to date by running "docker compose --build". Then, login to OCI Registry in docker with
"docker login (region-key).ocir.io", where your region key is likely iad for US-East.

When prompted, enter your username as (tenancy-namespace)/(username), where tenancy namespace is found under Profile -> Tenancy, listed as "Object Storage Namespace".
It should look something like idkpm9sketnr. For your password, use the auth token we created in step 1.

You need to tag the images you created in step 1 with "docker tag (image-name) (region-key).ocir.io/(tenancy-namespace)/(repo-name):(tag)". Once you have tagged both 
the api and the frontend, run "docker push (region-key).ocir.io/(tenancy-namespace)/(repo-name):(tag)". Check the Container Registry on Oracle Cloud 
to make sure they are showing up. I found that on my first push I would get a conflict error, so simply pushing again with the same command fixed that issue.

### Step 3: Create Compute Instance
From the menu, go to Compute -> Instances and click Create Instance. Select your Image to be the Ubuntu version you are using, and the shape to be anything with 
at least 4 Gb of memory (I personally have been using one with 8 Gb). Make sure to generate an SSH key so you can access the VM, and same it somewhere. Once you have done so,
SSH into the VM.

### Step 4: Expose Necessary Endpoints
From your Instances page, click on the instance you just created, then click on the "Virtual Cloud Network" link. Next, click on the subnet that is associated with your 
VM (should be the only subnet there). Finally, click on the Default Security List. Once there, click on "Add Ingress Rule".

For Source, type 0.0.0.0/0 for all sources, and for destination port type 3000. Do the same thing again, except for destination port type 8080. These expose the endpoints 
for the OpenDI model hub to the open web for the frontend and API respectively.

### Step 5: Install Docker
This tutorial from the official Docker page works well: https://docs.docker.com/engine/install/ubuntu/

### Step 6: Pull Containers and Run Them
Go to your container registry, find the most up to date container for each service, and click "Copy pull command". Mine looks like this:
"docker pull iad.ocir.io/idkpm9sketnr/opendi-api:1.1" (Feel free to use this api, as my repository is public so you should be able to run this command).
Ensure that you have pulled images for both the API and the frontend. 

Next, copy over compose.prod.yaml, I found that the simplest way to do this was to copy the entire file to my clipboard, and then use vim to paste it into the
file system on the VM. So run "touch prod.compose.yaml", then run "vim prod.compose.yaml". Once in vim, click the i key, paste the file in, and then type ":wq" to 
save and exit. Now, you should be able to run "docker compose up", and access 
