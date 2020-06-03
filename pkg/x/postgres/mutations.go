package postgres

import (
	"fmt"
	"time"

	processor "github.com/mapofzones/txs-processor/pkg/types"
)

func addIbcStats(origin string, ibcData map[string]map[string]map[time.Time]int) []string {
	// add zones if we have any
	queries := make([]string, 0, 32)

	// process ibc transfers
	for source, destMap := range ibcData {
		for dest, hourMap := range destMap {
			for hour, count := range hourMap {
				queries = append(queries, fmt.Sprintf(addIbcStatsQuery,
					fmt.Sprintf("('%s', '%s', '%s', '%s', %d, %d)", origin, source, dest, hour.Format(Format), count, 1),
					count))
			}
		}
	}
	return queries
}

func markChannel(origin, channelID string, state bool) string {
	return fmt.Sprintf(markChannelQuery,
		state,
		origin,
		channelID)
}

func addChannels(origin string, data map[string]string) string {
	values := ""
	for channelID, connectionID := range data {
		values += fmt.Sprintf("('%s', '%s', '%s',%t),", origin, channelID, connectionID, false)
	}
	values = values[:len(values)-1]

	return fmt.Sprintf(addChannelsQuery, values)
}

func addConnections(origin string, data map[string]string) string {
	values := ""
	for connectionID, clientID := range data {
		values += fmt.Sprintf("('%s', '%s', '%s'),", origin, connectionID, clientID)
	}
	values = values[:len(values)-1]

	return fmt.Sprintf(addConnectionsQuery, values)
}

func addClients(origin string, data map[string]string) []string {
	zonesValues := ""
	for _, chainID := range data {
		zonesValues += fmt.Sprintf("('%s', '%s', %t, %t),", chainID, chainID, false, false)
	}
	zonesValues = zonesValues[:len(zonesValues)-1]
	impicitZones := fmt.Sprintf(addImplicitZoneQuery, zonesValues)

	values := ""
	for clientID, chainID := range data {
		values += fmt.Sprintf("('%s', '%s', '%s'),", origin, clientID, chainID)
	}
	values = values[:len(values)-1]

	return []string{impicitZones, fmt.Sprintf(addClientsQuery, values)}
}

func addTxStats(stats processor.TxStats) string {
	return fmt.Sprintf(addTxStatsQuery,
		fmt.Sprintf("('%s', '%s', %d, %d, %d)", stats.ChainID, stats.Hour.Format(Format), stats.Count, stats.TxWithIBCTransfer, 1),
		stats.Count,
		stats.TxWithIBCTransfer,
	)
}

func markBlock(chainID string) string {
	t := time.Now().Format(Format)
	return fmt.Sprintf(markBlockQuery,
		fmt.Sprintf("('%s', %d, '%s')", chainID, 1, t), t)
}

func addZone(chainID string) string {
	return fmt.Sprintf(addZoneQuery,
		fmt.Sprintf("('%s', '%s', %t, %t)", chainID, chainID, true, false),
		true,
	)
}
