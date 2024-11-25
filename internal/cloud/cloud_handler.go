package cloud

import "fmt"

func uploadLocalFilePathToCloud(localFilePath string) (string, error) {
	return "", nil
}

func downloadCloudFilePathToLocal(cloudFilePath string, localPath string) (string, error) {
	fmt.Printf("successfully uploaded %s to %s\n", cloudFilePath, localPath)
	return localPath, nil
}

func deleteCloudFilePath(cloudFilePath string) error {
	return nil
}
