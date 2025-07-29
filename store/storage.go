package store

import (
	"bytes"
	"fm/models"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Client struct {
	client     *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	bucket     string
}

func NewS3Client(config models.S3Config) (*S3Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:                        aws.String(config.Region),
		Endpoint:                      aws.String(config.Endpoint),
		Credentials:                   credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
		S3ForcePathStyle:              aws.Bool(true),
		S3DisableContentMD5Validation: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	client := s3.New(sess)
	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)

	return &S3Client{
		client:     client,
		uploader:   uploader,
		downloader: downloader,
		bucket:     config.Bucket,
	}, nil
}

// CreateFolder creates a new folder (directory) in S3
func (s3c *S3Client) CreateFolder(folderPath string) error {
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}

	_, err := s3c.client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3c.bucket),
		Key:    aws.String(folderPath),
		Body:   strings.NewReader(""),
	})
	if err != nil {
		return fmt.Errorf("failed to create folder '%s': %w", folderPath, err)
	}

	fmt.Printf("✅ Folder created: %s\n", folderPath)
	return nil
}

// ===================== PRESIGNED URL OPERATIONS =====================

// GeneratePresignedUploadURL creates a presigned URL for direct file uploads
func (s3c *S3Client) GeneratePresignedUploadURL(options models.PresignedUploadOptions) (*models.PresignedURL, error) {
	if options.ExpiryMinutes <= 0 {
		options.ExpiryMinutes = 60 // Default 1 hour
	}

	if options.ContentType == "" {
		options.ContentType = "application/octet-stream"
	}

	expiry := time.Duration(options.ExpiryMinutes) * time.Minute

	req, _ := s3c.client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(s3c.bucket),
		Key:         aws.String(options.FilePath),
		ContentType: aws.String(options.ContentType),
	})

	// Add content length constraint if specified
	if options.MaxFileSize > 0 {
		req.HTTPRequest.Header.Set("Content-Length", fmt.Sprintf("%d", options.MaxFileSize))
	}

	url, err := req.Presign(expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	presignedURL := &models.PresignedURL{
		URL:        url,
		ExpiresAt:  time.Now().Add(expiry),
		HTTPMethod: "PUT",
		FilePath:   options.FilePath,
	}

	fmt.Printf("✅ Presigned upload URL generated: %s (expires in %d minutes)\n", options.FilePath, options.ExpiryMinutes)
	return presignedURL, nil
}

// GeneratePresignedDownloadURL creates a presigned URL for direct file downloads
func (s3c *S3Client) GeneratePresignedDownloadURL(options models.PresignedDownloadOptions) (*models.PresignedURL, error) {
	if options.ExpiryMinutes <= 0 {
		options.ExpiryMinutes = 60 // Default 1 hour
	}

	expiry := time.Duration(options.ExpiryMinutes) * time.Minute

	req, _ := s3c.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s3c.bucket),
		Key:    aws.String(options.FilePath),
	})

	// Add custom filename for download if specified
	if options.Filename != "" {
		req.HTTPRequest.Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", options.Filename))
	}

	url, err := req.Presign(expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned download URL: %w", err)
	}

	presignedURL := &models.PresignedURL{
		URL:        url,
		ExpiresAt:  time.Now().Add(expiry),
		HTTPMethod: "GET",
		FilePath:   options.FilePath,
	}

	fmt.Printf("✅ Presigned download URL generated: %s (expires in %d minutes)\n", options.FilePath, options.ExpiryMinutes)
	return presignedURL, nil
}

// GeneratePresignedDeleteURL creates a presigned URL for direct file deletion
func (s3c *S3Client) GeneratePresignedDeleteURL(filePath string, expiryMinutes int) (*models.PresignedURL, error) {
	if expiryMinutes <= 0 {
		expiryMinutes = 60 // Default 1 hour
	}

	expiry := time.Duration(expiryMinutes) * time.Minute

	req, _ := s3c.client.DeleteObjectRequest(&s3.DeleteObjectInput{
		Bucket: aws.String(s3c.bucket),
		Key:    aws.String(filePath),
	})

	url, err := req.Presign(expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned delete URL: %w", err)
	}

	presignedURL := &models.PresignedURL{
		URL:        url,
		ExpiresAt:  time.Now().Add(expiry),
		HTTPMethod: "DELETE",
		FilePath:   filePath,
	}

	fmt.Printf("✅ Presigned delete URL generated: %s (expires in %d minutes)\n", filePath, expiryMinutes)
	return presignedURL, nil
}

