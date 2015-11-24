package database

import (
    "errors"
    "time"
)

type TimeRange struct {
    Start       time.Time
    End         time.Time
}

// Initialize and validate a TimeRange struct with parameters
func NewTimeRange(start string, end string) (*TimeRange, error) {
    // Return nil if neither start nor end provided
    if len(start) == 0 && len(end) == 0 {
        return nil, nil
    }

    timeRange := TimeRange{}

    var startTime   time.Time
    var endTime     time.Time

    if len(start) > 0 {
        startTime, err := time.Parse(time.RFC3339, start)
        if err != nil {
            return nil, err
        }
        timeRange.Start = startTime
    }

    if len(end) > 0 {
        endTime, err := time.Parse(time.RFC3339, end)
        if err != nil {
            return nil, err
        }
        timeRange.End = endTime
    }

    // Ensure endTime < startTime
    if len(start) > 0 && len(end) > 0 && endTime.Before(startTime) {
        err := errors.New("end must be before start in a TimeRange")
        return nil, err
    }

    return &timeRange, nil
}

// Validates a string interval and returns its value in seconds
func NormalizeInterval(intervalStr string) (int, error) {
    // Map allowed durations in seconds
    durations := map[string]int{
        "1 hour":   60*60*1000,
        "6 hours":  6*60*60*1000,
        "1 day":    24*60*60*1000,
        "7 days":   7*24*60*60*1000,
        "30 days":  30*24*60*60*1000,
    }

    if interval, ok := durations[intervalStr]; !ok {
        err := errors.New("Wrong interval in query. Please check the documentations and retry.")
        return 0, err
    } else {
        return interval, nil
    }
}
