package builder

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/machinebox/graphql"
	"github.com/mapofzones/txs-processor/types"
)

func TestBuilder(t *testing.T) {
	b := MutationBuilder{}

	b.AddZone("test")
	b.PushBlock(types.Block{
		ChainID: "test",
		Height:  1,
		Results: nil,
		T:       time.Now(),
		Txs:     nil,
	})

	b.CreateTxStats(types.TxStats{
		Zone:  "test",
		Count: 2,
		Hour:  time.Now().Truncate(time.Hour)})

	b.CreateIbcStats(types.IbcStats{
		Destination: "irishub",
		Count:       2,
		Hour:        time.Now().Truncate(time.Hour),
		Source:      "test",
	})

	b.PushTransfer(types.Transfer{
		Hash:      "123",
		Matched:   false,
		Quantity:  3,
		Recipient: "B",
		Sender:    "A",
		Timestamp: types.ToTimestamp(time.Now()),
		Token:     "coin",
		Type:      "send",
		Zone:      "test",
	})

	c := graphql.NewClient(os.Getenv("GRAPHQL"))

	var r interface{}

	err := c.Run(context.Background(), graphql.NewRequest(b.Mutation()), r)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(r)
}
