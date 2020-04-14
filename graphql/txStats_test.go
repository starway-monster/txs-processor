package graphql

import (
	"context"
	"fmt"
	types "test/types"
	"testing"
	"time"
)

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
