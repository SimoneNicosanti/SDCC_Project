import pika, pika.channel, pika.adapters.blocking_connection, os

__CHANNEL : pika = None

def getRabbitChannel() -> pika.adapters.blocking_connection.BlockingChannel :
    global __CHANNEL
    if (__CHANNEL == None) :
        connection = pika.BlockingConnection(
            pika.ConnectionParameters(host='rabbit_mq', port = '5672')
        )

        channel = connection.channel()
        channel.queue_declare(queue = os.environ.get("QUEUE_NAME"), durable = True)
        __CHANNEL = channel
    
    return __CHANNEL