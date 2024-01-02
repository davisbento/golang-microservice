package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/davisbento/golang-microservice/utils"
	"github.com/streadway/amqp"
)

func main() {
	conn, err := amqp.Dial(utils.RabbitMQURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	// Declare the exchange
	err = ch.ExchangeDeclare(
		utils.ExchangeName, // name
		"direct",           // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	queue, err := ch.QueueDeclare(
		utils.QueueName, // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)

	failOnError(err, "Failed to declare a queue")

	// Bind the queue to the exchange
	err = ch.QueueBind(
		queue.Name,         // queue name
		utils.RoutingKey,   // routing key
		utils.ExchangeName, // exchange
		false,              // no-wait
		nil,                // arguments
	)
	failOnError(err, "Failed to bind a queue to the exchange")

	msgs, err := ch.Consume(
		queue.Name,        // queue
		utils.ConsumerTag, // consumer
		false,             // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	failOnError(err, "Failed to register a consumer")

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopChan:
				return
			case delivery, ok := <-msgs:
				if ok {
					processMessage(delivery)
				}
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	wg.Wait()
}

func processMessage(delivery amqp.Delivery) {
	log.Printf("Received a message: %s", delivery.Body)

	// Simulate processing time
	time.Sleep(2 * time.Second)

	log.Printf("Done processing message: %s", delivery.Body)

	// Acknowledge the message
	err := delivery.Ack(false)
	failOnError(err, "Failed to acknowledge message")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
