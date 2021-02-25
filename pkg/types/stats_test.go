package processor

import (
    "github.com/stretchr/testify/assert"
    "reflect"
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

func TestIbcData_ToIbcStats(t *testing.T) {
    timeArgs, _ := time.Parse("2006-01-02T15:04:05", "2006-01-02T15:04:05")
    sourceName := "mySource"
    destination1 := "myDestination"
    destination2 := "myDestination2"
    counter1 := 2
    counter2 := 7
    tests := []struct {
        name     string
        ibcData  IbcData
        expected [][]IbcStats
    }{
        {
            "IbcData(map)_to_IbcStats(slice)",
            map[string]map[string]map[time.Time]int{sourceName: {destination1: {timeArgs: counter1}, destination2: {timeArgs: counter2}}},
            [][]IbcStats{
                {
                    {sourceName, destination1, timeArgs, counter1},
                    {sourceName, destination2, timeArgs, counter2},
                },
                {
                    {sourceName, destination2, timeArgs, counter2},
                    {sourceName, destination1, timeArgs, counter1},
                },
            },
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := tt.ibcData.ToIbcStats()

            if !reflect.DeepEqual(tt.expected[0], actual) {
                assert.Equal(t, tt.expected[1], actual)
            } else {
                assert.NotEqual(t, tt.expected[1], actual)
            }

            if !reflect.DeepEqual(tt.expected[1], actual) {
                assert.Equal(t, tt.expected[0], actual)
            } else {
                assert.NotEqual(t, tt.expected[0], actual)
            }
        })
    }
}
