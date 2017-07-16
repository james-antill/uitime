package main

import (
    "flag"
    "fmt"
    "os"
    "sort"
    "strings"
    "time"
    "golang.org/x/crypto/ssh/terminal"
    "github.com/fatih/color"
)

// From: https://github.com/lucylw/dec_datetime
func dectime(now time.Time, showsecs bool) string {
    day10 := now.YearDay()

    // get total minutes into a day: 24*60 = 1440 minutes
    // ratio is 1440 minutes to 1000 minutes according to decimal clock
    var totalMin24 float64
    totalMin24 = float64(now.Hour()*60 + now.Minute())
    totalMin10 := totalMin24 * 100.0/144.0
    hours10 := int(totalMin10/100)
    minutes10 := int(totalMin10) % 100

    if showsecs {
        seconds10 := now.Nanosecond() / 1000000
        seconds10 += now.Second() * 1000
        seconds10 /= 600

        return fmt.Sprintf("%03d %02d:%02d:%02d", day10, hours10, minutes10, seconds10)
    }

    return fmt.Sprintf("%03d %02d:%02d", day10, hours10, minutes10)
}

func calc_weekday(datetime string, tm time.Time) time.Time {
    otm := tm
    num := 0
    d, _ := time.ParseDuration("24h")
    for !strings.HasPrefix(datetime, tm.Weekday().String())  {
        tm = tm.Add(d)
        if num++; num >= 7 {
            return otm
        }
    }

    return tm
}

