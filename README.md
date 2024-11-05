# Google Drive Shared File Copier

### What dis?

Are you a Google Workspace Admin?
Do you have Users whose work depends on files that were shared with them by accounts outside the Organization?
Fret not! This program will help you copy those files to a folder owned by the Organization.

### How does it work?

1. The program will impersonate every user in your Google Workspace domain.
2. Then it will list all the files that were shared with that user.
3. Then it will copy those files to a folder owned by that user.
4. This ensures that all shared files have at least a copy owned by the Organization.

### How to use?

#### Set up Service Account with Domain-wide Delegation

([Reference Documentation](https://developers.google.com/workspace/guides/create-credentials#google-cloud-console))

1. Go [Here](https://console.cloud.google.com/projectselector2/apis/library) and enable the `Google Drive API` and the
   `Admin SDK API` for your project.
2. Go [Here](https://console.cloud.google.com/projectselector2/iam-admin/serviceaccounts) and create a new Service
   Account in your project (no need for extra permissions, just click through)
3. Once created, click on the service account, and create a new key. (Tab `Keys` -> `Add Key` -> `Create new key` ->
   `JSON`)
4. Save the JSON file that was downloaded to your computer.
5. In the Service Account in the `Details` tab, expand the section on the bottom named `Advanced settings`.
6. Under `Domain-wide delegation`, click on `Enable domain-wide delegation`.
7. Copy the `Client ID` of the service account (some long string of numbers right under where the `enable` button was.
8. In the Admin Console: `Security` -> `Access and Data control` -> `API Controls` -> `Manage Domain-wide Delegation` (
   or click [here](https://admin.google.com/u/1/ac/owl/domainwidedelegation))
9. Click on `Add new` and paste the `Client ID` you copied earlier, and the following OAuth scopes:
    - `https://www.googleapis.com/auth/drive`
    - `https://www.googleapis.com/auth/documents`
    - `https://www.googleapis.com/auth/spreadsheets`
    - `https://www.googleapis.com/auth/presentations`
    - `https://www.googleapis.com/auth/forms`
    - `https://www.googleapis.com/auth/admin.directory.user.readonly`
    - `https://www.googleapis.com/auth/gmail.readonly`
    - (these scopes might be overkill, but I couldn't be bothered to figure out the minimum required - this works tho)
10. Configure the OAuth consent screen as an `internal` app (just fill in required fields, no need for verification).
11. Add the same 7 scopes as above to the consent screen.

#### Run the program

1. Clone this repository
2. Make sure you have Golang 1.23 installed
3. Run `go mod download` in the repository folder
4. Run
   `go run . -credentials /path/to/your/credentials.json -admin your-admin-user@domain.com -skip-user-substr user-to-skip@domain.com,@some-other-domain.com,@some-part-of-a-domain`
5. Wait. This will take a (long) time. Make sure your computer doesn't go to sleep and that the screensaver doesn't turn
   on. I'm not sure this will keep running. Also, consider running this over night, since with many users it will take
   LONG!
