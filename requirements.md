# Model Hub Web Requirements

## Authentication

### [UC 01] Log In with Google

**Preconditions**

- The user is not logged in.
- The user has a Google account.

**Main Flow**

- The user chooses to log in using Google authentication.
- The user provides their Google credentials through the pop-up window that appears. [Invalid Account] [Create Username] [Existing Username]
- The user is successfully logged in.
- The user is directed to the Model Hub home page which contains information about OpenDI and a list of public repositories.

**Subflows**

- [Create Username] On the user’s first login with Google, they are directed to create a Model Hub account.  The user enters a desired username and the system verifies it is unique and valid. Once created, the Google account is linked to this Model Hub username for all future logins.
- [Existing Username] On subsequent logins with a Google account that already has a linked Model Hub username, the system logs in the user under their existing username without prompting for the username.

**Alternative Flows**

- [Invalid Account] If the Google authentication fails, the operation is cancelled and the error message from Google is displayed.

---

### [UC 02] View Profile

**Preconditions**

- The user is logged in to the Model Hub.

**Main Flow**

- The user selects the option for viewing their profile.
- The user’s username and email associated with the Google account they used to log in are displayed.

---

### [UC 03] Log Out

**Preconditions**

- The user is logged in to the Model Hub.

**Main Flow**

- The user chooses to log out.
- The user is no longer logged in.
- The user no longer has access to view any private repositories.

---

## Basic Repository Actions

### [UC 04] Create Repository

**Preconditions**

- The user is logged in.

**Main Flow**

- The user enters a required name for the new repository. [Invalid Name] [Duplicate Name]
- The user enters an optional description for the new repository.
- The user selects a visibility option for the new repository (private or public). [Public Repository]
- The repository is created and appears in the user’s list of owned repositories.

**Subflows**

- [Public Repository] The default visibility level for a new repository is private. If a user selects the public option, a confirmation window will appear confirming that the user wishes to create a public repository. If they click cancel, they are returned to the repository creation window.

**Alternative Flows**

- [Invalid Name] If the repository name is invalid, an appropriate error message is displayed.
- [Duplicate Name] If the repository name already exists under the user’s account, an appropriate error message is displayed.

---

### [UC 05] Search Repositories

**Main Flow**

- The user navigates to a page displaying a list of repositories. [Public Repositories] [User’s Repositories] [My Repositories]
- The user enters text in the search bar and optionally changes sorting and ordering options in the dropdown menus. [Sort By] [Order]
- The system displays a list of repositories where the repo name or description or owner username contain a match for the search text entered. [No Results]
- The system displays the number of repositories in the search results.

**Subflows**

- [Public Repositories] If the user is on the Model Hub home page, only public repositories will be displayed.
- [User’s Repositories] If the user is on another user’s page, only repositories owned by that user will be displayed.
- [My Repositories] If the user is logged in and is on the My Repositories page, they have two views: repositories that they own and repositories that have been shared with them.
- [Sort By] There are three options for what field to sort the search results by: name, updated date, and created date.
- [Order] There are two options for ordering of the sorted search results: descending and ascending.

**Alternative Flows**

- [No Results] If no repositories match the search criteria, an appropriate message is displayed.

---

### [UC 06] View Repository

**Preconditions**

- The user has read access to the target repository (public, owned, or shared with).

**Main Flow**

- The user selects a repository from the Model Hub home page, their owned/shared repos page, or another user’s profile.
- The user can view additional information:
  - The name and visibility level of the repository.
  - General description about the repository and what it represents.
  - Tags (named versions) associated with the repository.
  - A lineage tree with links to repositories which are parents/children of this repository, if applicable and the user has read access.
  - A list of collaborators and permissions, if the user owns the repository or has explicit shared access.
  - An option to fork the repository (create a “child” copy) if the user is logged in.
  - A permalink to the repository (ID-based rather than name-based) which can be copied.

---

### [UC 07] Update Repository

**Preconditions**

- The user is logged in.
- The user is the owner of the target repository.

**Main Flow**

- The user selects the option to edit the repository.
- The user can specify a new name for the repository. [Invalid Name] [Duplicate Name]
- The user can specify a new description for the repository.
- The user can change the visibility level of the repository (public or private). [Public Repository]
- The repository is updated and the user is directed to the updated repository page.

**Subflows**

- [Public Repository] If the user changes the visibility from private to public, a confirmation window will appear confirming that the user wishes to make the repository public. If they click cancel from this window, they will be returned to the edit repository window.