// GenerateBatchPresignedUploadURLs generates multiple presigned URLs for batch uploads
func (s3c *S3Client) GenerateBatchPresignedUploadURLs(options []models.PresignedUploadOptions) ([]*models.PresignedURL, error) {
	var urls []*models.PresignedURL

	for _, option := range options {
		url, err := s3c.GeneratePresignedUploadURL(option)
		if err != nil {
			return nil, fmt.Errorf("failed to generate presigned URL for %s: %w", option.FilePath, err)
		}
		urls = append(urls, url)
	}

	fmt.Printf("✅ Generated %d presigned upload URLs\n", len(urls))
	return urls, nil
}

// GenerateBatchPresignedDownloadURLs generates multiple presigned URLs for batch downloads
func (s3c *S3Client) GenerateBatchPresignedDownloadURLs(options []models.PresignedDownloadOptions) ([]*models.PresignedURL, error) {
	var urls []*models.PresignedURL

	for _, option := range options {
		url, err := s3c.GeneratePresignedDownloadURL(option)
		if err != nil {
			return nil, fmt.Errorf("failed to generate presigned URL for %s: %w", option.FilePath, err)
		}
		urls = append(urls, url)
	}

	fmt.Printf("✅ Generated %d presigned download URLs\n", len(urls))
	return urls, nil
}

// TestPresignedUpload tests a presigned upload URL by uploading sample content
func (s3c *S3Client) TestPresignedUpload(presignedURL *models.PresignedURL, content []byte) error {
	req, err := http.NewRequest("PUT", presignedURL.URL, bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload via presigned URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload failed with status: %s", resp.Status)
	}

	fmt.Printf("✅ File uploaded successfully via presigned URL: %s\n", presignedURL.FilePath)
	return nil
}

// TestPresignedDownload tests a presigned download URL by downloading content
func (s3c *S3Client) TestPresignedDownload(presignedURL *models.PresignedURL) ([]byte, error) {
	resp, err := http.Get(presignedURL.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to download via presigned URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status: %s", resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("✅ File downloaded successfully via presigned URL: %s (%d bytes)\n", presignedURL.FilePath, len(content))
	return content, nil
}

// GetPresignedURLInfo returns information about a presigned URL
func (s3c *S3Client) GetPresignedURLInfo(presignedURL *models.PresignedURL) {
	fmt.Printf("📋 Presigned URL Info:\n")
	fmt.Printf("   File Path: %s\n", presignedURL.FilePath)
	fmt.Printf("   HTTP Method: %s\n", presignedURL.HTTPMethod)
	fmt.Printf("   Expires At: %s\n", presignedURL.ExpiresAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("   Time Remaining: %v\n", time.Until(presignedURL.ExpiresAt))
	fmt.Printf("   URL: %s\n", presignedURL.URL)
}

// IsPresignedURLExpired checks if a presigned URL has expired
func (s3c *S3Client) IsPresignedURLExpired(presignedURL *models.PresignedURL) bool {
	return time.Now().After(presignedURL.ExpiresAt)
}

// ListFolder lists all items in a folder
func (s3c *S3Client) ListFolder(folderPath string) ([]models.FileInfo, error) {
	if folderPath != "" && !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s3c.bucket),
		Prefix: aws.String(folderPath),
		//Delimiter: aws.String("/"),
	}

	result, err := s3c.client.ListObjectsV2(input)
	if err != nil {
		return nil, fmt.Errorf("failed to list folder '%s': %w", folderPath, err)
	}

	var files []models.FileInfo
	for _, obj := range result.Contents {
		fileInfo := models.FileInfo{
			Key:          *obj.Key,
			Size:         *obj.Size,
			LastModified: *obj.LastModified,
			ETag:         strings.Trim(*obj.ETag, "\""),
			IsFolder:     strings.HasSuffix(*obj.Key, "/"),
		}
		files = append(files, fileInfo)
	}

	return files, nil
}

