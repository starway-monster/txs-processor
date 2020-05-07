package graphql

import "github.com/machinebox/graphql"

func toRequest(query string) *graphql.Request {
	return graphql.NewRequest(header + query + footer)
}

const header = "query MyQuery{\n"

const footer = "\n}"

const lastProcessedBlock = `blocks_log_hub(where: {chain_id: {_eq: "%s"}}) {
    last_processed_block
  }`

const ibcStatsExist = `ibc_tx_hourly_stats_hub(where: {zone_src: {_eq: "%s"}, zone_dest: {_eq: "%s"}, hour: {_eq: "%s"}}) {
    txs_cnt
  }`

const connectionIDFromChannelID = `channels(where: {source: {_eq: "%s"}, channel_id: {_eq: "%s"}}) {
    connection_id
  }`

const clientIDFromConnectionID = `connections(where: {connection_id: {_eq: "%s"}, source: {_eq: "%s"}}) {
    client_id
  }`

const chainIDFromClientID = `clients(where: {source: {_eq: "%s"}, client_id: {_eq: "%s"}}) {
    chain_id
  }`
