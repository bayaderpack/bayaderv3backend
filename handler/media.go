package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/tinkerbaj/gintemp/database/model"
)

// FileInfo represents information about a file or folder

// GetFileInfo gets information about a file or folder and its children
// func GetFileInfo(path string) (model.Media, error) {
// 	fileInfo, err := os.Stat(path)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return model.Media{}, err
// 	}

// 	info := model.Media{
// 		Name:     fileInfo.Name(),
// 		IsFolder: fileInfo.IsDir(),
// 	}

// 	if info.IsFolder {
// 		children, err := getChildren(path)
// 		if err != nil {
// 			return info, err
// 		}
// 		info.Children = children
// 	} else {
// 		// Get extension for files
// 		info.Type = getFileExtension(path)
// 	}

// 	return info, nil
// }
func GetFileInfo(path string) (model.Media, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
	  return model.Media{}, err
	}
  
	info := model.Media{
	  Name:  fileInfo.Name(),
	  IsFolder: fileInfo.IsDir(),
	}
  
	if info.IsFolder {
	  // Get children info (without recursion)
	  children, err := getChildrenInfo(path)
	  if err != nil {
		return info, err
	  }
	  info.Children = children
	} else {
		fmt.Println("Why",path)
	  // Get extension for files
	  info.Type = getFileExtension(path)
	}
  
	return info, nil
  }
  
  func getChildrenInfo(path string) ([]model.Media, error) {
	var children []model.Media
  
	files, err := os.ReadDir(path)
	if err != nil {
	  return nil, err
	}
  
	for _, file := range files {
	//   childPath := filepath.Join(path, file.Name())
	  childInfo := model.Media{
		Name:  file.Name(), // Set basic info for child
		IsFolder: file.IsDir(),
		Type: getFileExtension(file.Name()),
	  }
	  children = append(children, childInfo)
	}
  
	return children, nil
  }

// getChildren gets information about child files and folders
func getChildren(path string) ([]model.Media, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var children []model.Media
	for _, file := range files {
		childInfo, err := GetFileInfo(path + string(os.PathSeparator) + file.Name())
		if err != nil {
			return nil, err
		}
		children = append(children, childInfo)
	}

	return children, nil
}

// getFileExtension gets the extension of a file
func getFileExtension(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
	  return "" // Return empty string for files without extension
	}
	return ext[1:] // Skip leading "." only if extension exists
  }


func GetMedia(path string) (httpResponse model.HTTPResponse, httpStatusCode int) {
	info, err := GetFileInfo(path)
	if err != nil {
		fmt.Println("Error:", err)
	}

	jsonData, err := json.Marshal(info)
	if err != nil {
		fmt.Println("Error:", err)
	}


	//convert jsonData to struct type

	err = json.Unmarshal(jsonData, &info)
	if err != nil {
		fmt.Println("Error:", err)
	}

	httpResponse.Message = info
	httpStatusCode = http.StatusOK
	return

	// httpResponse.Message = hobbies
	// httpStatusCode = http.StatusOK
	// return
}

