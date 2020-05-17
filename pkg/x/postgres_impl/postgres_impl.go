package processor

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/mapofzones/txs-processor/pkg/types"
)

type PostgresProcessor struct {
	// not safe for concurrent use
	conn *sql.DB
	// list of zones encountered in block
	chainID string

	// should we mark this block as processed
	mark bool

	// block tx data
	txStats types.TxStats

	// ibc protocol data
	// client-id -> chain-id map
	clients map[string]string
	// connection-id -> client-id map
	connections map[string]string
	// channel-id -> connection-id map
	channels map[string]string

	// channel-id -> opened(true) || closed
	channelStates map[string]bool
	// ibc message data for block
	// zone_src -> zone_dest -> hour -> txs_cnt
	ibcStats map[string]map[string]map[time.Time]int
}

func NewPostgresProcessor(addr string) (*PostgresProcessor, error) {
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, err
	}

	return &PostgresProcessor{
		conn:          db,
		ibcStats:      map[string]map[string]map[time.Time]int{},
		txStats:       types.TxStats{},
		chainID:       "",
		clients:       map[string]string{},
		connections:   map[string]string{},
		channels:      map[string]string{},
		channelStates: map[string]bool{},
	}, nil
}

func (p *PostgresProcessor) AddZone(chainID string) {
	p.chainID = chainID
}

func (p *PostgresProcessor) MarkBlock() {
	p.mark = true
}

func (p *PostgresProcessor) AddTxStats(stats types.TxStats) {
	p.txStats = stats
}

func (p *PostgresProcessor) AddClient(clientID string, chainID string) {
	p.clients[clientID] = chainID
}

func (p *PostgresProcessor) AddConnection(connectionID string, clientID string) {
	p.connections[connectionID] = clientID
}

func (p *PostgresProcessor) AddChannel(channelID string, connectionID string) {
	p.channels[channelID] = connectionID
}

func (p *PostgresProcessor) MarkChannelOpened(channelID string) {
	p.channelStates[channelID] = true
}

func (p *PostgresProcessor) MarkChannelClosed(channelID string) {
	p.channelStates[channelID] = false
}

func (p *PostgresProcessor) AddIbcStats(stats types.IbcStats) {
	if p.ibcStats[stats.Source] == nil {
		p.ibcStats[stats.Source] = map[string]map[time.Time]int{}
		p.ibcStats[stats.Source][stats.Destination] = map[time.Time]int{}
		p.ibcStats[stats.Source][stats.Destination][stats.Hour] = stats.Count
		return
	}

	if p.ibcStats[stats.Source][stats.Destination] == nil {
		p.ibcStats[stats.Source][stats.Destination] = map[time.Time]int{}
		p.ibcStats[stats.Source][stats.Destination][stats.Hour] = stats.Count
		return
	}

	p.ibcStats[stats.Source][stats.Destination][stats.Hour] += stats.Count
}

func (p *PostgresProcessor) Commit(ctx context.Context) error {
	conn, err := p.conn.Conn(ctx)
	if err != nil {
		return err
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// add if this is the first block from it
	_, err = tx.Exec(addZone(p.chainID))
	if err != nil {
		tx.Rollback()
		return err
	}

	// mark block as processed
	if p.mark {
		_, err = tx.Exec(markBlock(p.chainID))
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// update TxStats
	if p.txStats.Count > 0 {
		_, err = tx.Exec(addTxStats(p.txStats))
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// insert ibc clients
	if len(p.clients) > 0 {
		_, err = tx.Exec(addClients(p.chainID, p.clients))
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	// insert ibc connections
	if len(p.connections) > 0 {
		_, err = tx.Exec(addConnections(p.chainID, p.connections))
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	// insert ibc channels
	if len(p.channels) > 0 {
		_, err = tx.Exec(addChannels(p.chainID, p.channels))
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// update channelStates
	for channel, state := range p.channelStates {
		_, err = tx.Exec(markChannel(p.chainID, channel, state))
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// update ibc stats and add untraced zones
	for _, query := range addIbcStats(p.chainID, p.ibcStats) {
		_, err = tx.Exec(query)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return conn.Close()
}

// reset the state of our processor
func (p *PostgresProcessor) Reset() {
	p.ibcStats = map[string]map[string]map[time.Time]int{}

	p.txStats = types.TxStats{}

	p.chainID = ""
	p.mark = false

	p.connections = map[string]string{}
	p.connections = map[string]string{}
	p.channels = map[string]string{}

	p.channelStates = map[string]bool{}
}

func (p *PostgresProcessor) LastProcessedBlock(chainID string) (int64, error) {
	res, err := p.conn.Query(fmt.Sprintf(lastProcessedBlockQuery, chainID))
	if err != nil {
		return -1, err
	}
	if res.Next() {
		block := new(int64)
		err = res.Scan(&block)
		if err != nil {
			return -1, err
		}
		return *block, nil
	}
	return 0, nil
}

func (p *PostgresProcessor) ChainIDFromClientID(clientID string) (string, error) {
	res, err := p.conn.Query(fmt.Sprintf(chainIDFromClientIDQuery, clientID, p.chainID))
	if err != nil {
		return "", err
	}

	if res.Next() {
		chainID := ""
		err = res.Scan(&chainID)
		if err != nil {
			return "", err
		}
		return chainID, nil
	}
	return "", nil
}

func (p *PostgresProcessor) ChainIDFromConnectionID(connectionID string) (string, error) {
	res, err := p.conn.Query(fmt.Sprintf(clientIDFromConnectionIDQuery, connectionID, p.chainID))
	if err != nil {
		return "", err
	}

	if res.Next() {
		clientID := ""
		err = res.Scan(&clientID)
		if err != nil {
			return "", err
		}
		return p.ChainIDFromClientID(clientID)
	}
	return "", nil
}

func (p *PostgresProcessor) ChainIDFromChannelID(channelID string) (string, error) {
	res, err := p.conn.Query(fmt.Sprintf(connectionIDFromChannelIDQuery, channelID, p.chainID))
	if err != nil {
		return "", err
	}

	if res.Next() {
		connectionID := ""
		err = res.Scan(&connectionID)
		if err != nil {
			return "", err
		}
		return p.ChainIDFromConnectionID(connectionID)
	}
	return "", nil
}
