# skyzcopy

Pronounced sky-zed-copy as an homage to azcopy

The AIX (and IBM i) implementation of Azure Storage SDK, this application will let you move files into and out of (*coming soon!*) Azure Storage.

## Usage
Environment variables for the Storage Account and Access Key must be set:

```
export AZURE_STORAGE_ACCOUNT="xxxx"
export AZURE_STORAGE_ACCESS_KEY="xxxxx"
```

### Running the utility:

By default, AIX does not contain the CA certificates needed to connect to the blob storage. Please install them with the following command:


```yum -y install ca-certificates```


Please pass a filename or directory as a parameter to the executable. You can also specify an existing container (although this is optional):
```./skyzcopy_upload <filename_or_directory> [existing_container]```
