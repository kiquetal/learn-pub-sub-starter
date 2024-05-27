### Running rabbitmq-stomp
docker run -d --rm --name rabbitmq-stomp -p 61613:61613 -p 5672:5672 -p 15672:15672  rabbitmq-stomp:latest

### Testing connection
nc -vz localhost 61613


