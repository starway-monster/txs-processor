package graphql

import (
	"context"
	types "test/types"
	"testing"
	"time"
)

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
			Hour:        fromTimestamp(timestamp(c.Hour)),
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
			Hour:        fromTimestamp(timestamp(d.Hour)),
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
