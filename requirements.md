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

**Subflows**

- [Create Username] On the user's first login with Google, they are prompted to create a Model Hub username. The user enters a desired username and the system verifies it is unique and valid. Once created, the Google account is linked to this Model Hub username for all future logins.
- [Existing Username] On subsequent logins with a Google account that already has a linked Model Hub username, the system automatically logs in the user under their existing username without prompting.

**Alternative Flows**

- [Invalid Account] If the Google authentication fails, the operation is cancelled and the user is notified.

---

### [UC 02] Log Out

**Preconditions**

- The user is logged in to the Model Hub.

**Main Flow**

- The user chooses to log out.
- The user is no longer logged in.

---

## Basic Repository Actions

### [UC 03] Create Repository

**Preconditions**

- The user is logged in

**Main Flow**

- The user selects a name for the new repository. [Invalid Name] [Duplicate Name]
- The user specifies the name and description of the repository.

**Alternative Flows**

- [Invalid Name] If the repository name is invalid, an error message is displayed.
- [Duplicate Name] If the repository name already exists under the user's account, an error message is displayed.

---

### [UC 04] Search Repositories

**Main Flow**

- The user enters text to search for.
- The system displays a list of repositories that match the search criteria.
- Search results are displayed according to the user's authentication status. [Authenticated] [Not Authenticated]
- The user may select a repository to view its details.

**Subflows**

- [Authenticated] If the user is logged in, the search results will show public repositories, repositories owned by the user, and repositories shared with the user.
- [Not Authenticated] If the user is not logged in, the search results will show only public repositories.

**Alternative Flows**

- [No Results] If no repositories match the search criteria, an appropriate message is displayed to the user.

---

### [UC 05] View Repository

**Preconditions**

- The user has read access to the repository (public, owned, or shared with)

**Main Flow**

- The user selects a repository from the Model Hub home page, search results, or a user's profile.
- The user can view additional information:
  - General documentation about the repository and what it represents.
  - Tags (named versions) associated with the repository.
  - A list of links to repositories which are parents/children of this repository, if applicable and the user has read access.
  - A list of collaborators and permissions, if the user owns or has explicit shared access.

---

### [UC 06] Update Repository

**Preconditions**

- The user is logged in
- The user has edit access to the target repo (owner, shared with)

**Main Flow**

- The user selects a name for the repository. [Invalid Name] [Duplicate Name]
- The user specifies the updated name and description of the repository, if applicable.

**Alternative Flows**

- [Invalid Name] If the repository name is invalid, an error message is displayed.
- [Duplicate Name] If the repository name already exists under the user's account, an error message is displayed.

---

### [UC 07] Delete Repository

**Preconditions**

- The user is logged in.
- The user is the owner of the repository.

**Main Flow**

- The user selects the repository.
- The user selects the delete action to permanently delete the repository and associated data. [Access Denied]

**Alternative Flows**

- [Access Denied] Either the repository does not exist or the user is not the owner of the repository.

---

## Repository Sharing and Lineage

### [UC 08] Fork Repository

**Preconditions**

- The user is logged in
- The user has at least read access to the target repository.

**Main Flow**

- The user selects the target repository and chooses to fork it.
- The user provides a name for the new forked (child) repository. [Invalid Name] [Duplicate Name]
- The user provides a description for the new repository.
- The user selects the tags for a one time copy over to the child repository.
- The new repository is linked as a child of the original and the original is linked as a parent of the new repository. The new repository is independent and can be modified without affecting the original.

**Alternative Flows**

- [Invalid Name] If the repository name is invalid, an error message is displayed.
- [Duplicate Name] If the repository name already exists under the user's account, an error message is displayed.

---

### [UC 09] Set Repository Privacy

**Preconditions**

- The user is logged in
- The user owns the target repository

**Main Flow**

- The user selects the target repository to modify privacy settings.
- The user selects the privacy level. [Public] [Private]
- The user is notified of a successful privacy update.

**Subflows**

- [Public] The user can choose to make the repository publicly accessible to all Model Hub users.
- [Private] The user can choose to make the repository accessible only to themselves or can choose to share with specific users.

---

### [UC 10] Share Repository

**Preconditions**

- The user is logged in
- The user owns the target repository

**Main Flow**

- The user selects the target repository to modify sharing settings.
- The user provides a list of usernames to share the model with. [Invalid User] [Invalid Update]
- The user selects permission level for shared users (read only or read and write).
- The user can choose to revoke existing share access for any user. [Invalid User] [Invalid Update]
- The user is notified of the successful sharing update.
- Shared users are notified of the new access.

**Alternative Flows**

- [Invalid User] If one or more of the provided usernames are invalid, the operation is cancelled and the user is notified.
- [Invalid Update] If the action is the same as existing access for one or more specified users, the user is notified and can choose to update permissions or skip those users.

---

### [UC 11] Transfer Repository Ownership

**Preconditions**

- The user is logged in
- The user owns the target repository

**Main Flow**

