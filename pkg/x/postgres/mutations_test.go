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
        {"single_pair", args{map[string]string{"clientId1":"chainId1"}}, "insert into zones(name, chain_id, is_enabled, is_caught_up) values ('chainId1', 'chainId1', false, false)\n    on conflict (chain_id) do nothing;"},
        {"two_pair", args{map[string]string{"clientId1":"chainId1","clientId2":"chainId2"}}, "insert into zones(name, chain_id, is_enabled, is_caught_up) values ('chainId1', 'chainId1', false, false),('chainId2', 'chainId2', false, false)\n    on conflict (chain_id) do nothing;"},
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
            actual := markBlockConstruct(tt.args.chainID, tt.args.time)//"2006-01-02T15:04:05")
            assert.Equal(t, tt.expected, actual)
        })
    }
}