func ptime(now time.Time, datetime string) time.Time {
    // 1. First try full date/time/loc stamps and just use them
    fmts := []string{"2006-01-02 15:04", "2006-01-02 15:04:05",
                     "2006-01-02T15:04", "2006-01-02T15:04:05",
                     time.RFC3339,
                     time.UnixDate,
                     time.RFC822, time.RFC822Z,
                     time.RFC850,
                     time.RFC1123, time.RFC1123Z,
                     time.RFC3339}

    for _, tfmt := range fmts {
        tm, ok := time.Parse(tfmt, datetime)
        if ok == nil {
            return tm
        }
    }

    // 2. Inherit the loc, parsing a full date/time.
    fmts = []string{"2006-01-02 15:04", "2006-01-02 15:04:05",
                    "2006-01-02T15:04", "2006-01-02T15:04:05",
                    time.Stamp, time.ANSIC,
                    "02 Jan 06 15:04", "Monday, 02-Jan-06 15:04:05",
                    "Mon, 02 Jan 2006 15:04:05"}
    for _, tfmt := range fmts {
        tfmt += " MST"
        tdatetime := datetime + now.Format(" MST")
        tm, ok := time.Parse(tfmt, tdatetime)
        if ok == nil {
            return tm
        }
    }

    // 3. Inherit the time, parsing the date/loc
    fmts = []string{"2006-01-02 MST", "Jan _2 2006 MST", "Jan 02 2006 MST",
                    "2006-01-02Z07:00", "Jan _2 2006Z07:00", "Jan 02 2006Z07:00",
                    "2006-01-02 -0700", "Jan _2 2006 -0700", "Jan 02 2006 -0700"}
    for _, tfmt := range fmts { // Probably doesn't work right, at weird times
        tm, ok := time.Parse(tfmt, datetime)
        if ok == nil {
            return time.Date(tm.Year(), tm.Month(), tm.Day(),
                             now.Hour(), now.Minute(), now.Second(), now.Nanosecond(),
                             tm.Location())
        }
    }

    // 4. Inherit the time/loc, parsing the date.
    fmts = []string{"2006-01-02", "Jan _2 2006", "Jan 02 2006"}
    for _, tfmt := range fmts {
        tfmt += " 15:04:05.999999999 MST"
        tdatetime := datetime + now.Format(" 15:04:05.999999999 MST")
        tm, ok := time.Parse(tfmt, tdatetime)
        if ok == nil {
            return tm
        }
    }

    // 5. Inherit the date, parsing the time/loc.
    fmts = []string{time.Kitchen + " MST", "15:04:05 MST", "15:04 MST",
                    time.Kitchen + "Z07:00", "15:04:05Z07:00", "15:04Z07:00",
                    time.Kitchen + " -0700", "15:04:05 -0700", "15:04 -0700"}
    for _, tfmt := range fmts {
        tfmt = "2006-01-02 " + tfmt
        tdatetime := now.Format("2006-01-02 ") + datetime
        tm, ok := time.Parse(tfmt, tdatetime)
        if ok == nil {
            return tm
        }
    }

    // 6. Inherit the date/loc, parsing the time.
    fmts = []string{time.Kitchen, "15:04:05", "15:04"}
    for _, tfmt := range fmts {
        tfmt = "2006-01-02 " + tfmt + " MST"
        tdatetime := now.Format("2006-01-02 ") + datetime + now.Format(" MST")
        tm, ok := time.Parse(tfmt, tdatetime)
        if ok == nil {
            return tm
        }
    }

    // Now, for bits:
    // 1. Inherit the year/time, parsing the month-day/loc.

    // 2. Inherit the year/time/loc, parsing the month-day.
    fmts = []string{"01-02", "Jan _2", "Jan 02"}
    for _, tfmt := range fmts {
        time_fmt := " 15:04:05.999999999 MST"
        tfmt1 := "2006-" + tfmt + time_fmt
        tdatetime1 := now.Format("2006-") + datetime + now.Format(time_fmt)
        tm1, ok := time.Parse(tfmt1, tdatetime1)
        if ok == nil {
            return tm1
        }

        tfmt2 := tfmt + " 2006" + time_fmt
        tdatetime2 := datetime + now.Format(" 2006" + time_fmt)
        tm2, ok := time.Parse(tfmt2, tdatetime2)
        if ok == nil {
            return tm2
        }
    }

    // 3. Inherit the year-month/time, parsing the day/loc.

    // 4. Inherit the year-month/time/loc, parsing the day.
    fmts = []string{"02", "_2"}
    for _, tfmt := range fmts {
        time_fmt := " 15:04:05.999999999 MST"
        tfmt1 := "2006-01-" + tfmt + time_fmt
        tdatetime1 := now.Format("2006-01-") + datetime + now.Format(time_fmt)
        tm1, ok := time.Parse(tfmt1, tdatetime1)
        if ok == nil {
            return tm1
        }

        tfmt2 := tfmt + " Jan 2006" + time_fmt
        tdatetime2 := datetime + now.Format(" Jan 2006" + time_fmt)
        tm2, ok := time.Parse(tfmt2, tdatetime2)
        if ok == nil {
            return tm2
        }
    }

    // Magic Day of week stuff.. Same as 5 above, except work out date from Day
    // 5. Parsing the dayname/time/loc work out the date.
    fmts = []string{time.Kitchen + " MST", "15:04:05 MST", "15:04 MST",
                    time.Kitchen + "Z07:00", "15:04:05Z07:00", "15:04Z07:00",
                    time.Kitchen + " -0700", "15:04:05 -0700", "15:04 -0700"}
    for _, tfmt := range fmts {
        tfmt = "2006-01-02 Monday " + tfmt
        tdatetime := now.Format("2006-01-02 ") + datetime
        tm, ok := time.Parse(tfmt, tdatetime)
        if ok == nil {
            return calc_weekday(datetime, tm)
        }
    }

    // Magic Day of week stuff.. Same as 6 above, except work out date from Day
    // 6. Inherit the loc, parsing the dayname/time work out the date.
    fmts = []string{time.Kitchen, "15:04:05", "15:04"}
    for _, tfmt := range fmts {
        tfmt = "2006-01-02 Monday " + tfmt + " MST"
        tdatetime := now.Format("2006-01-02 ") + datetime + now.Format(" MST")
        tm, ok := time.Parse(tfmt, tdatetime)
        if ok == nil {
            return calc_weekday(datetime, tm)
        }
    }

    // 7. Inherit the loc/time, parsing the dayname work out the date.
    fmts = []string{"Monday"}
    for _, tfmt := range fmts {
        _, ok := time.Parse(tfmt, datetime)
        if ok == nil {
            return calc_weekday(datetime, now)
        }
    }

    fmt.Fprintln(os.Stderr, "Error: Couldn't parse time string:", datetime)
    return time.Now()
}

