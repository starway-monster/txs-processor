package builder

import (
	"context"
	"os"
	"testing"

	"github.com/machinebox/graphql"
)

func TestBuilder(t *testing.T) {
	b := MutationBuilder{}

	b.AddZone("test")

	// b.CreateTxStats(types.TxStats{
	// 	Zone:  "test",
	// 	Count: 2,
	// 	Hour:  time.Now().Truncate(time.Hour)})

	// b.CreateIbcStats(types.IbcStats{
	// 	Destination: "irishub",
	// 	Count:       2,
	// 	Hour:        time.Now().Truncate(time.Hour),
	// 	Source:      "test",
	// })

	c := graphql.NewClient(os.Getenv("GRAPHQL"))

	err := c.Run(context.Background(), graphql.NewRequest(b.Mutation()), nil)
	if err != nil {
		t.Fatal(err)
	}
}
