package builder

import (
	"fmt"

	"github.com/mapofzones/txs-processor/types"
)

const (
	header = "mutation Block {\n"
	footer = "\n}"
)

// here just for convenience
type Builder interface {
	AddZone(string)
	PushBlock(types.Block)
	CreateTxStats(types.TxStats)
	UpdateTxStats(types.TxStats)
	PushTransfer(types.Transfer)
	CreateIbcStats(types.IbcStats)
	UpdateIbcStats(types.IbcStats)
	Mutation() string
}

type MutationBuilder struct {
	actions []string
}

func (m *MutationBuilder) AddZone(chainID string) {
	m.actions = append(m.actions, fmt.Sprintf(addZone, chainID, chainID))
}

func (m *MutationBuilder) PushBlock(b types.Block) {
	m.actions = append(m.actions, fmt.Sprintf(pushBlock, b.Height, b.ChainID))
}

func (m *MutationBuilder) CreateTxStats(s types.TxStats) {
	m.actions = append(m.actions, fmt.Sprintf(insertTxStats, s.Hour.Format(types.Format), s.Count, s.Zone))
}

func (m *MutationBuilder) UpdateTxStats(s types.TxStats) {
	m.actions = append(m.actions, fmt.Sprintf(updateTxStats, s.Zone, s.Hour.Format(types.Format), s.Count))
}

func (m *MutationBuilder) PushTransfer(t types.Transfer) {
	m.actions = append(m.actions, fmt.Sprintf(pushTransfer, t.Hash, t.Quantity, t.Recipient, t.Sender, t.Timestamp, t.Token, t.Type, t.Zone))
}

func (m *MutationBuilder) CreateIbcStats(s types.IbcStats) {
	m.actions = append(m.actions, fmt.Sprintf(insertIbcStats, s.Hour.Format(types.Format), s.Count, s.Destination, s.Source))
}

func (m *MutationBuilder) UpdateIbcStats(s types.IbcStats) {
	m.actions = append(m.actions, fmt.Sprintf(updateIbcStats, s.Hour.Format(types.Format), s.Destination, s.Source, s.Count))
}

func (m MutationBuilder) Mutation() string {
	mutation := header
	for _, v := range m.actions {
		mutation += v + "\n"
	}
	return mutation + footer
}
