package utils

import "time"

func ParseDate(dateStr string) (time.Time, error) {
	date, err := time.Parse("02-01-2006", dateStr)
	if err != nil {
		date, err = time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return time.Time{}, err
		}
	}

	return date, nil
}
