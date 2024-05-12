package controller

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/tinkerbaj/gintemp/handler"
	"github.com/tinkerbaj/gintemp/lib/renderer"
)

func GetMedia(c *gin.Context) {
	path := "./public/european_honey/"
	if c.Query("path") != "" {
		path = "./public/european_honey/" + c.Query("path")
	}
	resp, statusCode := handler.GetMedia(path)

	renderer.Render(c, resp, statusCode)
}

func CreateFolder(c *gin.Context) {
	// Get the path parameter
	path := c.Query("path")

	// Sanitize the path to avoid potential security vulnerabilities
	cleanPath := filepath.Clean(path)

	// Construct the full path
	fullPath := "./public/european_honey/" + cleanPath

	// Check if the path already exists
	_, err := os.Stat(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Path doesn't exist, create the folder
			err = os.Mkdir(fullPath, os.ModeDir)
			if err != nil {
				renderer.Render(c, "Error creating folder: "+err.Error(), http.StatusInternalServerError)
				return
			}
			renderer.Render(c, "Folder created", http.StatusOK)
			return
		}
		// Handle other potential errors from os.Stat
		renderer.Render(c, "Error checking path: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Path already exists
	renderer.Render(c, "Folder already exists", http.StatusConflict)
}

func RenameFolder(c *gin.Context) {
	// Get the path parameter
	oldpath := c.Query("old")
	newpath := c.Query("new")

	// Sanitize the path to avoid potential security vulnerabilities
	cleanOldPath := filepath.Clean(oldpath)
	cleanNewPath := filepath.Clean(newpath)
	// Construct the full path
	fullOldPath := "./public/european_honey/" + cleanOldPath
	fullNewPath := "./public/european_honey/" + cleanNewPath

	// Check if the path already exists
	_, err := os.Stat(fullOldPath)
	if err != nil {

		// Path doesn't exist
		renderer.Render(c, "Folder dont exist", http.StatusInternalServerError)
		return

	}

	_, err = os.Stat(fullNewPath)
	if err == nil {

		// Path doesn't exist
		renderer.Render(c, "Folder with this name exist please try another name", http.StatusInternalServerError)
		return

	}

	err = os.Rename(fullOldPath, fullNewPath)
	if err != nil {
		renderer.Render(c, "Something fails on server side", http.StatusInternalServerError)
		return
	}
	// Path already exists
	renderer.Render(c, "Folder renamed successfully", http.StatusOK)
}

// Delete deletes a file or folder at the specified path
func Delete(path string) error {
	return os.RemoveAll(path) // Use with caution, may delete contents recursively
}

// DownloadFile downloads a file from the server to the client (implementation depends on your framework)
func DownloadFile(path string) error {
	// Implement logic to send the file content to the client (e.g., using HTTP)
	fmt.Println("Download functionality not implemented yet")
	return nil
}
