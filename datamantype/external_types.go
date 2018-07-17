package datamantype

import "time"

type Time time.Time

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(t).Format(DateTimeFormatStr) + `"`), nil
}