**Alternative Flows**

- [Invalid Name] If the repository name is invalid, an appropriate error message is displayed.
- [Duplicate Name] If the repository name already exists under the user’s account, an appropriate error message is displayed.

---

### [UC 08] Delete Repository

**Preconditions**

- The user is logged in.
- The user is the owner of the target repository.

**Main Flow**

- The user selects the delete action for the repository.
- A confirmation window appears and the user has to type the name of the repository to confirm deletion.
- The repository and its associated data are permanently deleted and it no longer appears on any pages. [Repository Not Found]

**Alternative Flows**

- [Repository Not Found] Any attempts to access a deleted repository via its URL will result in an error message stating that the repository was not found.

---

## Repository Sharing and Lineage

### [UC 09] Fork Repository

**Preconditions**

- The user is logged in.
- The user has at least read access to the target repository.

**Main Flow**

- The user selects the target repository and chooses to fork it.
- The user enters a name for the new forked (child) repository. [Invalid Name] [Duplicate Name]
- The user selects the tags for a one time copy over to the child repository.
- The user enters a description for the new repository and selects its visibility level (public or private). [Public Repository]
- The new repository is linked as a child of the original and the original is linked as a parent of the new repository. The new repository is independent and can be modified without affecting the original.

**Subflows**

- [Public Repository] The default visibility level for a new repository is private. If a user selects the public option, a confirmation window will appear confirming that the user wishes to create a public repository. If they click cancel, they will be returned to the repository creation page.

**Alternative Flows**

- [Invalid Name] If the repository name is invalid, an appropriate error message is displayed.
- [Duplicate Name] If the repository name already exists under the user’s account, an appropriate error message is displayed.

---

### [UC 10] Share Repository

**Preconditions**

- The user is logged in.
- The user owns the target repository or has admin access.

**Main Flow**

- The user selects the repository and selects the option for managing collaborators.
- The user provides a username to share the repository with. [Invalid User]
- The user selects a permission level for the user. [Read] [Write] [Admin]
- The user can choose to revoke or change existing share access for any user.
- The specified user’s access is updated with a success message.

**Subflows**

- [Read] Read access allows the user to view the repository along with its lineage tree and collaborators list, to download its tags, and to fork the repository.
- [Write] Write access includes read access and allows the user to edit the repository details and upload tags.
- [Admin] Admin access includes read and write access and allows the user to add, change and revoke collaborator permissions.

**Alternative Flows**

- [Invalid User] If the provided username doesn’t exist in the Model Hub, an appropriate error message is displayed.

---

### [UC 11] Transfer Repository Ownership

**Preconditions**

- The user is logged in.
- The user owns the target repository.

**Main Flow**

- The user selects the target repository to transfer ownership.
- The user provides the username of the new owner. [Invalid User] [Self Transfer]
- The user selects an option for the access rights to keep (no access, read, write, admin).
- The system prompts for confirmation of the ownership transfer. [Operation Cancelled]
- The repository ownership is transferred to the specified user.
- The original owner is either added as a collaborator with the access rights they specified, or not if they chose not to keep access rights.

**Subflows**

- [Operation Cancelled] If the user clicks cancel in the confirmation window, they will be returned to the transfer ownership options window.

**Alternative Flows**

- [Invalid User] If the provided username is invalid or the user does not exist, the operation is cancelled with an error message.
- [Self Transfer] If the user attempts to transfer ownership to themselves, the operation is cancelled with an error message.

---

## Model Actions

### [UC 12] Download Model

**Preconditions**

- The user has at least read access to the target repository (public, owned, or shared with).

**Main Flow**

- The user selects the repository and views available model tags.
- The user selects a tag to download the associated model.
- The model is exported as a JSON file to the user’s file system.

---

### [UC 13] Upload New Model

**Preconditions**

- The user is logged in.
- The user has at least write access to the target repository (owned or shared with).
- The user has a model file (JSON) to upload.

**Main Flow**

- The user selects the target repository for uploading the model.
- The user chooses to add a tag and enters the name of the tag. [Invalid Tag Name]
- The user chooses to upload a new CDM JSON file. [Invalid Model]
- The new model is added under the specified tag name and can be viewed in the repository’s list of tags.

**Alternative Flows**

- [Invalid Tag Name] If the tag name is invalid, an appropriate error message is displayed.
- [Invalid Model] If the model file fails OpenDI CDM schema validation, the upload operation is cancelled and an error message is displayed showing the line(s) of the file that caused the error.

