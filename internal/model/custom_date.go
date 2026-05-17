package model

import (
	"fmt"
	"strings"
	"time"
)

type CustomDate struct {
	Month int
	Year  int
}

const dateLayout = "01-2006"

func (cd *CustomDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	return cd.Parse(s)
}

func (cd CustomDate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + cd.String() + `"`), nil
}

func (cd *CustomDate) Parse(s string) error {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return fmt.Errorf("неверный формат даты: %s, ожидается MM-YYYY", s)
	}
	cd.Month = int(t.Month())
	cd.Year = t.Year()
	return nil
}

func (cd CustomDate) String() string {
	return fmt.Sprintf("%02d-%d", cd.Month, cd.Year)
}

func (cd CustomDate) ToTime() time.Time {
	return time.Date(cd.Year, time.Month(cd.Month), 1, 0, 0, 0, 0, time.UTC)
}

func ParseDate(dateStr string) (CustomDate, error) {
	var cd CustomDate
	err := cd.Parse(dateStr)
	return cd, err
}

func FirstDayOfMonth(cd CustomDate) time.Time {
	return time.Date(cd.Year, time.Month(cd.Month), 1, 0, 0, 0, 0, time.UTC)
}

func LastDayOfMonth(cd CustomDate) time.Time {
	return time.Date(cd.Year, time.Month(cd.Month)+1, 0, 0, 0, 0, 0, time.UTC)
}