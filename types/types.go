package processor

import (
	"time"

	watcher "github.com/attractor-spectrum/cosmos-watcher"
)

// Format used by postgres to store timestamps
const Format = "2006-01-02T15:04:05"

// Tx alias
type Tx = watcher.Tx

type Type = watcher.Type

const (
	Transfer   = watcher.Transfer
	IbcSend    = watcher.IbcSend
	IbcRecieve = watcher.IbcRecieve
)

// Ibc returns true if transaction is inter-blockchain
func Ibc(t Tx) bool {
	return t.Type == watcher.IbcRecieve || t.Type == watcher.IbcSend
}

// Txs is used to couple tx slice with errors, since they are closely related
// it allows us to get rid of extra channel
type Txs struct {
	Txs []Tx
	Err error
}

// SplitIBC splits tx slice into two, one which has ibc txs, and other everything else
func (t Txs) SplitIBC() (local Txs, ibc Txs) {
	local = Txs{}
	ibc = Txs{}

	for _, tx := range t.Txs {
		if tx.Type == watcher.IbcSend || tx.Type == watcher.IbcRecieve {
			ibc.Txs = append(ibc.Txs, tx)
		} else {
			local.Txs = append(local.Txs, tx)
		}
	}
	return local, ibc
}

// ToStats returns TxStats slice, where each slice time is already properly formated
func (t Txs) ToStats() []TxStats {
	stats := []TxStats{}

	// map[zone][hour]count
	raw := map[string]map[time.Time]int{}

	for _, tx := range t.Txs {
		if raw[tx.Network] == nil {
			raw[tx.Network] = make(map[time.Time]int)
		}
		raw[tx.Network][tx.T.Truncate(time.Hour)]++
	}

	for zone, hour := range raw {
		for hour := range hour {
			stats = append(stats, TxStats{Zone: zone, Hour: hour, Count: raw[zone][hour]})
		}
	}

	return stats
}

// TxStats structure is used to see how many txs were send during each hour
type TxStats struct {
	Zone  string
	Hour  time.Time //must have 0 minutes, seconods and micro/nano seconds
	Count int
}

// IbcStats represents statistics that we need to write to db
type IbcStats struct {
	Source      string
	Destination string
	Hour        time.Time
	Count       int
}

// IbcData is used to organize ibc tx data during each hour
type IbcData map[string]map[string]map[time.Time]int

// Append truncates timestamps and puts data into ibc data structure
func (m IbcData) Append(source, destination string, t time.Time) {
	t = t.Truncate(time.Hour)
	if m == nil {
		m = make(IbcData)
	}

	if m[source] == nil {
		m[source] = make(map[string]map[time.Time]int)
	}

	if m[source][destination] == nil {
		m[source][destination] = make(map[time.Time]int)
	}

	m[source][destination][t]++
	return
}

// ToIbcStats returns slice of ibc stats formed from ibcData maps
func (m IbcData) ToIbcStats() []IbcStats {
	stats := []IbcStats{}
	for source := range m {
		for destination := range m[source] {
			for hour := range m[source][destination] {
				stats = append(stats, IbcStats{
					Source:      source,
					Destination: destination,
					Hour:        hour,
					Count:       m[source][destination][hour],
				})
			}
		}
	}
	return stats
}