// DeleteFolder deletes a folder and all its contents
func (s3c *S3Client) DeleteFolder(folderPath string) error {
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}

	// First, list all objects in the folder
	files, err := s3c.ListFolder(folderPath)
	if err != nil {
		return fmt.Errorf("failed to list folder contents: %w", err)
	}

	// Delete all objects in the folder
	for _, file := range files {
		if err := s3c.DeleteFile(file.Key); err != nil {
			return fmt.Errorf("failed to delete file '%s': %w", file.Key, err)
		}
	}

	fmt.Printf("✅ Folder deleted: %s\n", folderPath)
	return nil
}

// MoveFolder moves/renames a folder
func (s3c *S3Client) MoveFolder(sourcePath, destPath string) error {
	if !strings.HasSuffix(sourcePath, "/") {
		sourcePath += "/"
	}
	if !strings.HasSuffix(destPath, "/") {
		destPath += "/"
	}

	// List all files in source folder
	files, err := s3c.ListFolder(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to list source folder: %w", err)
	}

	// Copy each file to destination and delete from source
	for _, file := range files {
		if file.IsFolder {
			continue // Skip folder markers
		}

		newKey := strings.Replace(file.Key, sourcePath, destPath, 1)
		if err := s3c.CopyFile(file.Key, newKey); err != nil {
			return fmt.Errorf("failed to copy file '%s': %w", file.Key, err)
		}

		if err := s3c.DeleteFile(file.Key); err != nil {
			return fmt.Errorf("failed to delete original file '%s': %w", file.Key, err)
		}
	}

	fmt.Printf("✅ Folder moved: %s -> %s\n", sourcePath, destPath)
	return nil
}

// ===================== FILE OPERATIONS =====================

// CreateFile creates a new file with content
func (s3c *S3Client) CreateFile(filePath string, content []byte) error {
	_, err := s3c.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3c.bucket),
		Key:    aws.String(filePath),
		Body:   bytes.NewReader(content),
	})
	if err != nil {
		return fmt.Errorf("failed to create file '%s': %w", filePath, err)
	}

	fmt.Printf("✅ File created: %s (%d bytes)\n", filePath, len(content))
	return nil
}

// ReadFile reads and returns file content
func (s3c *S3Client) ReadFile(filePath string) ([]byte, error) {
	buf := aws.NewWriteAtBuffer([]byte{})

	_, err := s3c.downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(s3c.bucket),
		Key:    aws.String(filePath),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w", filePath, err)
	}

	fmt.Printf("✅ File read: %s (%d bytes)\n", filePath, len(buf.Bytes()))
	return buf.Bytes(), nil
}

// UpdateFile updates existing file content
func (s3c *S3Client) UpdateFile(filePath string, content []byte) error {
	return s3c.CreateFile(filePath, content) // S3 overwrites by default
}

// DeleteFile deletes a file
func (s3c *S3Client) DeleteFile(filePath string) error {
	_, err := s3c.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s3c.bucket),
		Key:    aws.String(filePath),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file '%s': %w", filePath, err)
	}

	fmt.Printf("✅ File deleted: %s\n", filePath)
	return nil
}

// CopyFile copies a file from source to destination
func (s3c *S3Client) CopyFile(sourcePath, destPath string) error {
	copySource := fmt.Sprintf("%s/%s", s3c.bucket, sourcePath)

	_, err := s3c.client.CopyObject(&s3.CopyObjectInput{
		Bucket:     aws.String(s3c.bucket),
		CopySource: aws.String(copySource),
		Key:        aws.String(destPath),
	})
	if err != nil {
		return fmt.Errorf("failed to copy file '%s' to '%s': %w", sourcePath, destPath, err)
	}

	fmt.Printf("✅ File copied: %s -> %s\n", sourcePath, destPath)
	return nil
}

