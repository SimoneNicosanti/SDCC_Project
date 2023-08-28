class FileNotFoundException(Exception):
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)

class FailedToOpenException(Exception):
    def __init__(self, message):
        self.message = message
        super().__init__(self.message)

class InvalidTicketException(Exception):
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