---

### [UC 14] Add Tag From Existing Tag

**Preconditions**

- The user is logged in.
- The user has at least write access to the target repository (owned or shared with).

**Main Flow**

- The user selects the target repository for adding a tag.
- The user chooses to add a tag and enters the name of the tag. [Invalid Tag Name]
- The user chooses to use an existing tag and selects a tag from the dropdown menu.
- The new tag is added and points to the same model as the selected tag, which can be verified by the digests in the list of tags.

**Alternative Flows**

- [Invalid Tag Name] If the tag name is invalid, an appropriate error message is displayed.

---

### [UC 15] Edit Tag

**Preconditions**

- The user is logged in.
- The user has at least write access to the target repository (owned or shared with).

**Main Flow**

- The user selects the target repository for editing a tag.
- The user chooses to edit a tag from the list.
- The user can either upload a new model or use an existing tag. [Upload New] [Use Existing]
- The tag is updated and the user is redirected to the list of tags for the repository.

**Subflows**

- [Upload New] The user selects the option to upload a CDM JSON file from their filesystem. [Invalid Model]
- [Use Existing] The user selects an existing tag from the dropdown list and this tag now points to the same model as the selected tag.

**Alternative Flows**

- [Invalid Model] If the model file fails OpenDI CDM schema validation, the upload operation is cancelled and an error message is displayed showing the line(s) of the file that caused the error.

---

### [UC 16] Delete Tag

**Preconditions**

- The user is logged in.
- The user has at least write access to the target repository (owned or shared with).

**Main Flow**

- The user selects the option to delete a tag from a repository.
- A confirmation window appears informing the user that the deletion cannot be undone. [Operation Cancelled]
- The user proceeds with the operation and the tag is deleted; it no longer appears in the list of tags for the repository.

**Subflows**

- [Operation Cancelled] If the user chooses to cancel from the confirmation window, the deletion is cancelled and they are returned to the repository page.

---

### [UC 17] Compare Tags

**Preconditions**

- The user has at least read access to the repository (public, owned, or shared with)
- The target repository contains at least two tags.

**Main Flow**

- The user selects the option to compare two tags in the target repository.
- The user selects the two tags to generate a diff for. [Same Tag]
- The JSON diff is displayed for the selected tags. [Same Model]

**Subflows**

- [Same Model] If the user compares two tags that point to identical CDMs, a message is displayed stating that there is no difference between the tags.

**Alternative Flows**

- [Same Tag] If the user tries to compare a tag to itself, the compare option is not clickable in the UI.

---

# Model Hub CLI Requirements

## Authentication

### [UC 18] CLI Log in with Google

**Preconditions**

- The user is not logged in to the CLI.
- The user has a Google account.

**Main Flow**

- The user enters the login command in the CLI.
- A browser window opens with the Google login dialog.
- The user authenticates with their Google credentials. [Create Account] [Authentication Failed]
- A success message is displayed in the browser indicating that the user can now return to the CLI.
- The authentication token is securely saved on the user’s OS credential manager for further CLI use.
- A success message is displayed in the CLI including the authenticated username.

**Subflows**

- [Create Account] If the selected Google account is not yet linked to a Model Hub username, the user is redirected to the web flow for creating a username.

**Alternative Flows**

- [Authentication Failed] If the Google authentication fails or times out, the operation is cancelled and an error message is displayed.

---

### [UC 19] CLI Show Current User

**Preconditions**

- The user is logged in to the CLI.

**Main Flow**

- The user enters the whoami command in the CLI.
- The username of the currently logged-in user is displayed.

---

### [UC 20] CLI Log Out

**Preconditions**

- The user is logged in to the CLI.

**Main Flow**

- The user enters the logout command in the CLI.
- Stored user credentials are deleted from the OS credential manager.
- The user is no longer logged in and a confirmation message is displayed.

---

## Remote-based Actions

### [UC 21] CLI List Repositories

**Main Flow**

- The user enters the list repos command in the CLI.
- A list of repositories in the remote hub accessible to the user is returned.. [Authenticated] [Not Authenticated]

**Subflows**

- [Authenticated] If the user is logged in, the search results will show public repositories, repositories owned by the user, and repositories shared with the user.
- [Not Authenticated] If the user is not logged in, the search results will show only public repositories.

---

### [UC 22] CLI Search Repositories

**Main Flow**

