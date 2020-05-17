package rabbit

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	types "github.com/mapofzones/txs-processor/pkg/types"

	"github.com/streadway/amqp"
	"golang.org/x/net/context"
)

// BlockStream creates individual connection to rabbitmq and returns read-only block channel
func BlockStream(ctx context.Context, addr, queueName string) (<-chan types.Block, error) {
	msgs, err := connect(ctx, addr, queueName)
	if err != nil {
		return nil, fmt.Errorf("could not connect to rabbitmq, %s", err.Error())
	}
	return msgToBlocks(ctx, msgs), nil
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
	if err != nil {
		return nil, err
	}

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
		<-ctx.Done()
		// give last consumer time to read data from our channel
		time.Sleep(5 * time.Second)
		ch.Close()
		conn.Close()
	}()
	return msgs, nil
}

// msgToBlocks processes raw messages and transforms them blocks for further processing
func msgToBlocks(ctx context.Context, msgs <-chan amqp.Delivery) <-chan types.Block {
	blocks := make(chan types.Block)

	go func() {
		defer close(blocks)
		for {
			block := types.Block{}
			msg, ok := <-msgs
			if !ok {
				return
			}
			err := json.Unmarshal(msg.Body, &block)
			// if we received invalid block, we can just skip it because history plugin will fetch the blocks anyway
			if err != nil {
				log.Println(err)
				return
			}
			select {
			case blocks <- block:
			case <-ctx.Done():
				return
			}
		}
	}()
	return blocks
}
