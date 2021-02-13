////
//// This is the Skytap AZCopy Upload utility for AIX and IBM i
//// Created 12/02/2020
////
////
////

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"path/filepath"
	"time"
	"github.com/Azure/azure-storage-blob-go/azblob"
)

func randomString() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return strconv.Itoa(r.Int())
}

func uploadSingleFile(fileName string, containerURL azblob.ContainerURL) {
	ctx := context.Background() // This example uses a never-expiring context
	// Here's how to upload a blob.
	blobURL := containerURL.NewBlockBlobURL(fileName)
	file, err := os.Open(fileName)
	handleErrors(err)

	// You can use the low-level PutBlob API to upload files. Low-level APIs are simple wrappers for the Azure Storage REST APIs.
	// Note that PutBlob can upload up to 256MB data in one shot. Details: https://docs.microsoft.com/en-us/rest/api/storageservices/put-blob
	// Following is commented out intentionally because we will instead use UploadFileToBlockBlob API to upload the blob
	// _, err = blobURL.PutBlob(ctx, file, azblob.BlobHTTPHeaders{}, azblob.Metadata{}, azblob.BlobAccessConditions{})
	// handleErrors(err)

	// The high-level API UploadFileToBlockBlob function uploads blocks in parallel for optimal performance, and can handle large files as well.
	// This function calls PutBlock/PutBlockList for files larger 256 MBs, and calls PutBlob for any file smaller
	fmt.Printf("Uploading the file with blob name: %s\n", fileName)
	_, err = azblob.UploadFileToBlockBlob(ctx, file, blobURL, azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16})
	handleErrors(err)
}

func handleErrors(err error) {
	if err != nil {
		if serr, ok := err.(azblob.StorageError); ok { // This error is a Service-specific
			switch serr.ServiceCode() { // Compare serviceCode to ServiceCodeXxx constants
			case azblob.ServiceCodeContainerAlreadyExists:
				fmt.Println("Received 409. Container already exists")
				return
			}
		}
		log.Fatal(err)
	}
}

func main() {

	// First element in os.Args is always the program name,
	// So we need at least 2 arguments to have a file name argument.
	if len(os.Args) < 2 {
		fmt.Println("Missing parameter, usage as follows: \n ./skyzcopy_upload <filename_or_directory> [container_name]")
		return
	}
	fileNameorDirectory := os.Args[1]


	fmt.Printf("Azure Blob storage Skytap upload\n")

	// From the Azure portal, get your storage account name and key and set environment variables.
	accountName, accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT"), os.Getenv("AZURE_STORAGE_ACCESS_KEY")
	if len(accountName) == 0 || len(accountKey) == 0 {
		log.Fatal("Either the AZURE_STORAGE_ACCOUNT or AZURE_STORAGE_ACCESS_KEY environment variable is not set")
	}

	// Create a default request pipeline using your storage account name and account key.
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Fatal("Invalid credentials with error: " + err.Error())
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	var containerURL azblob.ContainerURL
	// This for loop asks whether the user needs to create a new container or if they want to use an already existing one.
	if len(os.Args) == 3 {
		containerName := os.Args[2]

		// From the Azure portal, get your storage account blob service URL endpoint.
		URL, _ := url.Parse(
			fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName))

		containerURL = azblob.NewContainerURL(*URL, p)
	} else {
		// Create a random string for the new container
		containerName := fmt.Sprintf("ibmupload%s", randomString())


		// From the Azure portal, get your storage account blob service URL endpoint.
		URL, _ := url.Parse(
			fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName))


		// Create a ContainerURL object that wraps the container URL and a request
		// pipeline to make requests.
		containerURL = azblob.NewContainerURL(*URL, p)

		// Create the container
		fmt.Printf("Creating a container named %s\n", containerName)
		ctx := context.Background() // This example uses a never-expiring context
		_, err = containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
		handleErrors(err)

	}


	fileInfo, err := os.Stat(fileNameorDirectory)
	if err != nil {
		handleErrors(err)
	}
	if fileInfo.IsDir() {
		fileList := make([]string, 0)
		e := filepath.Walk(fileNameorDirectory, func(path string, f os.FileInfo, err error) error {
			fileList = append(fileList, path)
			return err
		})
		
		if e != nil {
			panic(e)
		}
	
		for _, file := range fileList {
			fileInfo, _ := os.Stat(file)
			if fileInfo.IsDir() {
			} else {
				uploadSingleFile(file, containerURL)
			}
			//	fmt.Println(file)
		}	
	} else {
		uploadSingleFile(fileNameorDirectory, containerURL)
	}



	// List the container that we have created above
	fmt.Println("Finished uploading! \n\n Listing the blobs in the container:")
	for marker := (azblob.Marker{}); marker.NotDone(); {
		ctx := context.Background() // This example uses a never-expiring context
		// Get a result segment starting with the blob indicated by the current Marker.
		listBlob, err := containerURL.ListBlobsFlatSegment(ctx, marker, azblob.ListBlobsSegmentOptions{})
		handleErrors(err)

		// ListBlobs returns the start of the next segment; you MUST use this to get
		// the next segment (after processing the current result segment).
		marker = listBlob.NextMarker

		// Process the blobs returned in this result segment (if the segment is empty, the loop body won't execute)
		for _, blobInfo := range listBlob.Segment.BlobItems {
			fmt.Print("	Blob name: " + blobInfo.Name + "\n")
		}
	}
}