- The user enters the search command with a search query in the CLI.
- A list of repositories in the remote hub accessible to the user with a name or description containing a match for the search query is returned. [Authenticated] [Not Authenticated] [No Results]

**Subflows**

- [Authenticated] If the user is logged in, the search results will show public repositories, repositories owned by the user, and repositories shared with the user.
- [Not Authenticated] If the user is not logged in, the search results will show only public repositories.

**Alternative Flows**

- [No Results] If no repositories match the search criteria, an appropriate message is displayed.

---

### [UC 23] CLI Inspect Repository

**Preconditions**

- The user has at least read access to the target repository (public, owned, or shared with).

**Main Flow**

- The user enters the inspect command in the CLI, specifying the repository in owner/slug format.
- Repository metadata is displayed [Not Found] [Access Denied]
- Description
- Collaborators (if the user has explicit shared access)
- Lineage Tree
- Tags

**Alternative Flows**

- [Not Found] If the specified repository doesn’t exist, an appropriate error message is displayed.
- [Access Denied] If the user doesn’t have access to the specified repository, an appropriate error message is displayed.

---

### [UC 24] CLI Inspect Tag

**Preconditions**

- The user has at least read access to the repository containing the target tag (public, owned, or shared with).

**Main Flow**

- The user enters the inspect command in the CLI, specifying the tag in owner/slug:tag format.
- Tag metadata is displayed [Not Found] [Access Denied]
- Repository Name
- Tag Name
- Tag Digest
- Tag Size
- Updated Date
- Creator Username

**Alternative Flows**

- [Not Found] If the specified repository/tag doesn’t exist, an appropriate error message is displayed.
- [Access Denied] If the user doesn’t have access to the specified repository, an appropriate error message is displayed.

---

### [UC 25] CLI Pull Model

**Preconditions**

- The user has at least read access to the repository containing the target model (public, owned, or shared with).

**Main Flow**

- The user enters the pull command in the CLI, specifying the model in owner/slug:tag format.
- CDM JSON and metadata associated with the tag is retrieved (if needed) and cached locally (but not saved to the user’s file system). [Access Denied] [Not Found] [Update Cache]

**Subflows**

- [Update Cache] The local cache is updated if information has changed since the last pull.

**Alternative Flows**

- [Not Found] If the specified repository/tag doesn’t exist in the Model Hub, an appropriate error message is displayed.
- [Access Denied] If the user doesn’t have access to the specified repository, an appropriate error message is displayed.

---

### [UC 26] CLI Push Model

**Preconditions**

- The user is logged in.
- The user has at least write access to the target repository (owned or shared with).
- The user has a CDM JSON file to push to the Model Hub.

**Main Flow**

- The user enters the push command in the CLI specifying the remote model reference in owner/slug:tag format along with the path to the local CDM JSON file to push. [Access Denied] [Invalid Tag] [Invalid Model] [Overwrite Warning]
- The cache is updated if operation was successful [Update Cache]
- The model is now stored in the remote hub.

**Subflows**

- [Update Cache] The local cache is updated since information has changed since the last pull.
- [Overwrite Warning] A warning is given that certain fields within the CDM JSON will be overwritten by Model Hub such as name and version.

**Alternative Flows**

- [Access Denied] If the user doesn’t have write access to the specified repository, an appropriate error message is displayed.
- [Invalid Tag] If the tag name is invalid, an appropriate error message is displayed.
- [Invalid Model] If the model file fails OpenDI CDM schema validation, the upload operation is cancelled and an error message is displayed showing the line(s) of the file that caused the error.

---

### [UC 27] CLI Create Repository

**Preconditions**

- The user is logged in.

**Main Flow**

- The user uses the create repo command and selects a name for the new repository. [Invalid Name] [Duplicate Name]
- The user specifies an optional description for the new repository.
- The user specifies the visibility of the repository (public or private).
- The new repository is created under the user’s account.

**Alternative Flows**

- [Invalid Name] If the repository name is invalid, an appropriate error message is displayed.
- [Duplicate Name] If the repository name already exists under the user’s account, an appropriate error message is displayed.

---

### [UC 28] CLI Delete Repository

**Preconditions**

- The user is logged in.
- The user is the owner of the target repository.

**Main Flow**

- The user uses the delete repo command, specifying the repository in owner/slug format. [Access Denied] [Not Found]
- The repository and its associated data are permanently deleted.

**Alternative Flows**

