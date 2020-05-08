package builder

const addZone = `insert_zones(objects: {chain_id: "%s", name: "%s", is_enabled: true}, on_conflict: {constraint: zones_chain_id_key, update_columns: chain_id}) {
    affected_rows
  }`

const insertProcessedBlocksEntry = `insert_blocks_log(objects: {chain_id: "%s", last_processed_block: %d}, on_conflict: {update_columns: chain_id, constraint: blocks_log_hub_pkey}) {
    affected_rows
  }`

const markBlock = `update_blocks_log(where: {chain_id: {_eq: "%s"}}, _inc: {last_processed_block: 1}) {
    affected_rows
  }`

const insertTxStats = `insert_total_tx_hourly_stats(objects: {hour: "%s", txs_cnt: %d, zone: "%s"}) {
    affected_rows
  }`

const updateTxStats = `update_total_tx_hourly_stats(where: {zone: {_eq: "%s"}, hour: {_eq: "%s"}}, _inc: {txs_cnt: %d}) {
    affected_rows
  }`

const insertIbcStats = `insert_ibc_tx_hourly_stats(objects: {hour: "%s", txs_cnt: %d, zone_dest: "%s", zone_src: "%s"}) {
    affected_rows
  }`

const updateIbcStats = `update_ibc_tx_hourly_stats(where: {hour: {_eq: "%s"}, zone_dest: {_eq: "%s"}, zone_src: {_eq: "%s"}}, _inc: {txs_cnt: %d}) {
    affected_rows
  }`

const insertClient = `insert_clients(objects: {chain_id: "%s", client_id: "%s", source: "%s"}) {
    affected_rows
  }`

const insertConnection = `insert_connections(objects: {client_id: "%s", connection_id: "%s", source: "%s"}) {
    affected_rows
  }`

const insertChannel = `insert_channels(objects: {channel_id: "%s", connection_id: "%s", source: "%s"}, on_conflict: {constraint: channels_pkey, update_columns: connection_id}) {
    affected_rows
  }`

const setChannelStatus = `update_channels(where: {channel_id: {_eq: "%s"}, source: {_eq: "%s"}}, _set: {opened: %t}) {
    affected_rows
  }`
