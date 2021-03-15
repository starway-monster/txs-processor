package postgres

import (
    "github.com/stretchr/testify/assert"
    "testing"
)

func Test_addZone(t *testing.T) {
    type args struct {
        chainID string
    }
    tests := []struct {
        name        string
        args        args
        expected    string
    }{
        {"empty_args", args{}, "insert into zones(name, chain_id, is_enabled, is_caught_up) values ('', '', true, false)\n    on conflict (chain_id) do update\n        set is_enabled = true;"},
        {"first_args", args{"myChain1"}, "insert into zones(name, chain_id, is_enabled, is_caught_up) values ('myChain1', 'myChain1', true, false)\n    on conflict (chain_id) do update\n        set is_enabled = true;"},
        {"second_args", args{"myChain2"}, "insert into zones(name, chain_id, is_enabled, is_caught_up) values ('myChain2', 'myChain2', true, false)\n    on conflict (chain_id) do update\n        set is_enabled = true;"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := addZone(tt.args.chainID)
            assert.Equal(t, tt.expected, actual)
        })
    }
}

func Test_addImplicitZones(t *testing.T) {
    type args struct {
        clients map[string]string
    }
    tests := []struct {
        name string
        args args
        expected string
    }{
        {"empty_args", args{}, "insert into zones(name, chain_id, is_enabled, is_caught_up) values \n    on conflict (chain_id) do nothing;"},
        {"first_pair", args{map[string]string{"clientId1":"chainId1"}}, "insert into zones(name, chain_id, is_enabled, is_caught_up) values ('chainId1', 'chainId1', false, false)\n    on conflict (chain_id) do nothing;"},
        {"second_pair", args{map[string]string{"clientId2":"chainId2"}}, "insert into zones(name, chain_id, is_enabled, is_caught_up) values ('chainId2', 'chainId2', false, false)\n    on conflict (chain_id) do nothing;"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := addImplicitZones(tt.args.clients)
            assert.Equal(t, tt.expected, actual)
        })
    }
}

func Test_markBlockConstruct(t *testing.T) {
    type args struct {
        chainID string
        time    string
    }
    tests := []struct {
        name string
        args args
        expected string
    }{
        {"empty_args", args{}, "insert into blocks_log(zone, last_processed_block, last_updated_at) values ('', 1, '')\n    on conflict (zone) do update\n        set last_processed_block = blocks_log.last_processed_block + 1,\n            last_updated_at = '';"},
        {"first_args", args{"chainID1", "2006-01-02T15:04:05"}, "insert into blocks_log(zone, last_processed_block, last_updated_at) values ('chainID1', 1, '2006-01-02T15:04:05')\n    on conflict (zone) do update\n        set last_processed_block = blocks_log.last_processed_block + 1,\n            last_updated_at = '2006-01-02T15:04:05';"},
        {"second_args", args{"chainID2", "2016-12-02T06:14:55"}, "insert into blocks_log(zone, last_processed_block, last_updated_at) values ('chainID2', 1, '2016-12-02T06:14:55')\n    on conflict (zone) do update\n        set last_processed_block = blocks_log.last_processed_block + 1,\n            last_updated_at = '2016-12-02T06:14:55';"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := markBlockConstruct(tt.args.chainID, tt.args.time)
            assert.Equal(t, tt.expected, actual)
        })
    }
}

func Test_addClients(t *testing.T) {
    type args struct {
        origin  string
        clients map[string]string
    }
    tests := []struct {
        name string
        args args
        expected string
    }{
        {
            "empty_args",
            args{},
            "insert into ibc_clients(zone, client_id, chain_id) values \n    on conflict (zone, client_id) do nothing;",
        },
        {
            "first_args",
            args{"myOrigin1", map[string]string{"clientID1":"chainID1"}},
            "insert into ibc_clients(zone, client_id, chain_id) values ('myOrigin1', 'clientID1', 'chainID1')\n    on conflict (zone, client_id) do nothing;",
        },
        {
            "second_args",
            args{"myOrigin2", map[string]string{"clientID2":"chainID2"}},
            "insert into ibc_clients(zone, client_id, chain_id) values ('myOrigin2', 'clientID2', 'chainID2')\n    on conflict (zone, client_id) do nothing;",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := addClients(tt.args.origin, tt.args.clients)
            assert.Equal(t, tt.expected, actual)
        })
    }
}

func Test_addConnections(t *testing.T) {
    type args struct {
        origin string
        data   map[string]string
    }
    tests := []struct {
        name string
        args args
        expected string
    }{
        {
            "empty_args",
            args{},
            "insert into ibc_connections(zone, connection_id, client_id) values \n    on conflict (zone, connection_id) do nothing;",
        },
        {
            "first_args",
            args{"origin1", map[string]string{"connectionID1": "clientID1"}},
            "insert into ibc_connections(zone, connection_id, client_id) values ('origin1', 'connectionID1', 'clientID1')\n    on conflict (zone, connection_id) do nothing;",
        },
        {
            "second_args",
            args{"origin2", map[string]string{"connectionID2": "clientID2"}},
            "insert into ibc_connections(zone, connection_id, client_id) values ('origin2', 'connectionID2', 'clientID2')\n    on conflict (zone, connection_id) do nothing;",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := addConnections(tt.args.origin, tt.args.data)
            assert.Equal(t, tt.expected, actual)
        })
    }
}

func Test_addChannels(t *testing.T) {
    type args struct {
        origin string
        data   map[string]string
    }
    tests := []struct {
        name string
        args args
        expected string
    }{
        {
            "empty_args",
            args{},
            "insert into ibc_channels(zone, channel_id, connection_id, is_opened) values \n    on conflict(zone, channel_id) do nothing;",
        },
        {
            "first_args",
            args{"origin1", map[string]string{"channelID1": "connectionID1"}},
            "insert into ibc_channels(zone, channel_id, connection_id, is_opened) values ('origin1', 'channelID1', 'connectionID1',false)\n    on conflict(zone, channel_id) do nothing;",
        },
        {
            "second_args",
            args{"origin2", map[string]string{"channelID2": "connectionID2"}},
            "insert into ibc_channels(zone, channel_id, connection_id, is_opened) values ('origin2', 'channelID2', 'connectionID2',false)\n    on conflict(zone, channel_id) do nothing;",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := addChannels(tt.args.origin, tt.args.data)
            assert.Equal(t, tt.expected, actual)
        })
    }
}

func Test_markChannel(t *testing.T) {
    type args struct {
        origin    string
        channelID string
        state     bool
    }
    tests := []struct {
        name string
        args args
        expected string
    }{
        {
            "empty_args",
            args{},
            "update ibc_channels\n    set is_opened = false\n        where zone = ''\n        and channel_id = '';",
        },
        {
            "first_args",
            args{"origin1", "myChannelID1", true},
            "update ibc_channels\n    set is_opened = true\n        where zone = 'origin1'\n        and channel_id = 'myChannelID1';",
        },
        {
           "second_args",
           args{"origin2", "myChannelID2", false},
            "update ibc_channels\n    set is_opened = false\n        where zone = 'origin2'\n        and channel_id = 'myChannelID2';",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := markChannel(tt.args.origin, tt.args.channelID, tt.args.state)
            assert.Equal(t, tt.expected, actual)
        })
    }
}