- [Access Denied] If the user is not the owner of the repository, an appropriate error message is displayed.
- [Not Found] If the repository doesn’t exist, an appropriate error message is displayed.

---

### [UC 29] CLI Add Tag

**Preconditions**

- The user is logged in.
- The user has at least write access to the repository (owned or shared with).

**Main Flow**

- The user enters the add tag command in the CLI specifying the name of the new tag to add, and the existing model for this new tag to point to in owner/slug:tag format. [Access Denied] [Invalid Tag]
- The cache is updated if the operation was successful. [Update Cache]
- The new tag now points to the specified model.

**Subflows**

- [Update Cache] The local cache is updated since information has changed since the last pull.

**Alternative Flows**

- [Access Denied] If the user doesn’t have write access to the specified repository, an appropriate error message is displayed.
- [Invalid Tag] If the tag name is invalid, an appropriate error message is displayed.

---

### [UC 30] CLI Delete Tag

**Preconditions**

- The user is logged in.
- The user has at least write access to the repository containing the target tag (owned or shared with).

**Main Flow**

- The user enters the delete tag command in the CLI specifying the tag in owner/slug:tag format. [Access Denied] [Not Found]
- The cache is updated if the operation was successful. [Update Cache]
- The tag no longer exists in the repository.

**Subflows**

- [Update Cache] The local cache is updated since information has changed since the last pull.

**Alternative Flows**

- [Access Denied] If the user doesn’t have write access to the specified repository, an appropriate error message is displayed.
- [Not Found] If the specified repository or tag doesn’t exist, an appropriate error message is displayed.

---

## Local Actions

### [UC 31] CLI Validate Model

**Main Flow**

- The user enters the validate command in the CLI specifying the local file.
- The validation results are returned to the user. A model is valid if it conforms to the OpenDI CDM schema specification. [Invalid Model]

**Alternative Flows**

- [Invalid Model] If the model file fails OpenDI CDM schema validation, an error message is displayed showing the line(s) of the file that caused the error.

---

## Remote-assisted Actions

### [UC 32] CLI Save Model

**Preconditions**

- The user has at least read access to the repository (public, owned, or shared with).

**Main Flow**

- The user enters the save command in the CLI specifying the model in owner/slug:tag format along with the desired local path to save the model to. [Access Denied] [Not Found]
- The cache is updated if operation was successful [Update Cache].
- A local working copy of the model is saved to the specified path.

**Subflows**

- [Update Cache] The local cache is updated since information has changed since the last pull.

**Alternative Flows**

- [Access Denied] If the user doesn’t have read access to the specified repository, an appropriate error message is displayed.
- [Not Found] If the specified repository or tag doesn’t exist, an appropriate error message is displayed.

---

### [UC 33] CLI Diff Models

**Preconditions**

- The user has at least read access to the specified remote repositories (owned or shared with)

**Main Flow**

- The user enters the diff command in the CLI specifying either 2 remote models in owner/slug:tag format, or a remote and a local model file. [Access Denied] [Not Found]
- The JSON diff between the specified files is displayed.

**Alternative Flows**

- [Access Denied] If the user doesn’t have read access to the specified repository, an appropriate error message is displayed.
- [Not Found] If the specified repository or tags don’t exist, an appropriate error message is displayed.

---

## Local Actions

### [UC 34] CLI List Local Models

**Main Flow**

- The user uses the list local command in the CLI. [No Local Models]
- Metadata of any locally cached models is displayed:
  - Username
  - Repository name
  - Tag name
  - Date/time pulled from Model Hub

**Alternative Flows**

- [No Local Models] If no models exist in the local cache, a message indicating this is displayed.

---

### [UC 35] CLI Remove Local Model

**Main Flow**

- The user uses the delete local command in the CLI, specifying the model in owner/slug:tag format. [Tag Not Found]
- The selected model is removed from the local cache.

**Alternative Flows**

- [Tag Not Found] The specified tag does not exist in the local cache.

---

### [UC 36] CLI Help

**Main Flow**

- The user enters the help command in the CLI. [Global Help] [Command Help]
- The CLI displays help information based on the selected context.

**Subflows**

- [Global Help] The user requests top-level help and the CLI displays a list of available commands with descriptions.
- [Command Help] The user requests help for a specific command and the CLI displays the syntax, arguments, and options for that command. [Invalid Command]

**Alternative Flows**

- [Invalid Command] If the command which help was requested for doesn’t exist, an appropriate error message is displayed.
