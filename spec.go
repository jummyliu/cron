package cron

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type ParseOption int

const (
	starBit = 1 << 63
)

const (
	Second ParseOption = 1 << iota
	Minute
	Hour
	DayOfMonth
	Month
	DayOfWeek
	Descriptor

	// ParseOptionAll Second | Minute | Hour | DayOfMonth | Month | DayOfWeek | Descriptor
	ParseOptionAll = Second | Minute | Hour | DayOfMonth | Month | DayOfWeek | Descriptor
	// ParseOptionStandard Minute | Hour | DayOfMonth | Month | DayOfWeek | Descriptor
	//
	// Without second.
	ParseOptionStandard = Minute | Hour | DayOfMonth | Month | DayOfWeek | Descriptor
)

const (
	DescriptorYearly      = "@yearly"
	DescriptorAnnually    = "@annually"
	DescriptorMonthly     = "@monthly"
	DescriptorWeekly      = "@weekly"
	DescriptorDaily       = "@daily"
	DescriptorMidnight    = "@midnight"
	DescriptorHourly      = "@hourly"
	DescriptorEveryPrefix = "@every "
)

// places of all fields
var places = []ParseOption{
	Second,
	Minute,
	Hour,
	DayOfMonth,
	Month,
	DayOfWeek,
}

// defaults of all fields
var defaults = []string{
	"0",
	"0",
	"0",
	"*",
	"*",
	"*",
}

type Parser interface {
	Parse(spec string) (Schedule, error)
}

type SpecParser struct {
	options ParseOption
}

func NewParser(options ParseOption) Parser {
	return &SpecParser{options & ParseOptionAll}
}

func (p *SpecParser) Parse(spec string) (Schedule, error) {
	var err error
	// descriptor
	if strings.HasPrefix(spec, "@") {
		if p.options&Descriptor == 0 {
			return nil, fmt.Errorf("Parser does not accept descriptor: %v", spec)
		}
		return parseDescriptor(spec)
	}
	// normalize
	fields, err := normalizeFields(strings.Fields(spec), p.options)
	if err != nil {
		return nil, err
	}

	fieldWrap := func(field string, b bounds) uint64 {
		if err != nil {
			return 0
		}
		var bits uint64
		bits, err = getField(field, b)
		return bits
	}
	var (
		second = fieldWrap(fields[0], seconds)
		minute = fieldWrap(fields[1], minutes)
		hour   = fieldWrap(fields[2], hours)
		dom    = fieldWrap(fields[3], dayOfMonth)
		month  = fieldWrap(fields[4], months)
		dow    = fieldWrap(fields[5], dayOfWeek)
	)
	if err != nil {
		return nil, err
	}

	return &SpecSchedule{
		Second:     second,
		Minute:     minute,
		Hour:       hour,
		DayOfMonth: dom,
		Month:      month,
		DayOfWeek:  dow,
	}, nil
}

func normalizeFields(fields []string, options ParseOption) ([]string, error) {
	count := 0
	for _, place := range places {
		if place&options > 0 {
			count++
		}
	}
	if count != len(fields) {
		return nil, fmt.Errorf("Parser accept %d fields, found %d: %v", count, len(fields), fields)
	}
	normalize := make([]string, len(places))
	copy(normalize, defaults)
	n := 0
	for i, place := range places {
		if place&options > 0 {
			normalize[i] = fields[n]
			n++
		}
	}
	return normalize, nil
}

// * - , /
// *
// 1-10
// */3
// 5/3
// 1-10,3,4
func getField(field string, b bounds) (uint64, error) {
	exprs := strings.FieldsFunc(field, func(r rune) bool { return r == ',' })
	var bits uint64
	for _, expr := range exprs {
		bit, err := parseExpr(expr, b)
		if err != nil {
			return 0, err
		}
		bits |= bit
	}
	return bits, nil
}

