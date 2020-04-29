package graphql

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mapofzones/txs-processor/types"
)

func TestBlock(t *testing.T) {
	c := NewClient(endpoint)
	err := c.AddBlock(context.Background(), types.Block{ChainID: "irishub", Height: 0})
	if err != nil {
		t.Fatal(err)
	}
}

const endpoint = os.Getenv("GRAPHQL")

func TestSendTransfer(t *testing.T) {
	client := NewClient(endpoint)
	err := client.sendIbcTransfer(context.Background(), types.Transfer{Hash: "123",
		Matched:   false,
		Quantity:  123,
		Recipient: "B",
		Sender:    "A",
		Timestamp: types.ToTimestamp(time.Now()),
		Token:     "coin",
		Type:      "send",
		Zone:      "irishub"})

	if err != nil {
		t.Fatal(err)
	}

	err = client.sendIbcTransfer(context.Background(), types.Transfer{Hash: "1234",
		Matched:   false,
		Quantity:  123,
		Recipient: "A",
		Sender:    "B",
		Timestamp: types.ToTimestamp(time.Now()),
		Token:     "coin",
		Type:      "receive",
		Zone:      "ping-ibc"})

	if err != nil {
		t.Fatal(err)
	}
}

func TestIbcTxExists(t *testing.T) {
	type data struct {
		Source      string
		Destination string
		Hour        string
	}
	cases := []data{
		{
			Source:      "z1",
			Destination: "z2",
			Hour:        "2020-03-25T18:00:00",
		},
	}
	output := []bool{true}

	C := NewClient(endpoint)

	for i, c := range cases {
		x := types.IbcStats{
			Source:      c.Source,
			Destination: c.Destination,
			Hour:        types.FromTimestamp(types.Timestamp(c.Hour)),
		}

		out, err := C.ibcTxExists(context.Background(), x)
		if err != nil {
			t.Fatal(err)
		}
		if out != output[i] {
			t.Fatalf("expected %v for pair %s, got %v", output[i], x.Source+"-"+x.Destination, out)
		}
	}
}

func TestIbcTxGet(t *testing.T) {
	type data struct {
		Source      string
		Destination string
		Hour        string
		Count       int
	}
	cases := []data{
		{
			Source:      "z1",
			Destination: "z2",
			Hour:        "2020-03-25T18:00:00",
			Count:       17,
		},
	}

	c := NewClient(endpoint)

	for _, d := range cases {
		x := types.IbcStats{
			Source:      d.Source,
			Destination: d.Destination,
			Hour:        types.FromTimestamp(types.Timestamp(d.Hour)),
		}

		exists, err := c.ibcTxExists(context.Background(), x)
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			payload, err := c.ibcTxGet(context.Background(), x)
			if err != nil {
				t.Fatal(err)
			}
			if payload.Count != d.Count {
				t.Fatalf("expected count to be %d, got %d", d.Count, payload.Count)
			}
		}
	}
}

