package output

import (
	"fmt"
	"time"
)

var (
	binaryAbbrs = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
)

func SizeToShow(size *int64) string {
	if size == nil {
		return "/"
	}
	return customSize("%.4g%s", float64(*size), 1024.0, binaryAbbrs)
}

func TimeToShow(created *time.Time) string {
	if created == nil {
		return "-"
	}
	return created.Format(time.RFC3339)
}

func getSizeAndUnit(size float64, base float64, _map []string) (float64, string) {
	i := 0
	unitsLimit := len(_map) - 1
	for size >= base && i < unitsLimit {
		size = size / base
		i++
	}
	return size, _map[i]
}

func customSize(format string, size float64, base float64, _map []string) string {
	size, unit := getSizeAndUnit(size, base, _map)
	return fmt.Sprintf(format, size, unit)
}
