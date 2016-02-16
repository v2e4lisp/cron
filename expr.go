/*
Parse cron expression
http://crontab.org/
*/

package cron

import (
        "errors"
        "fmt"
        "strconv"
        "strings"
        "time"
)

type CronFieldType int

func (f CronFieldType) String() string {
        switch f {
        case MinuteField:
                return "minute field"
        case HourField:
                return "hour field"
        case DayField:
                return "day field"
        case MonthField:
                return "month field"
        case WeekdayField:
                return "weekday field"
        default:
                return "unknown field"
        }
}

const (
        MinuteField CronFieldType = 1 + iota
        HourField
        DayField
        MonthField
        WeekdayField
)

// A cron expression consists of five fields
type CronExpr struct {
        Minute  CronField
        Hour    CronField
        Day     CronField
        Month   CronField
        Weekday CronField
}

func NewCronExpr(timeString string) (expr CronExpr, err error) {
        ss := strings.SplitN(strings.TrimSpace(timeString), " ", 5)
        if len(ss) != 5 {
                err = errors.New("parse error: " + timeString)
                return
        }

        if expr.Minute, err = NewCronField(ss[0], MinuteField); err != nil {
                return
        }
        if expr.Hour, err = NewCronField(ss[1], HourField); err != nil {
                return
        }
        if expr.Day, err = NewCronField(ss[2], DayField); err != nil {
                return
        }
        if expr.Month, err = NewCronField(ss[3], MonthField); err != nil {
                return
        }
        if expr.Weekday, err = NewCronField(ss[4], WeekdayField); err != nil {
                return
        }

        return
}

const any = "*"

func (c CronExpr) Match(t time.Time) bool {
        _, month, day := t.Date()
        if c.Minute.Match(t.Minute()) &&
                c.Hour.Match(t.Hour()) &&
                c.Month.Match(int(month)) {

                // if the day field and the weekday field are both restricted,
                // return true when either field matches the current time.
                if !(c.Day.raw == any) && !(c.Weekday.raw == any) {
                        return c.Weekday.Match(int(t.Weekday())) || c.Day.Match(day)
                }
                return c.Weekday.Match(int(t.Weekday())) && c.Day.Match(day)
        }

        return false
}

// Each cron field contains a time expression.
//
// Time expression can be any of the following forms:
//
//         * list: 1,2,7
//         * range: 3-28 or 1-5/2
//         * asterisk: * or */3
//         * name: JAN, FEB ... for month;  MON, SUN ... for weekday
//
// The valid values of time expressions are as follows:
//
//         field          allowed values
//         -----          --------------
//         minute         0-59
//         hour           0-23
//         day of month   0-31
//         month          0-12 (or names: JAN, FEB)
//         day of week    0-7 (0 or 7 is SUN)
//
// For more detailed information please refer to: http://crontab.org/
type CronField struct {
        raw   string
        field CronFieldType
        t     timeExpr
}

func NewCronField(s string, field CronFieldType) (c CronField, err error) {
        raw := strings.TrimSpace(s)
        if raw == "" {
                return c, errors.New("parse error: empty string")
        }
        t, err := parseTimeExpr(raw, field, enableNames[field])
        if err != nil {
                return c, err
        }

        c.raw = raw
        c.field = field
        c.t = t
        return c, nil
}

func (cf CronField) Match(n int) bool { return cf.t.match(n) }

type timeExpr interface {
        match(n int) bool
}

func parseTimeExpr(s string, field CronFieldType, enableName bool) (timeExpr, error) {
        s = strings.TrimSpace(s)
        if s == "" {
                return nil, errors.New("parse error: empty string")
        }

        ss := strings.Split(s, ",")
        if len(ss) > 1 {
                expr, err := parseList(ss, field)
                return expr, err
        }

        if s[0] == '*' {
                expr, err := parseAny(s, field)
                return expr, err
        }

        startEnd := strings.SplitN(s, "-", 2)
        if len(startEnd) == 2 {
                expr, err := parseRng(startEnd, field)
                return expr, err
        }

        if '0' <= s[0] && s[0] <= '9' {
                expr, err := parseNum(s, field)
                return expr, err
        }

        if !enableName {
                return nil, errors.New("parse error: string is not allowed: " + s)
        }
        expr, err := parseName(s, field)
        return expr, err
}

func parseList(ss []string, field CronFieldType) (timeExpr, error) {
        // list: 11,23
        exprs := make([]timeExpr, len(ss))
        for i, ts := range ss {
                e, err := parseTimeExpr(ts, field, false)
                if err != nil {
                        return nil, err
                }
                exprs[i] = e
        }
        return list{exprs}, nil
}