- The user selects the target repository to transfer ownership.
- The user provides the username of the new owner. [Invalid User] [Self Transfer]
- The system prompts for confirmation of the ownership transfer. [Transfer Cancelled]
- The repository ownership is transferred to the specified user.
- Both the original owner and new owner are notified of the successful transfer. [Notification]

**Subflows**

- [Notification] Both users receive notifications about the ownership change via email.

**Alternative Flows**

- [Invalid User] If the provided username is invalid or the user does not exist, the operation is cancelled and the user is notified.
- [Self Transfer] If the user attempts to transfer ownership to themselves, the operation is cancelled and the user is notified.
- [Transfer Cancelled] If the user cancels the confirmation prompt, the operation is cancelled and no changes are made.

---

## Model Actions

### [UC 12] Download Model

**Preconditions**

- The user has read access to the repository (public, owned, or shared with)

**Main Flow**

- The user selects the repository and views available model tags.
- The user selects a tag to download the associated model.
- The model is exported as a JSON file to the user's file system.

---

### [UC 13] Upload Model

**Preconditions**

- The user is logged in
- The user has write access to target repository (owned, or shared with)
- The user has an updated model file (JSON) to upload

**Main Flow**

- The user selects the target repository for uploading the model.
- The user specifies an existing or new tag name for this update. [Invalid Tag]
- The user uploads an updated CDM JSON file and the system validates the model structure against the CDM schema. [Invalid Model]
- The new model is uploaded and mapped to the given tag.

**Alternative Flows**

- [Invalid Tag] If the tag name is invalid (special characters), an error message is displayed telling the user which characters were invalid.
- [Invalid Model] If the model file fails OpenDI CDM schema validation, the upload operation is cancelled and an error message is displayed showing the line(s) of the file that caused the error.

---

### [UC 14] View Model Diff

**Preconditions**

- The user has read access to the repository (public, owned, or shared with)
- Target repository contains at least two tags

**Main Flow**

- The user selects the target repository viewing diff.
- The user selects the two tags to generate a diff for.
- The JSON diff is displayed for selected tags.

---

# Model Hub CLI Requirements

## Authentication

### [UC 15] CLI Log in with Google

**Preconditions**

- The user is not logged in
- The user has a Google account
- The user has previously logged in with Model Hub Web

**Main Flow**

- The user enters the login command in the CLI.
- A browser window opens with the Google login dialog.
- The user authenticates with their Google credentials [Authentication Failed]
- The authentication token is securely saved on the user's system for further CLI use.
- A message is displayed in the browser window and CLI indicating successful login.

**Alternative Flows**

- [Authentication Failed] If the Google authentication fails or times out, the operation is cancelled and an error message is displayed.

---

### [UC 16] CLI Log Out

**Preconditions**

- The user is logged in to Model Hub CLI

**Main Flow**

- The user enters the logout command in the CLI.
- Stored user credentials are deleted from the system.
- The user is no longer logged in.

---

## Remote-based Actions

### [UC 17] CLI Search Repositories

**Main Flow**

- The user enters the search command in the CLI.
- Search results are displayed according to the user's authentication status. [Authenticated] [Not Authenticated]
- The user may select a repository to view its details.

**Subflows**

- [Authenticated] If the user is logged in, the search results will show public repositories, repositories owned by the user, and repositories shared with the user.
- [Not Authenticated] If the user is not logged in, the search results will show only public repositories.

**Alternative Flows**

- [No Results] If no repositories match the search criteria, an appropriate message is displayed to the user.

---

### [UC 18] CLI Inspect Repository

**Preconditions**

- The user has read access to the repository (public, owned, or shared with)

**Main Flow**

- The user enters the inspect repository command in the CLI specifying owner and name.
- Repository metadata details are displayed [Access Denied]:
  - Description
  - Collaborators (if provided explicit shared access)
  - Lineage information
  - Tags

**Alternative Flows**

- [Access Denied] The repository is private and the user is either not logged in or does not have access.

---

### [UC 19] CLI Inspect Model

**Preconditions**

- The user has read access to the repository (public, owned, or shared with)

**Main Flow**

- The user enters the inspect model command in the CLI specifying owner, name, and tag.
- Model (tag) metadata details are displayed [Access Denied]:
  - Owner, repo name
  - Creator or updater for tag
  - Digest and timestamps

**Alternative Flows**

- [Access Denied] The repository is private and the user is either not logged in or does not have access.

---

### [UC 20] CLI Pull Model

**Preconditions**

- The user has read access to the repository (public, owned, or shared with)

**Main Flow**

- The user enters the pull command in the CLI specifying owner, name, and tag.
- CDM JSON and metadata associated with the tag is retrieved (if needed) and cached locally. [Access Denied] [Update Cache]

**Subflows**

- [Update Cache] The local cache is updated if information has changed since the last pull.

**Alternative Flows**

- [Access Denied] The repository is private and the user is either not logged in or does not have access.

---

## Remote-assisted Actions