// MoveFile moves/renames a file
func (s3c *S3Client) MoveFile(sourcePath, destPath string) error {
	// Copy file to new location
	if err := s3c.CopyFile(sourcePath, destPath); err != nil {
		return err
	}

	// Delete original file
	if err := s3c.DeleteFile(sourcePath); err != nil {
		return err
	}

	fmt.Printf("✅ File moved: %s -> %s\n", sourcePath, destPath)
	return nil
}

// GetFileInfo gets metadata about a file
func (s3c *S3Client) GetFileInfo(filePath string) (*models.FileInfo, error) {
	result, err := s3c.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s3c.bucket),
		Key:    aws.String(filePath),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
			return nil, fmt.Errorf("file not found: %s", filePath)
		}
		return nil, fmt.Errorf("failed to get file info '%s': %w", filePath, err)
	}

	return &models.FileInfo{
		Key:          filePath,
		Size:         *result.ContentLength,
		LastModified: *result.LastModified,
		ETag:         strings.Trim(*result.ETag, "\""),
		IsFolder:     false,
	}, nil
}

// FileExists checks if a file exists
func (s3c *S3Client) FileExists(filePath string) bool {
	_, err := s3c.GetFileInfo(filePath)
	return err == nil
}

// UploadFromLocalFile uploads a local file to S3
func (s3c *S3Client) UploadFromLocalFile(localPath, s3Path string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	_, err = s3c.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3c.bucket),
		Key:    aws.String(s3Path),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	fmt.Printf("✅ File uploaded: %s -> %s (%d bytes)\n", localPath, s3Path, fileInfo.Size())
	return nil
}

// DownloadToLocalFile downloads an S3 file to local filesystem
func (s3c *S3Client) DownloadToLocalFile(s3Path, localPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	_, err = s3c.downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(s3c.bucket),
		Key:    aws.String(s3Path),
	})
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	fmt.Printf("✅ File downloaded: %s -> %s\n", s3Path, localPath)
	return nil
}

// ===================== UTILITY FUNCTIONS =====================

// PrintFileInfo prints file information in a readable format
func PrintFileInfo(files []models.FileInfo) {
	fmt.Println("\n📁 Folder Contents:")
	fmt.Println("Type | Size     | Modified             | Name")
	fmt.Println("-----|----------|----------------------|--------------------")

	for _, file := range files {
		fileType := "FILE"
		if file.IsFolder {
			fileType = "DIR "
		}

		sizeStr := fmt.Sprintf("%d B", file.Size)
		if file.Size > 1024*1024 {
			sizeStr = fmt.Sprintf("%.1f MB", float64(file.Size)/(1024*1024))
		} else if file.Size > 1024 {
			sizeStr = fmt.Sprintf("%.1f KB", float64(file.Size)/1024)
		}

		fmt.Printf("%s | %-8s | %s | %s\n",
			fileType,
			sizeStr,
			file.LastModified.Format("2006-01-02 15:04:05"),
			file.Key)
	}
}

// ===================== DEMONSTRATION =====================

