# skyzcopy

Pronounced sky-zed-copy

The AIX implementation of Azure Storage SDK

## Usage
Environment variables for the Storage Account and Access Key must be set:

```
export AZURE_STORAGE_ACCOUNT="xxxx"
export AZURE_STORAGE_ACCESS_KEY="xxxxx"
```

### Running the utility:

Please pass a filename or directory as a parameter to the executable. You can also specify an existing container (although this is optional):
```./skyzcopy_upload <filename_or_directory> [existing_container]```
