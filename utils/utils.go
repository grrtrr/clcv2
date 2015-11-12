package utils

import (
	"github.com/olekukonko/tablewriter"
	"reflect"
	"regexp"
	"time"
	"fmt"
	"os"
)

/* CLC Server Syntax. FIXME: possibly subject to change without notice. */
var serverRegexp = regexp.MustCompile(`(^[A-Z]{2}\d)[A-Z0-9-]{4,}$`)

/* Parse time zone offset(supported formats: -07:00:00, -7:00, -700, -0700, +00:00, 100) */
var tzRegexp = regexp.MustCompile(`^\s*([+-]?)(\d{1,2}):?(\d{2})(:?(\d{2}))?\s*$`)


// Return true if @s looks like a CLC server name
func LooksLikeServerName(s string) bool {
	return serverRegexp.MatchString(s)
}

// Extract the Location prefix from @serverName
func ExtractLocationFromServerName(serverName string) string {
	if m := serverRegexp.FindStringSubmatch(serverName); m != nil {
		return m[1]
	}
	return ""
}

// Parse time zone offset string @o
func ParseTimeZoneOffset(o string) (d time.Duration, err error) {
	if m := tzRegexp.FindStringSubmatch(o); m == nil {
		err = fmt.Errorf("Invalid time zone offset format %q", o)
	} else {
		s := fmt.Sprintf("%s%sh%sm", m[1], m[2], m[3])
		if m[5] == "" {
			s += fmt.Sprintf("%ss", m[4])
		} else {
			s += fmt.Sprintf("%ss", m[5])
		}
		d, err = time. ParseDuration(s)
	}
	return
}

// Print (pointer) to struct as table, using key/type/value
func PrintStruct(in interface{}) {
	t := reflect.TypeOf(in)
	v := reflect.ValueOf(in)

	if in == nil {
		fmt.Println("nil")
		return
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = reflect.Indirect(v)
	}

	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("Expected a struct, got %T", in))
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)

	table.SetHeader([]string{ t.Name(), "Type", "Value" } )
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		table.Append([]string{ f.Name, f.Type.Name(), fmt.Sprintf("%v", v.Field(i)) })
	}
	table.Render()
}
