package rabbit

import (
	"fmt"
	"log"
	"time"

	codec "github.com/mapofzones/cosmos-watcher/pkg/codec"
	watcher "github.com/mapofzones/cosmos-watcher/pkg/types"
	"github.com/tendermint/go-amino"

	"github.com/streadway/amqp"
	"golang.org/x/net/context"
)

// BlockStream creates individual connection to rabbitmq and returns read-only block channel
func BlockStream(ctx context.Context, addr, queueName string) (<-chan watcher.Block, error) {
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

	// get one message at a time
	if err := ch.Qos(1, 0, false); err != nil {
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
func msgToBlocks(ctx context.Context, msgs <-chan amqp.Delivery) <-chan watcher.Block {
	blocks := make(chan watcher.Block)
	cdc := amino.NewCodec()
	codec.RegisterTypes(cdc)

	go func() {
		defer close(blocks)
		for {
			var block watcher.Block
			msg, ok := <-msgs
			if !ok {
				return
			}
			err := cdc.UnmarshalJSON(msg.Body, &block)
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
