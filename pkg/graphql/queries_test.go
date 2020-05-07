package graphql

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mapofzones/txs-processor/pkg/types"
)

func TestLastProcessedBlock(t *testing.T) {
	block, err := LastProcessedBlock(context.Background(), "gameofzoneshub-1a")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(block)
}

func TestIbcStatsExist(t *testing.T) {
	exists, err := IbcStatsExist(context.Background(), "testchain", "otherchain",
		types.FromTimestamp(types.Timestamp("2020-04-21T16:00:00")))

	if err != nil {
		t.Fatal(err)
	}

	if exists != true {
		t.Fatal("expected true")
	}

	exists, err = IbcStatsExist(context.Background(), "notexist", "not even in your dreams", time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if exists != false {
		t.Fatal("expected false")
	}
}

func TestСonnectionIDFromChannelID(t *testing.T) {
	connectionID, err := ConnectionIDFromChannelID(context.Background(), "testchain", "testchannel")
	if err != nil {
		t.Fatal(err)
	}
	if connectionID != "testconnection" {
		t.Fatal("expected to get testconnection,got: ", connectionID)
	}
}

func TestClientIDFromConnectionID(t *testing.T) {
	clientID, err := ClientIDFromConnectionID(context.Background(), "testchain", "testconnection")
	if err != nil {
		t.Fatal(err)
	}
	if clientID != "testclient" {
		t.Fatal("expected to get testclient,got: ", clientID)
	}
}

func TestChainIDFromClientID(t *testing.T) {
	chainID, err := ChainIDFromClientID(context.Background(), "testchain", "testclient")
	if err != nil {
		t.Fatal(err)
	}

	if chainID != "otherchain" {
		t.Fatal("expected to get otherchain,got: ", chainID)
	}
}

func TestChainIDFromConnectionID(t *testing.T) {
	chainID, err := ChainIDFromConnectionID(context.Background(), "testchain", "testconnection")
	if err != nil {
		t.Fatal(err)
	}

	if chainID != "otherchain" {
		t.Fatal("expected to get otherchain,got: ", chainID)
	}
}

func TestChainIDFromChannelID(t *testing.T) {
	chainID, err := ChainIDFromChannelID(context.Background(), "testchain", "testchannel")
	if err != nil {
		t.Fatal(err)
	}

	if chainID != "otherchain" {
		t.Fatal("expected to get otherchain,got: ", chainID)
	}
}
