package controller

import (
	"fmt"
	"os"

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

// CreateFolder creates a new folder at the specified path
func CreateFolder(path string) error {
	return os.MkdirAll(path, os.ModeDir) // Adjust permissions as needed
}

// Rename renames a file or folder at the old path to the new path
func Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
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
