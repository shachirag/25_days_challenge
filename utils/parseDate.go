package utils

import "time"

func ParseDate(dateStr string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date, err = time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return time.Time{}, err
		}
	}

	return date, nil
}