func hdr(val string, title string, width int) string {
    dashes := width - len(title)
    dashes -= 2
    if dashes < 2 {
        dashes = 2
    }

    beg_dashes := dashes / 2
    end_dashes := dashes - beg_dashes
    return fmt.Sprintf("%s %s %s",
                       strings.Repeat(val, beg_dashes), title,
                       strings.Repeat(val, end_dashes))
}

var ampm_output_flag bool
var color_output_flag bool
var dec_output_flag bool
var debug_flag bool
var isoweek_output_flag bool
var sec_output_flag bool
var terminal_width int

func term_width() int {
    width, _, err := terminal.GetSize(int(os.Stdout.Fd()))
    if err != nil {
        width = 80
    }

    return width
}

func _otime(now time.Time, tm time.Time, tz string, lastday int) int {
    nowtz, _ := now.Zone()

    width := terminal_width
    if width <= 0 {
        width = term_width()
    }

    // Print headers ... 
    if lastday == -1 {
        title := fmt.Sprintf("Day: %s", tm.Format("Monday"))
        if isoweek_output_flag {
            isoyear, isoweek := tm.ISOWeek()
            title += fmt.Sprintf(", Week: %04d/%02d", isoyear, isoweek)
        }
        fmt.Println(hdr("=", title, width))
        lastday = tm.YearDay()
    }
    if lastday != tm.YearDay() {
        title := fmt.Sprintf("New Day: %s", tm.Format("Monday"))
        if isoweek_output_flag {
            isoyear, isoweek := tm.ISOWeek()
            title += fmt.Sprintf(", Week: %04d/%02d", isoyear, isoweek)
        }
        fmt.Println(hdr("-", title, width))
        lastday = tm.YearDay()
    }

    // Setup colours ...
    hours_fmt := fmt.Sprintf
    if color_output_flag {
        weekend := map[time.Weekday]bool{time.Sunday : true, time.Saturday : true}
        if !weekend[tm.Weekday()] {
            if tm.Hour() >= 8 && tm.Hour() < 18 {
                hours_fmt = color.New(color.FgGreen).Sprintf
            } else {
                hours_fmt = color.New(color.FgBlue).Sprintf
            }
        } else {
            if tm.Hour() >= 8 && tm.Hour() < 18 {
                hours_fmt = color.New(color.FgYellow).Sprintf
            } else {
                hours_fmt = color.New(color.FgWhite).Sprintf
            }
        }
    }

    /// Print datetime
    if dec_output_flag {
        clock := hours_fmt(dectime(tm, sec_output_flag))
        fmt.Printf("%s %s %-4s", tm.Format("2006"), clock, tm.Format("MST"))
    } else {
        clock := ""
        if ampm_output_flag {
            if sec_output_flag {
                clock = hours_fmt(tm.Format("03:04:05PM"))
            } else {
                clock = hours_fmt(tm.Format("03:04PM"))
            }
        } else {
            if sec_output_flag {
                clock = hours_fmt(tm.Format("15:04:05"))
            } else {
                clock = hours_fmt(tm.Format("15:04"))
            }
        }
        // FIXME: 4 needs to be cald too, really.
        fmt.Printf("%s %s %-4s", tm.Format("2006-01-02"), clock, tm.Format("MST"))
    }

    // Print user zoneinfo, and marker if it's "local"
    // FIXME: Needs to be 18 calcd.
    if tmtz, _ := tm.Zone(); tmtz == nowtz {
        if color_output_flag {
            fmt.Printf(" --> %s <--\n", color.New(color.FgRed).Sprintf("%-18s", tz))
        } else {
            fmt.Printf(" --> %-18s <--\n", tz)
        }
    } else {
        fmt.Printf("     %-18s\n", tz)
    }

    return lastday
}


type TMInfo struct {
    TM time.Time
    TZ string
}

type ByTime []TMInfo

