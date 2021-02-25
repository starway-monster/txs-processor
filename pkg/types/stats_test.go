package processor

import (
    "github.com/stretchr/testify/assert"
    "testing"
    "time"
)

func TestIbcData_Append(t *testing.T) {
    type args struct {
        source      string
        destination string
        t           time.Time
    }
    timeArgs, _ := time.Parse("2006-01-02T15:04:05", "2006-01-02T15:04:05")
    timeWant, _ := time.Parse("2006-01-02T15:00:00", "2006-01-02T15:00:00")
    m := IbcData{}
    tests := []struct {
        name    string
        ibcData IbcData
        args    args
        want    IbcData
    }{
        {
            "test_initial_increment",
            m,
            args{"mySource", "myDestination", timeArgs,},
            map[string]map[string]map[time.Time]int{"mySource": {"myDestination": {timeWant: 1}}},
        },
        {
            "test_increment_existing",
            m,
            args{"mySource", "myDestination", timeArgs,},
            map[string]map[string]map[time.Time]int{"mySource": {"myDestination": {timeWant: 2}}},
        },
        {
            "test_increment_with_second_destination",
            m,
            args{"mySource", "myDestination2", timeArgs,},
            map[string]map[string]map[time.Time]int{"mySource": {"myDestination": {timeWant: 2}, "myDestination2": {timeWant: 1}}},
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.ibcData.Append(tt.args.source, tt.args.destination, tt.args.t)
            assert.Equal(t, tt.want, tt.ibcData)
        })
    }
}