func parseAny(s string, field CronFieldType) (timeExpr, error) {
        // range: '*' == first-last
        valid := validNums[field]
        step := 1

        if len(s) != 1 {
                // range : '*/13' == first-last/13
                if len(s) < 3 || s[1] != '/' {
                        return nil, errors.New("parse error: " + s)
                }
                _step, err := strconv.ParseInt(s[2:len(s)], 10, 32)
                if err != nil {
                        return nil, err
                }
                step = int(_step)
        }

        expr, err := makeRng(valid[0], valid[1], step, field)
        return expr, err
}

func parseRng(startEnd []string, field CronFieldType) (timeExpr, error) {
        valid := validNums[field]
        endStep := strings.SplitN(startEnd[1], "/", 2)

        _start, err := strconv.ParseInt(startEnd[0], 10, 32)
        if err != nil {
                return nil, err
        }
        start := int(_start)
        if start < valid[0] || start > valid[1] {
                return nil, fmt.Errorf("parse error: invalid number for %s: %d", field, start)
        }

        _end, err := strconv.ParseInt(endStep[0], 10, 32)
        if err != nil {
                return nil, err
        }
        end := int(_end)
        if end < valid[0] || end > valid[1] {
                return nil, fmt.Errorf("parse error: invalid number for %s: %d", field, end)
        }

        step := 1
        // range: 13-31
        if len(endStep) != 1 {
                // range with step: 13-31/2
                _step, err := strconv.ParseInt(endStep[1], 10, 32)
                if err != nil {
                        return nil, err
                }
                step = int(_step)
        }

        expr, err := makeRng(start, end, step, field)
        return expr, err
}

func makeRng(start, end, step int, field CronFieldType) (timeExpr, error) {
        if step <= 0 {
                return nil, fmt.Errorf("parse error: invalid step %d", step)
        }
        // Internally we use 0 to represent SUNDAY. If 7 is used, change it to 0
        if field == WeekdayField && end == 7 && (end-start)%step == 0 {
                return list{[]timeExpr{num{0}, rng{start, end - 1, step}}}, nil
        }
        return rng{start, end, step}, nil
}

func parseName(s string, field CronFieldType) (timeExpr, error) {
        name := strings.ToUpper(s)
        // string: MON
        if field == WeekdayField {
                if wd, ok := weekdays[name]; ok {
                        return num{wd}, nil
                }
                return nil, errors.New("parse error: unknown string for weekday field: " + name)

        }
        // string: JAN
        if field == MonthField {
                if m, ok := months[name]; ok {
                        return num{m}, nil
                }
                return nil, errors.New("parse error: unknown string for month field: " + name)

        }

        return nil, fmt.Errorf("parse error: string is not allowed for %s", field)
}

func parseNum(s string, field CronFieldType) (timeExpr, error) {
        valid := validNums[field]

        // number: '13'
        n, err := strconv.ParseInt(s, 10, 32)
        if err != nil {
                return nil, err
        }
        m := int(n)
        if m < valid[0] || m > valid[1] {
                return nil, fmt.Errorf("parse error: invalid number for %s: %s", field, m)
        }

        // Internally we use 0 to represent SUNDAY. If 7 is used, change it to 0
        if field == WeekdayField && m == 7 {
                m = 0
        }
        return num{m}, nil
}

var validNums = map[CronFieldType][]int{
        MinuteField:  []int{0, 59},
        HourField:    []int{0, 23},
        DayField:     []int{0, 31},
        MonthField:   []int{0, 12},
        WeekdayField: []int{0, 7},
}

var enableNames = map[CronFieldType]bool{
        MinuteField:  false,
        HourField:    false,
        DayField:     false,
        MonthField:   true,
        WeekdayField: true,
}

var months = map[string]int{
        "JAN": 1,
        "FEB": 2,
        "MAR": 3,
        "APR": 4,
        "MAY": 5,
        "JUN": 6,
        "JUL": 7,
        "AUG": 8,
        "SEP": 9,
        "OCT": 10,
        "NOV": 11,
        "DEC": 12,
}

var weekdays = map[string]int{
        "SUN": 0,
        "MON": 1,
        "TUE": 2,
        "WED": 3,
        "THU": 4,
        "FRI": 5,
        "SAT": 6,
}

type num struct{ val int }

func (v num) match(n int) bool { return v.val == n }

type rng struct {
        start int
        end   int
        step  int
}

func (r rng) match(n int) bool { return r.start <= n && n <= r.end && (n-r.start)%r.step == 0 }

type list struct{ exprs []timeExpr }

func (l list) match(n int) bool {
        for _, expr := range l.exprs {
                if expr.match(n) {
                        return true
                }
        }
        return false
}
