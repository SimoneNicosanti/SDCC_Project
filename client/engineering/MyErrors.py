class FileNotFoundException(Exception):
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)

class FailedToOpenException(Exception):
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)

class InvalidMetadataException(Exception):
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)

class RequestFailedException(Exception):
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)

class ConnectionFailedException(Exception):
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)

class S3Exception(Exception):
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)

class NoServerAvailableException(Exception):
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)

class LocalFileNotFoundException(Exception):
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)

class UnauthenticatedUserException(Exception):
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)

class UnknownException(Exception):
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)
        
class DownloadFailedException(Exception):
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)