package structures

import (
    "errors"
    "time"
)

type TimeRange struct {
    Start time.Time
    End   time.Time
}

// Initialize and validate a TimeRange struct with parameters
func NewTimeRange(start string, end string) (*TimeRange, error) {
    // Return nil if neither start nor end provided
    if len(start) == 0 && len(end) == 0 {
        return nil, nil
    }

    timeRange := TimeRange{}

    var startTime time.Time
    var endTime time.Time

    if len(start) > 0 {
        // Try to parse as RFC3339
        startTime, err := time.Parse(time.RFC3339, start)
        if err != nil {
            // Try to parse as RFC1123
            startTime, err = time.Parse(time.RFC1123, start)
            if err != nil {
                return nil, err
            }
        }
        timeRange.Start = startTime
    }

    if len(end) > 0 {
        // Try to parse as RFC3339
        endTime, err := time.Parse(time.RFC3339, end)
        if err != nil {
            // Try to parse as RFC1123
            endTime, err = time.Parse(time.RFC1123, end)
            if err != nil {
                return nil, err
            }
        }
        timeRange.End = endTime
    }

    // Ensure endTime < startTime
    if len(start) > 0 && len(end) > 0 && endTime.Before(startTime) {
        err := errors.New("start must be before end in a TimeRange")
        return nil, err
    }

    return &timeRange, nil
}
