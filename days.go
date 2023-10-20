package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// truncNow is the current time, truncated to the day.
var truncNow time.Time

func main() {
	if len(os.Args) < 3 {
		fmt.Println("provide a command (until, since, from) and date arguments")
		return
	}

	truncNow = dayTrunc(time.Now())

	cmd := os.Args[1]
	dates := os.Args[2:]

	if len(dates) > 7 {
		die("too many date arguments")
	}
	if !contains(cmd, []string{"until", "since", "from"}) {
		die("unknown command: %s", cmd)
	}

	times, err := parseDates(cmd, dates)
	if err != nil {
		die("%v", err)
	}
	// Validate
	if (cmd == "until" || cmd == "since") && len(times) != 1 {
		die("too many dates for command '%s'", cmd)
	}
	if cmd == "from" {
		if len(times) != 2 {
			die("command 'from' requires only two dates")
		}
		if times[0].After(times[1]) {
			die("first date occurs after second date, which is invalid for the 'from' command")
		}
	}

	// Print output
	switch cmd {
	case "until":
		fmt.Printf("%d\n", times[0].Sub(truncNow)/(time.Hour*24))
	case "since":
		fmt.Printf("%d\n", truncNow.Sub(times[0])/(time.Hour*24))
	case "from":
		fmt.Printf("%d\n", times[1].Sub(times[0])/(time.Hour*24))
	}
}

func die(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func contains(s string, strs []string) bool {
	for _, v := range strs {
		if v == s {
			return true
		}
	}
	return false
}

func dayTrunc(t time.Time) time.Time {
	// Using .Truncate won't work to set the time to midnight when not in UTC
	// https://github.com/golang/go/issues/10894
	yy, mm, dd := t.Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, t.Location())
}

// parseNoYearDate parses dates like "june 16".
// It uses the current year as the year.
// It always returns a time truncated to the day.
func parseNoYearDate(dates []string) (time.Time, error) {
	dateFmtStr := fmt.Sprintf("%s %s %d", dates[0], dates[1], truncNow.Year())
	t, err := time.ParseInLocation("Jan 2 2006", dateFmtStr, time.Local)
	if err != nil {
		t, err = time.ParseInLocation("January 2 2006", dateFmtStr, time.Local)
		if err != nil {
			return time.Time{}, fmt.Errorf("can't parse date: %w", err)
		}
	}
	return dayTrunc(t), nil
}

// parseNoYearDate parses dates like "feb 23 2004".
// It always returns a time truncated to the day.
func parseYearDate(dates []string) (time.Time, error) {
	dateFmtStr := fmt.Sprintf("%s %s %s", dates[0], dates[1], dates[2])
	t, err := time.ParseInLocation("Jan 2 2006", dateFmtStr, time.Local)
	if err != nil {
		t, err = time.ParseInLocation("January 2 2006", dateFmtStr, time.Local)
		if err != nil {
			return time.Time{}, fmt.Errorf("can't parse date: %w", err)
		}
	}
	return dayTrunc(t), nil
}

// parseDates parses arguments from the command-line into actual time.Time dates.
// The caller needs to check the command against the number of returned values,
// as this function does not fully validate everything.
func parseDates(cmd string, argdates []string) ([]time.Time, error) {
	// Split on spaces to simplify parsing
	dates := make([]string, 0)
	for _, date := range argdates {
		for _, s := range strings.Split(strings.ToLower(date), " ") {
			if s == "to" {
				// Ignore "to" to allow for commands like: from jun 1 to aug 1
				continue
			}
			dates = append(dates, s)
		}
	}

	times := make([]time.Time, 0)

	switch len(dates) {
	case 0, 1:
		return nil, fmt.Errorf("too few date arguments")
	case 2:
		// Day with no year, like "june 16" or "feb 23"
		// Year must be determined. The final date must be in the future for
		// the "until" command, and the past for the "since" command.
		// If today is that day, the year is kept as this year.

		t, err := parseNoYearDate(dates)
		if err != nil {
			return nil, err
		}
		if cmd == "until" {
			if t.Before(truncNow) {
				// Must be in the future so increment the year
				t = t.AddDate(1, 0, 0)
			}
		} else { // "since"
			if t.After(truncNow) {
				// Must be in the past so decrement the year
				t = t.AddDate(-1, 0, 0)
			}
		}
		times = append(times, t)

	case 3:
		// Single date with year, like "feb 23 2004"
		t, err := parseYearDate(dates)
		if err != nil {
			return nil, err
		}
		times = append(times, t)

	case 4:
		// Two dates with no year, like "jan 3 march 3"
		// Both years are assumed to be the current year, unless that would
		// put the first date after the second date. In that case the year of
		// the second date is incremented

		t1, err := parseNoYearDate(dates[:2])
		if err != nil {
			return nil, err
		}
		t2, err := parseNoYearDate(dates[2:])
		if err != nil {
			return nil, err
		}
		if t1.After(t2) {
			t2 = t2.AddDate(1, 0, 0)
		}
		times = append(times, t1, t2)
	case 5:
		// One date with a year and one without, but the order is unknown
		if len(dates[2]) == 4 {
			// The third arg is 4 chars long, so it must be a year
			// So the format is: jan 3 2004 march 3
			//
			// Current year is assumed for the second date. Having a negative
			// time difference is invalid in this case but that's handled elsewhere

			t1, err := parseYearDate(dates[:3])
			if err != nil {
				return nil, err
			}
			t2, err := parseNoYearDate(dates[3:])
			if err != nil {
				return nil, err
			}
			times = append(times, t1, t2)
		} else {
			// Format is the opposite way: jan 3 march 3 2030

			t1, err := parseNoYearDate(dates[:2])
			if err != nil {
				return nil, err
			}
			t2, err := parseYearDate(dates[2:])
			if err != nil {
				return nil, err
			}
			times = append(times, t1, t2)
		}
	case 6:
		// Two dates with years, like "jan 3 2004 march 3 2006"
		t1, err := parseYearDate(dates[:3])
		if err != nil {
			return nil, err
		}
		t2, err := parseYearDate(dates[3:])
		if err != nil {
			return nil, err
		}
		times = append(times, t1, t2)
	default:
		// Anything beyond 6 is is guaranteed not to happen due to checks in main()
		return nil, fmt.Errorf("invalid number of date args (%d) -- how did we get here?", len(dates))
	}

	return times, nil
}
