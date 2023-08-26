import pika
import pika.channel
import pika.adapters.blocking_connection

__CHANNEL : pika = None

def getRabbitChannel() -> pika.adapters.blocking_connection.BlockingChannel :
    global __CHANNEL
    if (__CHANNEL == None) :
        connection = pika.BlockingConnection(
            pika.ConnectionParameters(host='rabbit_mq', port = '5672')
        )

        channel = connection.channel()
        __CHANNEL = channel
    
    return __CHANNEL