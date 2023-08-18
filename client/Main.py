import pika

connection = pika.BlockingConnection(
    pika.ConnectionParameters(host='rabbit_mq', port = '5672'))

channel = connection.channel()

channel.queue_declare(queue='hello')

for i in range(0, 10) :
    channel.basic_publish(exchange='', routing_key='hello', body = str(i) + ' Hello World!')
    print(" [x] Sent 'Hello World!'")
connection.close()