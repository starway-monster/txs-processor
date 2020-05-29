package processor

import (
	"context"
	"log"
	"os"

	"errors"

	watcher "github.com/mapofzones/cosmos-watcher/pkg/types"
	rabbit "github.com/mapofzones/txs-processor/pkg/rabbit_mq"
	processor "github.com/mapofzones/txs-processor/pkg/types"
	postgres "github.com/mapofzones/txs-processor/pkg/x/postgres"
)

// Processor holds handles for all our connections
// and holds an interface which defines what has to be done
// on the received block
type Processor struct {
	Blocks <-chan watcher.Block
	processor.Processor
}

// NewProcessor returns instance of initialized processor and error if something goes wrong
func NewProcessor(ctx context.Context, amqpEndpoint, queueName string) (*Processor, error) {
	txs, err := rabbit.BlockStream(ctx, amqpEndpoint, queueName)
	if err != nil {
		return nil, err
	}

	impl, err := postgres.NewProcessor(ctx, os.Getenv("postgres"))
	if err != nil {
		return nil, err
	}

	return &Processor{
		Blocks:    txs,
		Processor: impl,
	}, nil
}

// Process consumes and processes transactions from rabbitmq
func (p *Processor) Process(ctx context.Context) error {
	// here we store chains from which we could not decode txs
	badChains := map[string]bool{}

	// receive and send txs
	for {
		select {
		case block, ok := <-p.Blocks:
			if !ok {
				return errors.New("block channel is closed")
			}

			// we won't be able to do anything valid with this blockchain
			if badChains[block.ChainID()] {
				continue
			}

			if err := p.ProcessBlock(ctx, block); err != nil {
				// if we can't decode txs from this chain
				// or order of blocks is messed up, ignore it until broker restart
				if errors.Is(err, processor.BlockHeightError) {
					badChains[block.ChainID()] = true
				}

				// if we have error in our logic or there is no connection
				if errors.Is(err, processor.ConnectionError) ||
					errors.Is(err, processor.CommitError) {
					return err
				}
				log.Printf("could not process block from %s: %s\n", block.ChainID(), err)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (p *Processor) ProcessBlock(ctx context.Context, block watcher.Block) error {
	err := p.Validate(ctx, block)
	if err != nil {
		return err
	}

	for _, message := range block.Messages() {
		handler := p.Handler(message)
		if handler != nil {
			err := handler(ctx, processor.MessageMetadata{
				ChainID:   block.ChainID(),
				BlockTime: block.Time(),
			}, message)
			if err != nil {
				return err
			}
		}
	}

	return p.Commit(ctx, block)
}