func parseDescriptor(expr string) (Schedule, error) {
	switch expr {
	case DescriptorYearly, DescriptorAnnually:
		return &SpecSchedule{
			Second:     1 << seconds.min,
			Minute:     1 << minutes.min,
			Hour:       1 << hours.min,
			DayOfMonth: 1 << dayOfMonth.min,
			Month:      1 << months.min,
			DayOfWeek:  getBits(dayOfWeek.min, dayOfWeek.max, 1),
		}, nil
	case DescriptorMonthly:
		return &SpecSchedule{
			Second:     1 << seconds.min,
			Minute:     1 << minutes.min,
			Hour:       1 << hours.min,
			DayOfMonth: 1 << dayOfMonth.min,
			Month:      getBits(months.min, months.max, 1),
			DayOfWeek:  getBits(dayOfWeek.min, dayOfWeek.max, 1),
		}, nil
	case DescriptorWeekly:
		return &SpecSchedule{
			Second:     1 << seconds.min,
			Minute:     1 << minutes.min,
			Hour:       1 << hours.min,
			DayOfMonth: getBits(dayOfMonth.min, dayOfMonth.max, 1),
			Month:      getBits(months.min, months.max, 1),
			DayOfWeek:  1 << dayOfWeek.min,
		}, nil
	case DescriptorDaily, DescriptorMidnight:
		return &SpecSchedule{
			Second:     1 << seconds.min,
			Minute:     1 << minutes.min,
			Hour:       1 << hours.min,
			DayOfMonth: getBits(dayOfMonth.min, dayOfMonth.max, 1),
			Month:      getBits(months.min, months.max, 1),
			DayOfWeek:  getBits(dayOfWeek.min, dayOfWeek.max, 1),
		}, nil
	case DescriptorHourly:
		return &SpecSchedule{
			Second:     1 << seconds.min,
			Minute:     1 << minutes.min,
			Hour:       getBits(hours.min, hours.max, 1),
			DayOfMonth: getBits(dayOfMonth.min, dayOfMonth.max, 1),
			Month:      getBits(months.min, months.max, 1),
			DayOfWeek:  getBits(dayOfWeek.min, dayOfWeek.max, 1),
		}, nil
	}
	if strings.HasPrefix(expr, DescriptorEveryPrefix) {
		duration, err := time.ParseDuration(expr[len(DescriptorEveryPrefix):])
		if err != nil {
			return nil, fmt.Errorf("Parse duration failure: %s", err)
		}
		return Every(duration), nil
	}
	return nil, fmt.Errorf("Invalid descriptor: %s", expr)
}

func parseExpr(expr string, b bounds) (uint64, error) {
	// *
	// 2
	// 1-10
	// */3 5/3
	var (
		min, max, step uint
		rangeAndStep   = strings.Split(expr, "/")
		lowToHigh      = strings.Split(rangeAndStep[0], "-")
		singleDigit    = len(lowToHigh) == 1
		err            error
	)
	var extra uint64
	if lowToHigh[0] == "*" || lowToHigh[0] == "?" {
		min = b.min
		max = b.max
		extra = starBit
	} else {
		min, err = parseIntOrName(lowToHigh[0], b.names)
		if err != nil {
			return 0, err
		}
		switch len(lowToHigh) {
		case 1:
			max = min
		case 2:
			max, err = parseIntOrName(lowToHigh[1], b.names)
			if err != nil {
				return 0, err
			}
		default:
			return 0, fmt.Errorf("Too many hypends: %s", expr)
		}
	}
	switch len(rangeAndStep) {
	case 1:
		step = 1
	case 2:
		step, err = parseIntOrName(rangeAndStep[1], nil)
		if err != nil {
			return 0, err
		}
		// N/step means N-max/step
		if singleDigit {
			max = b.max
		}
		if step > 1 {
			extra = 0
		}
	default:
		return 0, fmt.Errorf("Too many slashes: %s", expr)
	}
	if min < b.min || max > b.max {
		return 0, fmt.Errorf("The effective range is [%d, %d], but got [%d, %d]: %s", b.min, b.max, min, max, expr)
	}
	if min > max {
		return 0, fmt.Errorf("Beginning of range (%d) beyond end of range (%d): %s", min, max, expr)
	}
	if step == 0 {
		return 0, fmt.Errorf("The step (0) is invalid: %s", expr)
	}
	return getBits(min, max, step) | extra, nil
}