func (a ByTime) Len() int           { return len(a) }
func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func cmp(tmi1, tmi2 TMInfo) int {
    tm1 := tmi1.TM
    tm2 := tmi2.TM

    if tm1.Before(tm2) { return -1 }
    if tm1.After(tm2)  { return  1 }

    // We are dealing with the "same" time here, so we sort based on what the user sees.
    // Note: This always happens in this program :).

    if tm1.Year() < tm2.Year() { return -1 }
    if tm1.Year() > tm2.Year() { return  1 }

    if tm1.Month() < tm2.Month() { return -1 }
    if tm1.Month() > tm2.Month() { return  1 }

    if tm1.Day() < tm2.Day() { return -1 }
    if tm1.Day() > tm2.Day() { return  1 }

    if tm1.Hour() < tm2.Hour() { return -1 }
    if tm1.Hour() > tm2.Hour() { return  1 }

    if tm1.Minute() < tm2.Minute() { return -1 }
    if tm1.Minute() > tm2.Minute() { return  1 }

    if tm1.Second() < tm2.Second() { return -1 }
    if tm1.Second() > tm2.Second() { return  1 }

    if tmi1.TZ < tmi2.TZ { return -1 }
    if tmi1.TZ > tmi2.TZ { return  1 }

    return 0
}

func (a ByTime) Less(i, j int) bool {
    return cmp(a[i], a[j]) < 0
}

func alltime(now time.Time, datetime time.Time, duration time.Duration, tzs []string) {
    var tms []TMInfo

    for _, tz := range tzs {
        loc, ok := time.LoadLocation(tz)
        if ok != ok {
            if debug_flag {
                fmt.Fprintln(os.Stderr, "JDBG:", "Can't find:", tz)
            }
            continue
        }

        tm := datetime.In(loc)
        tm = tm.Add(duration)

        tms = append(tms, TMInfo{tm, tz})
    }

    sort.Sort(ByTime(tms))

    lastday := -1
    for _, tm := range tms {
        lastday = _otime(now, tm.TM, tm.TZ, lastday)
    }
}

func deftime(now time.Time, datetime time.Time, duration time.Duration) {
    tzs := []string{"Asia/Calcutta",
                    "Asia/Singapore",
                    "Asia/Hong_Kong",
                    "Asia/Tokyo",
                    "Australia/Brisbane",
                    "Europe/London",
                    "Europe/Paris",
                    "Europe/Berlin",
                    "US/Eastern",
                    "US/Pacific",
                    "UTC"}

    alltime(now, datetime, duration, tzs)
}

func init() {
    flag.BoolVar(&ampm_output_flag, "12h", false, "Use AM/PM instead of 24hr output")
    flag.BoolVar(&color_output_flag, "color", true, "Use color output")
    flag.BoolVar(&dec_output_flag, "decimal-time", false, "Use decimal time output")
    flag.BoolVar(&debug_flag, "debug", false, "Print debugging output")
    flag.BoolVar(&isoweek_output_flag, "week", false, "Hdr includes week")
    flag.BoolVar(&sec_output_flag, "seconds", false, "Show seconds in output")
    flag.IntVar(&terminal_width, "width", 0, "Size of terminal (default: auto)")
}

func cotime(local *bool, short *string, pduration *time.Duration,
            now time.Time, datetime time.Time) {
    if !*local && *short == "" { // Print our default list
        deftime(now, datetime, *pduration)
        return
    }

    // Print a user specified list, and UTC:
    tzs := []string{"UTC"}

    if *local {
        tzs = append(tzs, "Local")
    }
    if *short != "" {
        for _, data := range strings.Split(*short, ",") {
            tzs = append(tzs, data)
        }
    }

    alltime(now, datetime, *pduration, tzs)
}

func main() {
    pduration := flag.Duration("duration", 0, "Add a duration to the given date")
    short := flag.String("short", "", "Just print a single zone and UTC")
    local := flag.Bool("local", false, "Just print the local zone and UTC")

    flag.Parse()

    now := time.Now()
    if debug_flag {
        fmt.Fprintln(os.Stderr, "JDBG:", "now:", now, now.Location().String())
    }

    args := flag.Args()
    if len(args) < 1 {
        cotime(local, short, pduration, now, now)
    }

    for _, arg := range args {
        datetime := time.Now()
        if arg == "now" {
        } else if arg == "tomorrow" {
            d, _ := time.ParseDuration("24h")
            datetime = datetime.Add(d)
        } else if arg == "yesterday" {
            d, _ := time.ParseDuration("-24h")
            datetime = datetime.Add(d)
        } else {
            datetime = ptime(now, arg)
            if debug_flag {
                fmt.Fprintln(os.Stderr, "JDBG: datetime:", datetime)
            }
        }

        cotime(local, short, pduration, now, datetime)
    }
}
