package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/davisbento/golang-microservice/utils"
	"github.com/streadway/amqp"
)

func sendMessageToQueue(message []byte) error {
	conn, err := amqp.Dial(utils.RabbitMQURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		utils.ExchangeName, // name
		"direct",           // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return err
	}

	// Publish the message to the exchange
	err = ch.Publish(
		utils.ExchangeName, // exchange
		utils.RoutingKey,   // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		},
	)
	return err
}

func main() {
	// Define your routes
	http.HandleFunc("/", handleGetRequest)
	http.HandleFunc("/post", handlePostRequest)

	// Start the HTTP server
	port := 8080
	log.Printf("Server listening on :%d...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func handleGetRequest(w http.ResponseWriter, r *http.Request) {
	// Handle GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Your GET request logic here
	fmt.Fprint(w, "Hello, this is a GET request!")
}

func handlePostRequest(w http.ResponseWriter, r *http.Request) {
	// Handle POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Read the body of the POST request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	// Close the request body
	defer r.Body.Close()

	// Send the message to the queue
	err = sendMessageToQueue(body)
	if err != nil {
		http.Error(w, "Error sending message to queue", http.StatusInternalServerError)
		return
	}

	// Respond to the client
	fmt.Fprint(w, "Data received from POST request!")
}
