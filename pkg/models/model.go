package models

import (
	"fmt"
	"time"
)

type Purchase struct {
	Date     Date   `json:"date"`
	ID       string `json:"id"`
	NOK      int    `json:"nok"`
	Account  string `json:"account"`
	Category string `json:"category"`
	Location string `json:"location"`
	Vendor   string `json:"vendor"`
}

type Date struct {
	Year     int        `json:"year"`
	Month    time.Month `json:"month"`
	MonthNum int        `json:"-"`
	Day      int        `json:"day"`
}

func DateToday() Date {
	now := time.Now()
	return Date{
		Year:     now.Year(),
		Month:    now.Month(),
		MonthNum: int(now.Month()),
		Day:      now.Day(),
	}
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
