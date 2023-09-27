import os, boto3
from engineering import MyErrors
from engineering.Method import Method
from utils import Utils

def serveRequestDirectlyFromS3(requestType : Method, fileName : str, username: str):
    try:
        match requestType:
            case Method.UPLOAD:
                uploadFileToS3(fileName=fileName, username=username)
            case Method.DOWNLOAD:
                downloadFileFromS3(fileName=fileName)
            case Method.DELETE:
                deleteFileFromS3(fileName=fileName)
    except IOError as e:
        raise MyErrors.FileNotFoundException("Il file non esiste nella memoria locale")
    except Exception as e:
        raise MyErrors.S3Exception(str(e))

def downloadFileFromS3(fileName : str) :
    s3 = boto3.client('s3')
    s3 = boto3.Session()
    with open(os.environ.get("FILES_PATH") + fileName, 'wb') as f:
        s3.download_fileobj(os.environ.get("S3_BUCKET_NAME"), fileName, f)

def uploadFileToS3(fileName : str, username : str):
    s3 = boto3.client('s3')
    with open(os.environ.get("FILES_PATH") + fileName, "rb") as f:
        s3.upload_fileobj(f, os.environ.get("S3_BUCKET_NAME"), Utils.buildUploadFileName(fileName=fileName, username=username))

def deleteFileFromS3(fileName : str):
    s3 = boto3.client('s3')
    s3.delete_object(os.environ.get("S3_BUCKET_NAME"), fileName)