### [UC 21] CLI Create Repository

**Preconditions**

- The user is logged in

**Main Flow**

- The user uses create repository command and selects a name for the new repository. [Invalid Name] [Duplicate Name]
- The user specifies the visibility of the repository.
- The new repository is created under the user's account.

**Alternative Flows**

- [Invalid Name] If the repository name is invalid, an error message is displayed.
- [Duplicate Name] If the repository name already exists under the user's account, an error message is displayed.

---

### [UC 22] CLI Delete Repository

**Preconditions**

- The user is logged in.
- The user is the owner of the repository.

**Main Flow**

- The user uses the delete repository command to permanently delete the repository and associated data. [Access Denied]

**Alternative Flows**

- [Access Denied] Either the repository does not exist or the user is not the owner of the repository.

---

### [UC 23] CLI Push Model

**Preconditions**

- The user is logged in.
- The user has write access to the repository (owned, or shared with)

**Main Flow**

- The user enters the push command in the CLI specifying owner, name, and tag along with the local file. [Access Denied] [Invalid Tag] [Invalid Model] [Overwrite Warning]
- The cache is updated if operation was successful [Update Cache]

**Subflows**

- [Update Cache] The local cache is updated since information has changed since the last pull.
- [Overwrite Warning] A warning is given that certain fields within the CDM JSON will be overwritten by Model Hub such as name and version.

**Alternative Flows**

- [Access Denied] The repository is private and the user is either not logged in or does not have access.
- [Invalid Tag] If the tag name is invalid (special characters), an error message is displayed telling the user which characters were invalid.
- [Invalid Model] If the model file fails OpenDI CDM schema validation, the upload operation is cancelled and an error message is displayed showing the line(s) of the file that caused the error.

---

### [UC 24] CLI Delete Tag

**Preconditions**

- The user is logged in.
- The user has write access to the repository (owned, or shared with)

**Main Flow**

- The user enters the delete tag command in the CLI specifying owner, name, and tag. [Access Denied] [Tag Not Found]
- The cache is updated if operation was successful [Update Cache]

**Subflows**

- [Update Cache] The local cache is updated since information has changed since the last pull.

**Alternative Flows**

- [Access Denied] The repository is private and the user is either not logged in or does not have access.
- [Tag Not Found] Tag does not exist in the specified repository.

---

### [UC 25] CLI Add Tag

**Preconditions**

- The user is logged in.
- The user has write access to the repository (owned, or shared with)

**Main Flow**

- The user enters the add tag command in the CLI specifying owner, name, and tag along with the target tag. [Access Denied] [Invalid Tag]
- The cache is updated if operation was successful [Update Cache]

**Subflows**

- [Update Cache] The local cache is updated since information has changed since the last pull.

**Alternative Flows**

- [Access Denied] The repository is private and the user is either not logged in or does not have access.
- [Invalid Tag] If the tag name is invalid (special characters), an error message is displayed telling the user which characters were invalid.

---

### [UC 26] CLI Save Model

**Preconditions**

- The user has read access to the repository (public, owned, or shared with)

**Main Flow**

- The user enters the save command in the CLI specifying owner, name, and tag along with desired path. [Access Denied] [Tag Not Found]
- The cache is updated if operation was successful [Update Cache].
- A local working copy of the model is saved to the specified path.

**Subflows**

- [Update Cache] The local cache is updated since information has changed since the last pull.

**Alternative Flows**

- [Access Denied] The repository is private and the user is either not logged in or does not have access.
- [Tag Not Found] The repository does not contain the specified tag.

---

## Local Actions

### [UC 27] CLI Validate Model

**Main Flow**

- The user enters the validate command in the CLI specifying the local file. [Invalid Model]
- The validation results are returned to the user along with any error messages.

**Alternative Flows**

- [Invalid Model] If the model file fails OpenDI CDM schema validation, the upload operation is cancelled and an error message is displayed showing the line(s) of the file that caused the error.

---

### [UC 28] CLI Diff Models

**Preconditions**

- The user has read access to the specified remote repositories (owned, or shared with)

**Main Flow**

- The user enters the diff command in the CLI specifying either 2 remote models (owner, name, and tag), or a remote and a local model file. [Access Denied] [Invalid Tag] [Update Cache]
- The JSON diff between the specified files is displayed.

**Subflows**

- [Update Cache] The local cache is updated since information has changed since the last pull.

**Alternative Flows**

- [Access Denied] The repository is private and the user is either not logged in or does not have access.
- [Tag Not Found] The repository does not contain the specified tag.

---

### [UC 29] CLI List Local Models

**Main Flow**

- The user uses the list local models command in the CLI.
- Any locally cached models and their information is displayed.

---

### [UC 30] CLI Remove Local Models

**Main Flow**

- The user uses the remove local models command in the CLI along with the id or tag information. [Tag Not Found]
- The selected model is removed from local cache.

**Alternative Flows**

- [Tag Not Found] The specified tag does not exist in local cache.
