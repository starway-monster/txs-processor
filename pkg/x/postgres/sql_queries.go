package postgres

import (
	"fmt"
	"time"

	processor "github.com/mapofzones/txs-processor/pkg/types"
)

const addZoneQuery = `insert into zones(name,chain_id,is_enabled) values %s
    on conflict (name) do update
        set is_enabled = %t;`

func addZone(chainID string) string {
	return fmt.Sprintf(addZoneQuery,
		fmt.Sprintf("('%s', '%s', %t)", chainID, chainID, true),
		true,
	)
}

const markBlockQuery = `insert into blocks_log(chain_id,last_processed_block) values %s
    on conflict(chain_id) do update
        set last_processed_block = blocks_log.last_processed_block +1;`

func markBlock(chainID string) string {
	return fmt.Sprintf(markBlockQuery,
		fmt.Sprintf("('%s', %d)", chainID, 1))
}

const addTxStatsQuery = `insert into total_tx_hourly_stats(zone,hour,txs_cnt) values %s
    on conflict(hour,zone) do update
        set txs_cnt = total_tx_hourly_stats.txs_cnt + %d;`

func addTxStats(stats processor.TxStats) string {
	return fmt.Sprintf(addTxStatsQuery,
		fmt.Sprintf("('%s', '%s', %d)", stats.ChainID, stats.Hour.Format(Format), stats.Count),
		stats.Count,
	)
}

const addClientsQuery = `insert into clients(source, client_id,chain_id) values %s
    on conflict(source,client_id) do nothing;`

func addClients(origin string, data map[string]string) string {
	values := ""
	for clientID, chainID := range data {
		values += fmt.Sprintf("('%s', '%s', '%s'),", origin, clientID, chainID)
	}
	values = values[:len(values)-1]

	return fmt.Sprintf(addClientsQuery, values)
}

const addConnectionsQuery = `insert into connections(source, connection_id,client_id) values %s
    on conflict(source,connection_id) do nothing;`

func addConnections(origin string, data map[string]string) string {
	values := ""
	for connectionID, clientID := range data {
		values += fmt.Sprintf("('%s', '%s', '%s'),", origin, connectionID, clientID)
	}
	values = values[:len(values)-1]

	return fmt.Sprintf(addConnectionsQuery, values)
}

const addChannelsQuery = `insert into channels(source, channel_id, connection_id) values %s
    on conflict(source,channel_id) do nothing;`

func addChannels(origin string, data map[string]string) string {
	values := ""
	for channelID, connectionID := range data {
		values += fmt.Sprintf("('%s', '%s', '%s'),", origin, channelID, connectionID)
	}
	values = values[:len(values)-1]

	return fmt.Sprintf(addChannelsQuery, values)
}

const markChannelQuery = `update channels
    set opened = %t
        where source = '%s'
        and channel_id = '%s';`

func markChannel(origin, channelID string, state bool) string {
	return fmt.Sprintf(markChannelQuery,
		state,
		origin,
		channelID)
}

const addIbcStatsQuery = `insert into ibc_tx_hourly_stats(source,zone_src,zone_dest,hour,txs_cnt) values %s
    on conflict(source,zone_src,zone_dest,hour) do update
        set txs_cnt = ibc_tx_hourly_stats.txs_cnt + %d;`

const addImplicitZoneQuery = `insert into zones(name,chain_id,is_enabled) values %s
    on conflict (name) do nothing;`

func addIbcStats(origin string, ibcData map[string]map[string]map[time.Time]int) []string {
	// first we need to get all unique chains
	chainIDs := map[string]struct{}{}
	for source, destinationMap := range ibcData {
		for destination := range destinationMap {
			chainIDs[destination] = struct{}{}
		}
		chainIDs[source] = struct{}{}
	}

	// add zones if we have any
	queries := make([]string, 0, len(chainIDs))
	if len(chainIDs) > 0 {
		zonesValues := ""
		for chainID := range chainIDs {
			zonesValues += fmt.Sprintf("('%s', '%s', %t),", chainID, chainID, false)
		}
		zonesValues = zonesValues[:len(zonesValues)-1]
		queries = append(queries, fmt.Sprintf(addImplicitZoneQuery, zonesValues))
	}

	// process ibc transfers
	if len(ibcData) > 0 {
		for source, destMap := range ibcData {
			for dest, hourMap := range destMap {
				for hour, count := range hourMap {
					queries = append(queries, fmt.Sprintf(addIbcStatsQuery,
						fmt.Sprintf("('%s', '%s', '%s', '%s', %d)", origin, source, dest, hour.Format(Format), count),
						count))
				}
			}
		}
	}
	return queries
}

// queries for querier implementation
const lastProcessedBlockQuery = `select last_processed_block from blocks_log
    where chain_id = '%s';`

const chainIDFromClientIDQuery = `select chain_id from clients
	where client_id = '%s'
		and source = '%s';`

const clientIDFromConnectionIDQuery = `select client_id from connections
	where connection_id = '%s'
		and source = '%s';`

const connectionIDFromChannelIDQuery = `select connection_id from channels
	where channel_id = '%s'
		and source = '%s';`
