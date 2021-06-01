package cron

import (
	"strings"
	"testing"
	"time"
)

func TestGetBits(t *testing.T) {
	datas := []struct {
		min, max, step uint
		result         uint64
	}{
		{0, 0, 1, 0x1},                  // 0001
		{1, 3, 1, 0xe},                  // 1110
		{1, 4, 2, 0xa},                  // 1010
		{1, 5, 3, 0x12},                 // 10010
		{0, 59, 1, 1152921504606846975}, // 1 * 60
	}
	for _, data := range datas {
		result := getBits(data.min, data.max, data.step)
		if result != data.result {
			t.Logf("%b\n", result)
			t.Fatalf("getBits(%d, %d, %d) => %d, but got %d.", data.min, data.max, data.step, data.result, result)
		}
	}
}

func TestParseIntOrName(t *testing.T) {
	datas := []struct {
		expr   string
		names  map[string]uint
		result uint
		err    string
	}{
		{
			expr:   "jan",
			names:  months.names,
			result: 1,
		},
		{
			expr:   "wed",
			names:  dayOfWeek.names,
			result: 3,
		},
		{
			expr:   "10",
			names:  nil,
			result: 10,
		},
		{
			expr:   "-10",
			names:  nil,
			result: 0,
			err:    "invalid syntax",
		},
		{
			expr:   "test",
			names:  nil,
			result: 0,
			err:    "invalid syntax",
		},
	}
	for _, data := range datas {
		result, err := parseIntOrName(data.expr, data.names)
		if err != nil {
			if result != data.result || !strings.Contains(err.Error(), data.err) {
				t.Fatalf("parseIntOrName(%s, %v) => (%d, ...%s), but got (%d, %s)", data.expr, data.names, data.result, data.err, result, err)
			}
		} else {
			if result != data.result {
				t.Fatalf("parseIntOrName(%s, %v) => (%d, nil), but got (%d, nil)", data.expr, data.names, data.result, result)
			}
		}
	}
}

func TestParseExpr(t *testing.T) {
	datas := []struct {
		expr   string
		b      bounds
		result uint64
		err    string
	}{
		{
			expr:   "*",
			b:      dayOfWeek,
			result: 0x7f | starBit, // 1111111
		},
		{
			expr:   "2",
			b:      seconds,
			result: 0x4, // 0100
		},
		{
			expr:   "1-3",
			b:      seconds,
			result: 0xe, // 1110
		},
		{
			expr:   "*/2",
			b:      dayOfWeek,
			result: 0x55, // 1010101
		},
		{
			expr:   "3/2",
			b:      dayOfWeek,
			result: 0x28, // 0101000
		},
		{
			expr:   "wed/2",
			b:      dayOfWeek,
			result: 0x28, // 0101000
		},
		{
			expr:   "3-1",
			b:      dayOfWeek,
			result: 0,
			err:    "beyond end of range",
		},
		{
			expr:   "*/0",
			b:      seconds,
			result: 0,
			err:    "The step (0) is invalid:",
		},
		{
			expr:   "0-2",
			b:      dayOfMonth,
			result: 0,
			err:    "The effective range is ",
		},
		{
			expr:   "50-61",
			b:      seconds,
			result: 0,
			err:    "The effective range is ",
		},
		{
			expr:   "1/2/",
			b:      seconds,
			result: 0,
			err:    "Too many slashes:",
		},
		{
			expr:   "1-2-",
			b:      seconds,
			result: 0,
			err:    "Too many hypends:",
		},
		{
			expr:   "1-10/3",
			b:      seconds,
			result: 0x492, // 10010010010
			err:    "",
		},
		{
			expr:   "a/3",
			b:      seconds,
			result: 0,
			err:    "invalid syntax",
		},
	}
	for _, data := range datas {
		result, err := parseExpr(data.expr, data.b)
		if err != nil {
			if result != data.result || !strings.Contains(err.Error(), data.err) {
				t.Fatalf("parseExpr(%s, %v) => (%d, ...%s...), but got (%d, %s)", data.expr, data.b, data.result, data.err, result, err)
			}
		} else {
			if result != data.result {
				t.Fatalf("parseExpr(%s, %v) => (%d, nil), but got (%d, nil)", data.expr, data.b, data.result, result)
			}
		}
	}
}

