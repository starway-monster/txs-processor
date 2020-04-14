// Package rabbit gives us []Tx channel interface to get transactions from rabbitmq endpoint.
package rabbit

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	types "test/types"

	"github.com/streadway/amqp"
	"golang.org/x/net/context"
)

type (
	Tx  = types.Tx
	Txs = types.Txs
)

// TxQueue creates individual connection to rabbitmq and returns transaction slices
func TxQueue(ctx context.Context, addr, queueName string) (<-chan Txs, error) {
	msgs, err := connect(ctx, addr, queueName)
	if err != nil {
		return nil, fmt.Errorf("could not connect to rabitmq, %s", err.Error())
	}
	return msgToTx(ctx, msgs), nil
}

func connect(ctx context.Context, addr, queueName string) (<-chan amqp.Delivery, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, err
	}
	// here we monitor our context
	go func() {
		select {
		case <-ctx.Done():
			// give last consumer to read data from our channel
			time.Sleep(10 * time.Second)
			ch.Close()
			conn.Close()
		}
	}()
	return msgs, nil
}

// msgToTx processes raw messages and transforms them to Tx slices
func msgToTx(ctx context.Context, msgs <-chan amqp.Delivery) <-chan Txs {
	txs := make(chan Txs)
	// error channel has a large capacity because waiting to deliver on error channel would mean

	go func() {
		// close channels, becouse something bad happened if we exited here
		defer close(txs)

		for {
			select {
			case msg := <-msgs:
				data := []Tx{}
				err := json.Unmarshal(msg.Body, &data)
				// if data is invalid we either send the error to reciever, or if it's not possible, we log it and exit
				if err != nil {
					select {
					// send the error to the reciever
					case txs <- Txs{Err: MarshalError(err, msg.Body)}:
						continue
					// log invalid tx data for debbuging purposes before exiting
					case <-ctx.Done():
						log.Printf("invalid unprocessed tx data: %s", msg.Body)
						return
					}
				}

				// send valid data, with error signifying that we recieved normally from rabbit
				select {
				// send message to be processed down the line, everything is fine
				case txs <- Txs{Txs: data, Err: nil}:
					continue
				// something bad happend, try send message back to queue, if it's impossible atleast log the recieved data
				case <-ctx.Done():
					// we reject the message, sending it back to the queue
					err := msg.Reject(true)
					// we could not send the message back, at least log the error
					if err != nil {
						log.Printf("could not send messages back to rabbitmq,message data is: %s\n%s", msg.Body, err)
					}
					return
				}
			// if we are here, then we didn't read any msgs from queue and can shutdown gracefully
			case <-ctx.Done():
				return
			}
		}
	}()

	return txs
}
