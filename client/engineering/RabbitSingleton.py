import pika
import pika.channel

__CHANNEL : pika = None

def getRabbitChannel() -> pika.channel.Channel :
    global __CHANNEL
    if (__CHANNEL == None) :
        connection = pika.BlockingConnection(
            pika.ConnectionParameters(host='rabbit_mq', port = '5672')
        )

        channel = connection.channel()
        __CHANNEL = channel
    
    return __CHANNEL