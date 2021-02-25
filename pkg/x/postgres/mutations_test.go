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