func demonstrateOperations(s3Client *S3Client) {
	fmt.Println("🚀 Demonstrating S3 CRUD Operations")
	fmt.Println("=====================================")

	// 1. Create folders
	fmt.Println("\n1. Creating folders...")
	s3Client.CreateFolder("test-app/documents")
	s3Client.CreateFolder("test-app/images")
	s3Client.CreateFolder("test-app/backups")

	// 2. Create files
	fmt.Println("\n2. Creating files...")
	s3Client.CreateFile("test-app/documents/readme.txt", []byte("This is a test document"))
	s3Client.CreateFile("test-app/documents/config.json", []byte(`{"app": "test", "version": "1.0"}`))
	s3Client.CreateFile("test-app/images/placeholder.txt", []byte("Image placeholder"))

	// 3. List folder contents
	fmt.Println("\n3. Listing folder contents...")
	files, err := s3Client.ListFolder("test-app/")
	if err != nil {
		fmt.Printf("❌ Error listing folder: %v\n", err)
	} else {
		PrintFileInfo(files)
	}

	// 4. Read file
	fmt.Println("\n4. Reading file...")
	content, err := s3Client.ReadFile("test-app/documents/readme.txt")
	if err != nil {
		fmt.Printf("❌ Error reading file: %v\n", err)
	} else {
		fmt.Printf("📄 File content: %s\n", string(content))
	}

	// 5. Update file
	fmt.Println("\n5. Updating file...")
	s3Client.UpdateFile("test-app/documents/readme.txt", []byte("Updated content: This is the new version"))

	// 6. Copy file
	fmt.Println("\n6. Copying file...")
	s3Client.CopyFile("test-app/documents/readme.txt", "test-app/backups/readme-backup.txt")

	// 7. Move file
	fmt.Println("\n7. Moving file...")
	s3Client.MoveFile("test-app/documents/config.json", "test-app/backups/config-backup.json")

	// 8. Get file info
	fmt.Println("\n8. Getting file info...")
	info, err := s3Client.GetFileInfo("test-app/documents/readme.txt")
	if err != nil {
		fmt.Printf("❌ Error getting file info: %v\n", err)
	} else {
		fmt.Printf("📊 File info: %s (%d bytes, modified: %s)\n",
			info.Key, info.Size, info.LastModified.Format("2006-01-02 15:04:05"))
	}

	// 9. Check file existence
	fmt.Println("\n9. Checking file existence...")
	exists := s3Client.FileExists("test-app/documents/readme.txt")
	fmt.Printf("📁 File exists: %v\n", exists)

	// 10. Final folder listing
	fmt.Println("\n10. Final folder listing...")
	files, err = s3Client.ListFolder("test-app/")
	if err != nil {
		fmt.Printf("❌ Error listing folder: %v\n", err)
	} else {
		PrintFileInfo(files)
	}

	fmt.Println("\n🎉 All operations completed successfully!")
	fmt.Println("To clean up, uncomment the cleanup section below.")

	// Cleanup (uncomment if you want to delete test files)
	/*
		fmt.Println("\n🧹 Cleaning up...")
		s3Client.DeleteFolder("test-app/")
	*/
}

// ===================== PRESIGNED URL DEMONSTRATION =====================

