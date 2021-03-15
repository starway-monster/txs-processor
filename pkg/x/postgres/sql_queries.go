package postgres

// queries that write to db

const addZoneQuery = `insert into zones(name, chain_id, is_enabled, is_caught_up) values %s
    on conflict (chain_id) do update
        set is_enabled = %t;`

const addImplicitZoneQuery = `insert into zones(name, chain_id, is_enabled, is_caught_up) values %s
    on conflict (chain_id) do nothing;`

const markBlockQuery = `insert into blocks_log(zone, last_processed_block, last_updated_at) values %s
    on conflict (zone) do update
        set last_processed_block = blocks_log.last_processed_block + 1,
            last_updated_at = '%s';`

const addClientsQuery = `insert into ibc_clients(zone, client_id, chain_id) values %s
    on conflict (zone, client_id) do nothing;`

const addConnectionsQuery = `insert into ibc_connections(zone, connection_id, client_id) values %s
    on conflict (zone, connection_id) do nothing;`

const addChannelsQuery = `insert into ibc_channels(zone, channel_id, connection_id, is_opened) values %s
    on conflict(zone, channel_id) do nothing;`

const markChannelQuery = `update ibc_channels
    set is_opened = %t
        where zone = '%s'
        and channel_id = '%s';`

const lastProcessedBlockQuery = `select last_processed_block from blocks_log
    where zone = '%s';`

const chainIDFromClientIDQuery = `select chain_id from ibc_clients
	where client_id = '%s'
		and zone = '%s';`

const clientIDFromConnectionIDQuery = `select client_id from ibc_connections
	where connection_id = '%s'
		and zone = '%s';`

const connectionIDFromChannelIDQuery = `select connection_id from ibc_channels
	where channel_id = '%s'
		and zone = '%s';`