func TestGetField(t *testing.T) {
	datas := []struct {
		field  string
		b      bounds
		result uint64
		err    string
	}{
		{
			field:  "*",
			b:      dayOfWeek,
			result: 0x7f | starBit, // 1111111
		},
		{
			field:  "*/2",
			b:      dayOfWeek,
			result: 0x55, // 1010101
		},
		{
			field:  "1-10",
			b:      dayOfWeek,
			result: 0,
			err:    "The effective range is ",
		},
		{
			field:  "sun,mon,wed",
			b:      dayOfWeek,
			result: 0x0b, // 0001011
		},
		{
			field:  "sun,wed-fri/2",
			b:      dayOfWeek,
			result: 0x29, // 0101001
		},
		{
			field:  "error-string",
			b:      seconds,
			result: 0,
			err:    "invalid syntax",
		},
	}
	for _, data := range datas {
		result, err := getField(data.field, data.b)
		if err != nil {
			if result != data.result || !strings.Contains(err.Error(), data.err) {
				t.Fatalf("getField(%s, %v) => (%d, ...%s...), but got (%d, %s)", data.field, data.b, data.result, data.err, result, err)
			}
		} else {
			if result != data.result {
				t.Fatalf("getField(%s, %v) => (%d, nil), but got (%d, nil)", data.field, data.b, data.result, result)
			}
		}
	}
}

func TestNormalizeFields(t *testing.T) {
	datas := []struct {
		fields  []string
		options ParseOption
		result  []string
		err     string
	}{
		{
			fields:  []string{"0"},
			options: ParseOptionStandard,
			result:  nil,
			err:     "Parser accept",
		},
		{
			fields:  []string{"0", "1", "2", "3", "4", "5"},
			options: ParseOptionStandard,
			result:  nil,
			err:     "Parser accept",
		},
		{
			fields:  []string{"0", "1", "2", "3", "4"},
			options: ParseOptionStandard,
			result:  []string{"0", "0", "1", "2", "3", "4"},
		},
		{
			fields:  []string{"1", "*", "*", "*", "*", "*"},
			options: ParseOptionAll,
			result:  []string{"1", "*", "*", "*", "*", "*"},
		},
	}
	for _, data := range datas {
		result, err := normalizeFields(data.fields, data.options)
		if err != nil {
			if !compareStringSlice(result, data.result) || !strings.Contains(err.Error(), data.err) {
				t.Fatalf("normalizeFields(%v, %v) => (%v, ...%s...), but got (%v, %s)", data.fields, data.options, data.result, data.err, result, err)
			}
		} else {
			if !compareStringSlice(result, data.result) {
				t.Fatalf("normalizeFields(%v, %v) => (%v, nil), but got (%v, nil)", data.fields, data.options, data.result, result)
			}
		}
	}
}

func TestParseDescriptor(t *testing.T) {

}

func TestParse(t *testing.T) {
	parser := NewParser(ParseOptionAll)
	schedule, err := parser.Parse("0 30 8 3-5 * 0")
	if err != nil {
		t.Fatalf("build schedule failure: %s", err)
	}
	resultArr := []string{
		"2021-06-03 08:30:00",
		"2021-06-04 08:30:00",
		"2021-06-05 08:30:00",
		"2021-06-06 08:30:00",
		"2021-06-13 08:30:00",
		"2021-06-20 08:30:00",
		"2021-06-27 08:30:00",
	}

	ti := time.Date(2021, 6, 1, 12, 0, 0, 0, time.Local)
	for _, r := range resultArr {
		next := schedule.Next(ti)
		result, err := time.ParseInLocation("2006-01-02 15:04:05", r, time.Local)
		if err != nil {
			t.Fatalf("parse time err: %s", err)
		}
		if !result.Equal(next) {
			t.Fatalf("schedule.Next(%v) => (%v), but got %v", ti, result, next)
		}
		ti = next
	}
}

func compareStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestEverySchedule(t *testing.T) {
	duration, err := time.ParseDuration("5s60ms")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	t.Log(duration / time.Second * time.Second)
	t.Log(duration - time.Duration(duration.Nanoseconds())%time.Second)
	// s := &EverySchedule{duration}
	s := Every(duration)
	resultArr := []string{
		"2021-06-01 12:00:05",
		"2021-06-01 12:00:10",
		"2021-06-01 12:00:15",
		"2021-06-01 12:00:20",
		"2021-06-01 12:00:25",
		"2021-06-01 12:00:30",
		"2021-06-01 12:00:35",
		"2021-06-01 12:00:40",
	}

	ti := time.Date(2021, 6, 1, 12, 0, 0, 0, time.Local)
	for _, r := range resultArr {
		next := s.Next(ti)
		result, err := time.ParseInLocation("2006-01-02 15:04:05", r, time.Local)
		if err != nil {
			t.Fatalf("parse time err: %s", err)
		}
		if !result.Equal(next) {
			t.Fatalf("schedule.Next(%v) => (%v), but got %v", ti, result, next)
		}
		ti = next
	}
}
