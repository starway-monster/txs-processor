package postgres

const addZoneQuery = `insert into zones(name,chain_id,is_enabled) values %s
    on conflict (chain_id) do update
        set is_enabled = %t;`

const markBlockQuery = `insert into blocks_log(zone,last_processed_block) values %s
    on conflict(zone) do update
        set last_processed_block = blocks_log.last_processed_block +1;`

const addTxStatsQuery = `insert into total_tx_hourly_stats(zone,hour,txs_cnt,period) values %s
    on conflict(hour,zone, period) do update
        set txs_cnt = total_tx_hourly_stats.txs_cnt + %d;`

const addClientsQuery = `insert into ibc_clients(zone,client_id,chain_id) values %s
    on conflict(zone,client_id) do nothing;`

const addConnectionsQuery = `insert into ibc_connections(zone, connection_id,client_id) values %s
    on conflict(zone,connection_id) do nothing;`

const addChannelsQuery = `insert into ibc_channels(zone, channel_id, connection_id) values %s
    on conflict(zone,channel_id) do nothing;`

const markChannelQuery = `update channels
    set opened = %t
        where zone = '%s'
        and channel_id = '%s';`

const addIbcStatsQuery = `insert into ibc_transfer_hourly_stats(zone,zone_src,zone_dest,hour,txs_cnt) values %s
    on conflict(zone,zone_src,zone_dest,hour) do update
        set txs_cnt = ibc_transfer_hourly_stats.txs_cnt + %d;`

const addImplicitZoneQuery = `insert into zones(name,chain_id,is_enabled) values %s
    on conflict (chain_id) do nothing;`

const lastProcessedBlockQuery = `select last_processed_block from blocks_log
    where zone = '%s';`

const chainIDFromClientIDQuery = `select chain_id from clients
	where client_id = '%s'
		and zone = '%s';`

const clientIDFromConnectionIDQuery = `select client_id from connections
	where connection_id = '%s'
		and zone = '%s';`

const connectionIDFromChannelIDQuery = `select connection_id from channels
	where channel_id = '%s'
		and zone = '%s';`
