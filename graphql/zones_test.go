package graphql

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"
)

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
			t.Fatalf("for zone %s exepcted %v, got %v", in, out, exists)
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
			t.Fatalf("for zone %s exepcted %v, got %v", in, out, name)
		}
	}

}

func TestAddZone(t *testing.T) {
	c := NewClient(endpoint)

	err := c.AddZone(context.Background(), base64.StdEncoding.EncodeToString(randBytes()), true)
	if err != nil {
		t.Fatal(err)
	}

}
