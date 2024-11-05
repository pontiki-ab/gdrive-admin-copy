package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type CLIArgs struct {
	credentialsPath      string
	adminUser            string
	userSubstringsToSkip []string
}

func main() {
	args, err := getCLIArgs()
	if err != nil {
		log.Fatalf("Error getting CLI arguments: %v", err)
	}

	clientConfig, err := getClientConfig(args.credentialsPath, args.adminUser)
	if err != nil {
		log.Fatalf("Error getting client config: %v", err)
	}

	// Create Admin SDK client
	ctx := context.Background()
	adminService, err := admin.NewService(ctx, option.WithHTTPClient(clientConfig.Client(ctx)))
	if err != nil {
		log.Fatalf("Unable to create Admin service: %v", err)
	}

	// List users to test impersonation
	users, err := userList(ctx, adminService)
	if err != nil {
		log.Fatalf("Unable to retrieve users: %v", err)
	}

	// Process each user
	for _, email := range users {
		// Set subject to impersonate user
		clientConfig.Subject = email

		// Drive service
		driveService, err := drive.NewService(ctx, option.WithHTTPClient(clientConfig.Client(ctx)))
		if err != nil {
			log.Printf("Unable to impersonate user %s: %v", email, err)
			continue
		}

		// List, copy, and save shared files for the user
		err = listAndCopySharedFiles(driveService, email, args.userSubstringsToSkip)
		if err != nil {
			log.Printf("Unable to list and copy shared files for user %s: %v", email, err)
		} else {
			fmt.Printf("Shared files for %s saved and copied.\n", email)
		}
	}
}

func getCLIArgs() (CLIArgs, error) {
	// Command-line flag for credentials file
	credentialsPath := flag.String("credentials", "", "Path to the Google credentials JSON file")
	adminUser := flag.String("admin", "", "Admin email for impersonation")
	userSubstringsToSkip := flag.String("skip-user-substr", "", "Comma-separated list of substrings of file owners' email addresses to skip")
	flag.Parse()

	if *credentialsPath == "" {
		return CLIArgs{}, fmt.Errorf("please provide the path to the credentials file using the -credentials flag")
	}
	if *adminUser == "" {
		return CLIArgs{}, fmt.Errorf("please provide the admin email for impersonation using the -admin flag")
	}
	skipSubstrings := []string{}
	if *userSubstringsToSkip != "" {
		skipSubstrings = strings.Split(*userSubstringsToSkip, ",")
	}

	return CLIArgs{
		credentialsPath:      *credentialsPath,
		adminUser:            *adminUser,
		userSubstringsToSkip: skipSubstrings,
	}, nil
}

func getClientConfig(credentialsPath, adminUser string) (*jwt.Config, error) {
	// Load Google credentials
	creds, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("error reading credentials file: %w", err)
	}

	// Parse JWT config and set impersonation subject
	config, err := google.JWTConfigFromJSON(creds, drive.DriveScope, admin.AdminDirectoryUserReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("error creating JWT config: %w", err)
	}
	config.Subject = adminUser

	return config, nil
}

// userList retrieves all users in the Workspace domain.
func userList(ctx context.Context, srv *admin.Service) ([]string, error) {
	var emails []string
	call := srv.Users.List().Customer("my_customer").MaxResults(500).OrderBy("email") //nolint:gomnd,mnd
	err := call.Pages(ctx, func(page *admin.Users) error {
		for _, user := range page.Users {
			emails = append(emails, user.PrimaryEmail)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}
	return emails, nil
}

// createOrGetFolder checks if "my_copied_shared_files" exists, and creates it if not.
func createOrGetFolder(srv *drive.Service) (string, error) {
	query := "name = 'my_copied_shared_files' and mimeType = 'application/vnd.google-apps.folder' and 'root' in parents"
	folderList, err := srv.Files.List().Q(query).Fields("files(id)").Do()
	if err != nil {
		return "", fmt.Errorf("error checking for folder: %w", err)
	}
	if len(folderList.Files) > 0 {
		return folderList.Files[0].Id, nil
	}

	// Folder does not exist, create it
	folder := &drive.File{
		Name:     "my_copied_shared_files",
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{"root"},
	}
	createdFolder, err := srv.Files.Create(folder).Do()
	if err != nil {
		return "", fmt.Errorf("error creating folder: %w", err)
	}
	return createdFolder.Id, nil
}

// listAndCopySharedFiles lists files shared with a user, writes them to a file, and copies them to a folder.
//
//nolint:cyclop
func listAndCopySharedFiles(srv *drive.Service, email string, skipSubstrings []string) error {
	fileName := fmt.Sprintf("%s_shared_files.txt", email)
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	query := "sharedWithMe=true"
	call := srv.Files.List().Q(query).Fields("files(id, name, owners(displayName, emailAddress))")
	files, err := call.Do()
	if err != nil {
		return fmt.Errorf("error listing files: %w", err)
	}

	// Get or create the destination folder
	folderID, err := createOrGetFolder(srv)
	if err != nil {
		return fmt.Errorf("error getting or creating folder: %w", err)
	}

skip_file:
	for _, f := range files.Files {
		// Get owner's name and email
		var ownerInfo string
		if len(f.Owners) > 0 {
			ownerInfo = fmt.Sprintf("%s (%s)", f.Owners[0].DisplayName, f.Owners[0].EmailAddress)
		} else {
			ownerInfo = "Unknown Owner"
		}

		for _, owner := range f.Owners {
			for _, subString := range skipSubstrings {
				if strings.Contains(owner.EmailAddress, subString) {
					log.Printf("Skipping file %s for user %s: Owner is from %s domain", f.Name, email, subString)
					continue skip_file
				}
			}
		}

		// Write file and owner details to the text file
		_, err := file.WriteString(fmt.Sprintf("File ID: %s, Name: %s, Owner: %s\n", f.Id, f.Name, ownerInfo))
		if err != nil {
			return fmt.Errorf("error writing to file: %w", err)
		}

		// Copy file to the "my_copied_shared_files" folder
		target := &drive.File{
			Parents: []string{folderID},
		}
		_, err = srv.Files.Copy(f.Id, target).Do()
		if err != nil {
			log.Printf("Unable to target file %s for user %s: %v", f.Name, email, err)
		} else {
			log.Printf("Copied file %s for user %s\n", f.Name, email)
		}
	}
	return nil
}
