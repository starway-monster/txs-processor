package rabbit

import (
	"context"
	"os"
	"testing"
)

func TestBlockStream(t *testing.T) {
	blocks, err := BlockStream(context.Background(), os.Getenv("RABBIT"), "blocks")
	if err != nil {
		t.Fatal(err)
	}
	for block := range blocks {
		block.ChainID = ""
	}

}