func TestIbcTxAdd(t *testing.T) {
	c := NewClient(endpoint)
	T, err := time.Parse(types.Format, "2020-03-25T14:00:00")
	if err != nil {
		t.Fatal(err)
	}

	data := types.IbcStats{
		Source:      "z1",
		Destination: "z2",
		Count:       5,
		Hour:        T,
	}

	if exists, _ := c.ibcTxExists(context.Background(), data); !exists {
		err = c.ibcTxAdd(context.Background(), data)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestIbcTxIncrement(t *testing.T) {
	c := NewClient(endpoint)
	T, err := time.Parse(types.Format, "2020-03-25T18:00:00")
	if err != nil {
		t.Fatal(err)
	}

	data := types.IbcStats{
		Source:      "z1",
		Destination: "z2",
		Count:       5,
		Hour:        T,
	}

	if exists, _ := c.ibcTxExists(context.Background(), data); exists {
		err = c.ibcTxIncrement(context.Background(), data)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestGetTransfers(t *testing.T) {
	client := NewClient(endpoint)

	txs, err := client.GetUnmatchedIbcTransfers(context.Background(), 100, 0)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(txs)
}

func TestMatchTransfer(t *testing.T) {
	client := NewClient(endpoint)

	data, err := client.FindMatch(context.Background(), types.Transfer{Hash: "1234",
		Matched:   false,
		Quantity:  123,
		Recipient: "A",
		Sender:    "B",
		Timestamp: types.ToTimestamp(time.Now()),
		Token:     "coin",
		Type:      "receive",
		Zone:      "ping-ibc"})
	if err != nil {
		t.Fatal(err)
	}
	print(data.Hash, data.Zone)
}

func TestMatch(t *testing.T) {
	client := NewClient(endpoint)

	err := client.Match(context.Background(), "123")
	if err != nil {
		t.Fatal(err)
	}
}

func TestTotalTxExists(t *testing.T) {
	type data struct {
		Zone string
		Hour string
	}
	cases := []data{
		{
			Zone: "z1",
			Hour: "2020-03-25T18:00:00",
		},
		{
			Zone: "z2",
			Hour: "2020-03-25T18:00:00",
		},
		{
			Hour: "2020-03-25T18:00:00",
			Zone: "not exists",
		},
	}
	output := []bool{true, true, false}

	C := NewClient(endpoint)

	for i, c := range cases {
		x := types.TxStats{
			Zone: c.Zone,
		}
		T, err := time.Parse(types.Format, c.Hour)
		if err != nil {
			t.Fatal(err)
		}
		x.Hour = T

		out, err := C.totalTxExists(context.Background(), x)
		if err != nil {
			t.Fatal(err)
		}
		if out != output[i] {
			t.Fatalf("expected %v for zone %s, got %v", output[i], x.Zone, out)
		}
	}
}

func TestTotalTxGet(t *testing.T) {
	type data struct {
		Zone string
		Hour string
	}
	cases := []data{
		{
			Zone: "z1",
			Hour: "2020-03-25T18:00:00",
		},
		{
			Zone: "z2",
			Hour: "2020-03-25T18:00:00",
		},
		{
			Hour: "2020-03-25T18:00:00",
			Zone: "not exists",
		},
	}

	c := NewClient(endpoint)

	for _, d := range cases {
		x := types.TxStats{
			Zone: d.Zone,
		}
		T, err := time.Parse(types.Format, d.Hour)
		if err != nil {
			t.Fatal(err)
		}
		x.Hour = T
		exists, err := c.totalTxExists(context.Background(), x)
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			payload, err := c.totalTxGet(context.Background(), x)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println(payload)
		}
	}
}

func TestTotalTxAdd(t *testing.T) {
	c := NewClient(endpoint)
	T, err := time.Parse(types.Format, "2020-03-25T14:00:00")
	if err != nil {
		t.Fatal(err)
	}

	data := types.TxStats{
		Zone:  "z1",
		Count: 5,
		Hour:  T,
	}

	if exists, _ := c.totalTxExists(context.Background(), data); !exists {
		err = c.totalTxAdd(context.Background(), data)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestTotalTxIncrement(t *testing.T) {
	c := NewClient(endpoint)
	T, err := time.Parse(types.Format, "2020-03-25T18:00:00")
	if err != nil {
		t.Fatal(err)
	}

	data := types.TxStats{
		Zone:  "z1",
		Count: 5,
		Hour:  T,
	}

	if exists, _ := c.totalTxExists(context.Background(), data); exists {
		err = c.totalTxIncrement(context.Background(), data)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestZones(t *testing.T) {
	q := NewClient(endpoint)
	if zones, err := q.Zones(context.Background()); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println(zones)
	}
}

func TestZoneExists(t *testing.T) {
	cases := map[string]bool{
		"z10018":    true,
		"z279":      true,
		"z0338008":  true,
		"test_zone": false,
	}

	q := NewClient(endpoint)

	for in, out := range cases {
		exists, err := q.ZoneExists(context.Background(), in)
		if err != nil {
			t.Fatal(err)
		}
		if exists != out {
			t.Fatalf("for zone %s expected %v, got %v", in, out, exists)
		}
	}

}

func TestZoneNames(t *testing.T) {
	cases := map[string]string{
		"z10018":    "z1",
		"z279":      "z2",
		"z0338008":  "z3",
		"test_zone": "",
	}

	q := NewClient(endpoint)

	for in, out := range cases {
		name, err := q.ZoneName(context.Background(), in)
		if err != nil {
			t.Fatal(err)
		}
		if name != out {
			t.Fatalf("for zone %s expected %v, got %v", in, out, name)
		}
	}

}
