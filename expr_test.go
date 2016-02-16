package cron

import (
	"testing"
	"time"
)

var matches = []struct {
	field CronFieldType
	te    string
	t     []int
}{
	// any
	{MinuteField, "*", []int{2, 3, 1, 5, 6, 7, 23, 25, 50}},
	{HourField, "*", []int{2, 3, 1, 5, 6, 7, 23}},
	{DayField, "*", []int{2, 3, 1, 5, 6, 7, 23}},
	{MonthField, "*", []int{2, 3, 1, 5, 6, 7}},
	{WeekdayField, "*", []int{2, 3, 1, 5, 6, 0}},

	// number
	{MinuteField, "1", []int{1}},
	{HourField, "2", []int{2}},
	{DayField, "7", []int{7}},
	{MonthField, "12", []int{12}},
	{WeekdayField, "0", []int{0}},

	// step
	{MinuteField, "*/2", []int{2, 4, 6, 8, 24, 28}},
	{HourField, "*/2", []int{2, 4, 6, 8, 22}},
	{DayField, "*/2", []int{2, 4, 6, 8, 24, 28}},
	{MonthField, "*/2", []int{2, 4, 6, 8, 12}},
	{WeekdayField, "*/2", []int{2, 4, 6}},

	// range
	{MinuteField, "1-25", []int{2, 4, 6, 8, 24}},
	{HourField, "2-23", []int{2, 4, 6, 13, 23}},
	{DayField, "1-7", []int{1, 2, 3, 4}},
	{MonthField, "1-12", []int{7, 8, 9, 12}},
	{WeekdayField, "5-7", []int{0, 5, 6}},

	// range with step
	{MinuteField, "0-25/2", []int{2, 4, 6, 8, 24}},
	{HourField, "2-23/3", []int{2, 5, 8, 11, 20}},
	{DayField, "1-7/1", []int{1, 2, 3, 4}},
	{MonthField, "1-12/4", []int{1, 5, 9}},
	{WeekdayField, "5-7/2", []int{5, 0}},

	// list
	{MinuteField, "1,3,6,9,10,20", []int{3, 6, 9}},
	{HourField, "2,5,7,8", []int{2, 5, 8}},
	{DayField, "1,4,5", []int{1, 4}},
	{MonthField, "5,9,12", []int{5, 12}},
	{WeekdayField, "5,7", []int{0, 5}},

	// list with range
	{MinuteField, "1,3,6-9,10,20", []int{3, 6, 7, 8, 9}},
	{HourField, "2,5-8", []int{2, 5, 6, 8}},
	{DayField, "1-4,5", []int{1, 2, 3, 4}},
	{MonthField, "5,6-10/2,12", []int{5, 6, 8, 10, 12}},
	{WeekdayField, "4,5-7", []int{0, 4, 6}},
}

func TestTimeExpr(t *testing.T) {
	for _, test := range matches {
		field, err := NewCronField(test.te, test.field)
		if err != nil {
			t.Fatalf("parse %v, %s", test, err)
		}
		for _, n := range test.t {
			if !field.Match(n) {
				t.Fatalf("match %v", test)
			}
		}
	}
}

var x0 = []struct {
	te string
	ts string
}{
	{"* * * * *", "2015-02-01 10:30"},
	{"30 * * * *", "2015-02-01 10:30"},
	{"30 10 * * *", "2015-02-01 10:30"},
	{"1-33 10 * * *", "2015-02-01 10:30"},
	{"1-33 */10 * * *", "2015-02-01 10:30"},

	{"1-33 */10 1,2,3 * *", "2015-02-01 10:30"},
	{"1-33 */10 1,2-3,4 * *", "2015-02-01 10:30"},

	{"1-33 */10 1 * *", "2015-02-01 10:30"},
	{"1-33 */2 1 * *", "2015-02-01 10:30"},
	{"1-33 */5 1 * *", "2015-02-01 10:30"},
	{"* * * 2 *", "2015-02-01 10:30"},

	{"* * * * 2", "2016-02-16 10:30"},
	{"* * * * 7", "2016-02-14 10:30"},
	{"* * * * 0", "2016-02-14 10:30"},
	{"* * * * 0-5/2", "2016-02-14 10:30"},
	{"* * * * 6-7", "2016-02-14 10:30"},
	{"* * * * 6-7/1", "2016-02-14 10:30"},

	{"* * 15 * 2", "2016-02-16 10:30"},
	{"* * 16 * 6", "2016-02-16 10:30"},
}

const format = "2006-01-02 15:04"

func TestMatchTrue(t *testing.T) {
	for _, x := range x0 {
		expr, err := NewCronExpr(x.te)
		if err != nil {
			t.Fatalf("NewCronExpr: %s %s", x.te, err)
		}
		b, err := time.Parse(format, x.ts)
		if err != nil {
			t.Fatalf("erraaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		}
		if !expr.Match(b) {
			t.Fatalf("Not matched: %s %s", x.te, x.ts)
		}
	}
}
