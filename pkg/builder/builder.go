package builder

import (
	"fmt"

	"github.com/mapofzones/txs-processor/pkg/types"
)

const (
	header = "mutation Block {\n"
	footer = "\n}"
)

var _ Builder = &MutationBuilder{}

// here just for convenience
type Builder interface {
	AddZone(string)
	MarkBlock(types.Block)
	CreateTxStats(types.TxStats)
	UpdateTxStats(types.TxStats)
	CreateIbcStats(types.IbcStats)
	UpdateIbcStats(types.IbcStats)

	Mutation() string
}

type MutationBuilder struct {
	actions []string
}

func (m *MutationBuilder) AddZone(chainID string) {
	m.actions = append(m.actions, fmt.Sprintf(addZone, chainID, chainID))
	m.actions = append(m.actions, fmt.Sprintf(insertProcessedBlocksEntry, chainID, 0))
}

// marks block as processed
func (m *MutationBuilder) MarkBlock(b types.Block) {
	m.actions = append(m.actions, fmt.Sprintf(markBlock, b.ChainID))
}

func (m *MutationBuilder) CreateTxStats(s types.TxStats) {
	m.actions = append(m.actions, fmt.Sprintf(insertTxStats, s.Hour.Format(types.Format), s.Count, s.Zone))
}

func (m *MutationBuilder) UpdateTxStats(s types.TxStats) {
	m.actions = append(m.actions, fmt.Sprintf(updateTxStats, s.Zone, s.Hour.Format(types.Format), s.Count))
}

func (m *MutationBuilder) CreateIbcStats(s types.IbcStats) {
	m.actions = append(m.actions, fmt.Sprintf(insertIbcStats, s.Hour.Format(types.Format), s.Count, s.Destination, s.Source))
}

func (m *MutationBuilder) UpdateIbcStats(s types.IbcStats) {
	m.actions = append(m.actions, fmt.Sprintf(updateIbcStats, s.Hour.Format(types.Format), s.Destination, s.Source, s.Count))
}

func (m *MutationBuilder) InsertClient(source, clientID, chainID string) {
	m.actions = append(m.actions, fmt.Sprintf(insertClient, chainID, clientID, source))
}

func (m *MutationBuilder) InsertConnection(source, connectionID, clientID string) {
	m.actions = append(m.actions, fmt.Sprintf(insertConnection, clientID, connectionID, source))
}

func (m *MutationBuilder) InsertChannel(source, channelID, connectionID string) {
	m.actions = append(m.actions, fmt.Sprintf(insertChannel, channelID, connectionID, source))
}

func (m *MutationBuilder) MarkChannelOpened(source, channelID string) {
	m.actions = append(m.actions, fmt.Sprintf(setChannelStatus, channelID, source, true))
}

func (m *MutationBuilder) MarkChannelClosed(source, channelID string) {
	m.actions = append(m.actions, fmt.Sprintf(setChannelStatus, channelID, source, false))
}

func (m MutationBuilder) Mutation() string {
	mutation := header
	for _, v := range m.actions {
		mutation += v + "\n"
	}
	return mutation + footer
}
