package postgres

import (
    processor "github.com/mapofzones/txs-processor/pkg/types"
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
            actual := markBlockConstruct(tt.args.chainID, tt.args.time)
            assert.Equal(t, tt.expected, actual)
        })
    }
}

func Test_addTxStats(t *testing.T) {
    type args struct {
        stats processor.TxStats
    }
    tests := []struct {
        name string
        args args
        expected string
    }{
        {
            "empty_args",
            args{},
            "insert into total_tx_hourly_stats(zone, hour, txs_cnt, txs_w_ibc_xfer_cnt, period, txs_w_ibc_xfer_fail_cnt, total_coin_turnover_amount) values ('', '0001-01-01T00:00:00', 0, 0, 1, 0, 0)\n    on conflict (hour, zone, period) do update\n        set txs_cnt = total_tx_hourly_stats.txs_cnt + 0,\n\t\t\ttxs_w_ibc_xfer_cnt = total_tx_hourly_stats.txs_w_ibc_xfer_cnt + 0,\n\t\t\ttxs_w_ibc_xfer_fail_cnt = total_tx_hourly_stats.txs_w_ibc_xfer_fail_cnt + 0,\n            total_coin_turnover_amount = total_tx_hourly_stats.total_coin_turnover_amount + 0;",
        },
        {
            "first_args",
            args{processor.TxStats{Count: 1, TxWithIBCTransfer: 2, TxWithIBCTransferFail: 3, TurnoverAmount: 11111122222333333}},
            "insert into total_tx_hourly_stats(zone, hour, txs_cnt, txs_w_ibc_xfer_cnt, period, txs_w_ibc_xfer_fail_cnt, total_coin_turnover_amount) values ('', '0001-01-01T00:00:00', 1, 2, 1, 3, 11111122222333333)\n    on conflict (hour, zone, period) do update\n        set txs_cnt = total_tx_hourly_stats.txs_cnt + 1,\n\t\t\ttxs_w_ibc_xfer_cnt = total_tx_hourly_stats.txs_w_ibc_xfer_cnt + 2,\n\t\t\ttxs_w_ibc_xfer_fail_cnt = total_tx_hourly_stats.txs_w_ibc_xfer_fail_cnt + 3,\n            total_coin_turnover_amount = total_tx_hourly_stats.total_coin_turnover_amount + 11111122222333333;",
        },
        {
            "second_args",
            args{processor.TxStats{Count: 348, TxWithIBCTransfer: 3952, TxWithIBCTransferFail: 842, TurnoverAmount: 877581450957345}},
            "insert into total_tx_hourly_stats(zone, hour, txs_cnt, txs_w_ibc_xfer_cnt, period, txs_w_ibc_xfer_fail_cnt, total_coin_turnover_amount) values ('', '0001-01-01T00:00:00', 348, 3952, 1, 842, 877581450957345)\n    on conflict (hour, zone, period) do update\n        set txs_cnt = total_tx_hourly_stats.txs_cnt + 348,\n\t\t\ttxs_w_ibc_xfer_cnt = total_tx_hourly_stats.txs_w_ibc_xfer_cnt + 3952,\n\t\t\ttxs_w_ibc_xfer_fail_cnt = total_tx_hourly_stats.txs_w_ibc_xfer_fail_cnt + 842,\n            total_coin_turnover_amount = total_tx_hourly_stats.total_coin_turnover_amount + 877581450957345;",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := addTxStats(tt.args.stats)
            assert.Equal(t, tt.expected, actual)
        })
    }
}