func getBits(min, max, step uint) uint64 {
	if step == 1 {
		return ^(math.MaxUint64 << (max + 1)) & (math.MaxUint64 << min)
	}
	var bits uint64
	for i := min; i <= max; i += step {
		bits |= 1 << i
	}
	return bits
}

func parseIntOrName(expr string, names map[string]uint) (uint, error) {
	if val, ok := names[expr]; ok {
		return val, nil
	}
	val, err := strconv.ParseUint(expr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(val), nil
}

type Schedule interface {
	Next(t time.Time) time.Time
}

type SpecSchedule struct {
	Second, Minute, Hour, DayOfMonth, Month, DayOfWeek uint64
}

type bounds struct {
	min, max uint
	names    map[string]uint
}

var (
	seconds    = bounds{0, 59, nil}
	minutes    = bounds{0, 59, nil}
	hours      = bounds{0, 23, nil}
	dayOfMonth = bounds{1, 31, nil}
	months     = bounds{1, 12, map[string]uint{
		"jan": 1,
		"feb": 2,
		"mar": 3,
		"apr": 4,
		"may": 5,
		"jun": 6,
		"jul": 7,
		"aug": 8,
		"sep": 9,
		"oct": 10,
		"nov": 11,
		"dec": 12,
	}}
	dayOfWeek = bounds{0, 6, map[string]uint{
		"sun": 0,
		"mon": 1,
		"tue": 2,
		"wed": 3,
		"thu": 4,
		"fri": 5,
		"sat": 6,
	}}
)

// Next caculate the next time by spec
func (s *SpecSchedule) Next(t time.Time) time.Time {
	t = t.Add(time.Second).Truncate(time.Second)
	// Prevent leap year
	maxYear := t.Year() + 5

	added := false
	continued := false

	for t.Year() <= maxYear {
		continued = false

		for s.Month&(1<<uint64(t.Month())) == 0 {
			if !added {
				added = true
				t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
			}
			t = t.AddDate(0, 1, 0)
			if t.Month() == time.January {
				continued = true
			}
		}
		if continued {
			continue
		}
		for !dayMatches(s, t) {
			if !added {
				added = true
				t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
			}
			t = t.AddDate(0, 0, 1)
			// TODO: 处理夏令时
			if t.Day() == 1 {
				continued = true
			}
		}
		if continued {
			continue
		}
		for s.Hour&(1<<uint64(t.Hour())) == 0 {
			if !added {
				added = true
				t = t.Truncate(time.Hour)
			}
			t = t.Add(1 * time.Hour)
			if t.Hour() == 0 {
				continued = true
			}
		}
		if continued {
			continue
		}
		for s.Minute&(1<<uint64(t.Minute())) == 0 {
			if !added {
				added = true
				t = t.Truncate(time.Minute)
			}
			t = t.Add(1 * time.Minute)
			if t.Minute() == 0 {
				continued = true
			}
		}
		if continued {
			continue
		}
		for s.Second&(1<<uint64(t.Second())) == 0 {
			if !added {
				added = true
				t = t.Truncate(time.Second)
			}
			t = t.Add(1 * time.Second)
			if t.Second() == 0 {
				continued = true
			}
		}
		if continued {
			continue
		}
		return t
	}
	return time.Time{}
}

// dayMatches
func dayMatches(s *SpecSchedule, t time.Time) bool {
	var (
		domMatch bool = 1<<uint64(t.Day())&s.DayOfMonth > 0
		dowMatch bool = 1<<uint64(t.Weekday())&s.DayOfWeek > 0
	)
	if s.DayOfMonth&starBit > 0 || s.DayOfWeek&starBit > 0 {
		return domMatch && dowMatch
	}
	return domMatch || dowMatch
}

type EverySchedule struct {
	Delay time.Duration
}

func Every(duration time.Duration) *EverySchedule {
	if duration < time.Second {
		duration = time.Second
	}
	return &EverySchedule{
		Delay: duration / time.Second * time.Second,
	}
}

func (s *EverySchedule) Next(t time.Time) time.Time {
	return t.Add(s.Delay).Truncate(time.Second)
}