func demonstratePresignedURLs(s3Client *S3Client) {
	fmt.Println("\n\n🔐 Demonstrating Presigned URL Operations")
	fmt.Println("==========================================")

	// 1. Generate presigned upload URL
	fmt.Println("\n1. Generating presigned upload URL...")
	uploadOptions := models.PresignedUploadOptions{
		FilePath:      "presigned-test/upload-test.txt",
		ContentType:   "text/plain",
		MaxFileSize:   1024 * 1024, // 1MB limit
		ExpiryMinutes: 15,          // 15 minutes
	}

	uploadURL, err := s3Client.GeneratePresignedUploadURL(uploadOptions)
	if err != nil {
		fmt.Printf("❌ Error generating upload URL: %v\n", err)
		return
	}

	s3Client.GetPresignedURLInfo(uploadURL)

	// 2. Test presigned upload
	fmt.Println("\n2. Testing presigned upload...")
	testContent := []byte("This file was uploaded using a presigned URL!")
	if err := s3Client.TestPresignedUpload(uploadURL, testContent); err != nil {
		fmt.Printf("❌ Error testing upload: %v\n", err)
	}

	// 3. Generate presigned download URL
	fmt.Println("\n3. Generating presigned download URL...")
	downloadOptions := models.PresignedDownloadOptions{
		FilePath:      "presigned-test/upload-test.txt",
		ExpiryMinutes: 30,
		Filename:      "downloaded-file.txt", // Custom filename
	}

	downloadURL, err := s3Client.GeneratePresignedDownloadURL(downloadOptions)
	if err != nil {
		fmt.Printf("❌ Error generating download URL: %v\n", err)
		return
	}

	s3Client.GetPresignedURLInfo(downloadURL)

	// 4. Test presigned download
	fmt.Println("\n4. Testing presigned download...")
	downloadedContent, err := s3Client.TestPresignedDownload(downloadURL)
	if err != nil {
		fmt.Printf("❌ Error testing download: %v\n", err)
	} else {
		fmt.Printf("📄 Downloaded content: %s\n", string(downloadedContent))
	}

	// 5. Generate batch presigned URLs
	fmt.Println("\n5. Generating batch presigned upload URLs...")
	batchUploadOptions := []models.PresignedUploadOptions{
		{
			FilePath:      "presigned-test/batch1.txt",
			ContentType:   "text/plain",
			ExpiryMinutes: 60,
		},
		{
			FilePath:      "presigned-test/batch2.json",
			ContentType:   "application/json",
			ExpiryMinutes: 60,
		},
		{
			FilePath:      "presigned-test/batch3.csv",
			ContentType:   "text/csv",
			ExpiryMinutes: 60,
		},
	}

	batchUploadURLs, err := s3Client.GenerateBatchPresignedUploadURLs(batchUploadOptions)
	if err != nil {
		fmt.Printf("❌ Error generating batch upload URLs: %v\n", err)
	} else {
		fmt.Printf("📦 Generated %d batch upload URLs\n", len(batchUploadURLs))
		for i, url := range batchUploadURLs {
			fmt.Printf("   %d. %s (expires: %s)\n", i+1, url.FilePath, url.ExpiresAt.Format("15:04:05"))
		}
	}

	// 6. Generate presigned delete URL
	fmt.Println("\n6. Generating presigned delete URL...")
	deleteURL, err := s3Client.GeneratePresignedDeleteURL("presigned-test/upload-test.txt", 10)
	if err != nil {
		fmt.Printf("❌ Error generating delete URL: %v\n", err)
	} else {
		s3Client.GetPresignedURLInfo(deleteURL)
	}

	// 7. Create files for batch download demo
	fmt.Println("\n7. Creating files for batch download demo...")
	s3Client.CreateFile("presigned-test/file1.txt", []byte("Content of file 1"))
	s3Client.CreateFile("presigned-test/file2.txt", []byte("Content of file 2"))
	s3Client.CreateFile("presigned-test/file3.txt", []byte("Content of file 3"))

	// 8. Generate batch presigned download URLs
	fmt.Println("\n8. Generating batch presigned download URLs...")
	batchDownloadOptions := []models.PresignedDownloadOptions{
		{
			FilePath:      "presigned-test/file1.txt",
			ExpiryMinutes: 30,
			Filename:      "custom-file1.txt",
		},
		{
			FilePath:      "presigned-test/file2.txt",
			ExpiryMinutes: 30,
			Filename:      "custom-file2.txt",
		},
		{
			FilePath:      "presigned-test/file3.txt",
			ExpiryMinutes: 30,
			Filename:      "custom-file3.txt",
		},
	}

	batchDownloadURLs, err := s3Client.GenerateBatchPresignedDownloadURLs(batchDownloadOptions)
	if err != nil {
		fmt.Printf("❌ Error generating batch download URLs: %v\n", err)
	} else {
		fmt.Printf("📦 Generated %d batch download URLs\n", len(batchDownloadURLs))
		for i, url := range batchDownloadURLs {
			fmt.Printf("   %d. %s -> %s (expires: %s)\n", i+1, url.FilePath,
				batchDownloadOptions[i].Filename, url.ExpiresAt.Format("15:04:05"))
		}
	}

	// 9. Show practical usage examples
	fmt.Println("\n9. Practical usage examples...")
	fmt.Println("📋 Frontend JavaScript Examples:")
	fmt.Println("   // Upload file using presigned URL")
	fmt.Printf("   fetch('%s', {\n", uploadURL.URL)
	fmt.Println("     method: 'PUT',")
	fmt.Println("     body: fileContent,")
	fmt.Println("     headers: { 'Content-Type': 'text/plain' }")
	fmt.Println("   });")
	fmt.Println()
	fmt.Println("   // Download file using presigned URL")
	fmt.Printf("   fetch('%s')\n", downloadURL.URL)
	fmt.Println("     .then(response => response.blob())")
	fmt.Println("     .then(blob => { /* handle downloaded file */ });")

	fmt.Println("\n🎉 Presigned URL operations completed!")
	fmt.Println("💡 Use these URLs in your frontend applications for secure file operations.")

	// Cleanup demo files
	fmt.Println("\n🧹 Cleaning up presigned demo files...")
	s3Client.DeleteFolder("presigned-test/")
}
