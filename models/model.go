package models

import (
	"fmt"
	"time"
)

type Purchase struct {
	Date     Date
	ID       string
	NOK      int
	Account  string
	Category string
	Location string
	Vendor   string
}

type Date struct {
	Year     int
	Month    time.Month
	MonthNum int
	Day      int
}

func (d Date) String() string {
	if d.Day == 0 {
		return fmt.Sprintf("%s %04d", d.Month, d.Year)
	}
	return fmt.Sprintf("%d %s, %04d", d.Day, d.Month, d.Year)
}

func (d Date) Stamp() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

func (d Date) AddMonth() Date {
	if d.Month == time.December {
		return Date{Year: d.Year + 1, Month: time.January, MonthNum: int(time.January), Day: d.Day}
	}
	return Date{Year: d.Year, Month: d.Month + 1, MonthNum: int(d.Month + 1), Day: d.Day}
}

func (d Date) SubMonth() Date {
	if d.Month == time.January {
		return Date{Year: d.Year - 1, Month: time.December, MonthNum: int(time.December), Day: d.Day}
	}
	return Date{Year: d.Year, Month: d.Month - 1, MonthNum: int(d.Month - 1), Day: d.Day}
}
