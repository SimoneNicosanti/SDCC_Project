from enum import Enum


class Method(Enum) :
    GET = 0,
    PUT = 1,
    DEL = 2


class Ticket :

    def __init__(peer_addr : str, self, ticket_id : str) :
        self.peer_addr : str = peer_addr
        self.ticket_id : str = ticket_id
    
