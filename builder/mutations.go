package builder

const addZone = `insert_zones(objects: {chain_id: "%s", name: "%s", is_enabled: true}) {
    affected_rows
  }`

const pushBlock = `insert_blocks_log(objects: {height: %d, zone: "%s"}) {
    affected_rows
  }`

const insertTxStats = `insert_total_tx_hourly_stats(objects: {hour: "%s", txs_cnt: %d, zone: "%s"}) {
    affected_rows
  }`

const updateTxStats = `update_total_tx_hourly_stats(where: {zone: {_eq: "%s"}, hour: {_eq: "%s"}}, _inc: {txs_cnt: %d}) {
    affected_rows
  }`

// TODO: remove "matched: false" after new logic, which deletes txs when they are matched during block processing
const pushTransfer = `insert_ibc_tx_transfer_log(objects: {hash: "%s", matched_to_tx: false, quantity: %d, recipient: "%s", sender: "%s", timestamp: "%s", token: "%s", type: "%s", zone: "%s"}) {
    affected_rows
  }`

const insertIbcStats = `insert_ibc_tx_hourly_stats(objects: {hour: "%s", txs_cnt: %d, zone_dest: "%s", zone_src: "%s"}) {
    affected_rows
  }`

const updateIbcStats = `update_ibc_tx_hourly_stats(where: {hour: {_eq: "%s"}, zone_dest: {_eq: "%s"}, zone_src: {_eq: "%s"}}, _inc: {txs_cnt: %d}) {
    affected_rows
  }`
