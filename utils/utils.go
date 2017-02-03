package utils

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	prompt "github.com/segmentio/go-prompt"
)

var (
	// CLC Server Syntax. FIXME: possibly subject to change without notice.
	serverRegexp = regexp.MustCompile(`(?i)(^[A-Z]{2}\d)[A-Z0-9-]{4,}$`)

	// Parse time zone offset(supported formats: -07:00:00, -7:00, -700, -0700, +00:00, 100)
	tzRegexp = regexp.MustCompile(`^\s*([+-]?)(\d{1,2}):?(\d{2})(:?(\d{2}))?\s*$`)
)

// Return true if @s looks like a CLC server name
func LooksLikeServerName(s string) bool {
	return serverRegexp.MatchString(s)
}

// Extract the Location prefix from @serverName, return in upper-case if found.
func ExtractLocationFromServerName(serverName string) string {
	if m := serverRegexp.FindStringSubmatch(serverName); m != nil {
		return strings.ToUpper(m[1])
	}
	return ""
}

// ResolveUserAndPass supports multiple ways of resolving the username and password
// 1. directly (pass-through),
// 2. command-line flags (g_user, g_pass),
// 3. environment variables (CLC_USERNAME, CLC_PASSWORD),
// 4. prompt for values
func ResolveUserAndPass(userDefault, passDefault string) (username, password string) {
	var promptStr string = "Username"

	if username = userDefault; username == "" {
		username = os.Getenv("CLC_USERNAME")
	}
	if username == "" {
		username = prompt.StringRequired(promptStr)
		promptStr = "Password"
	} else {
		promptStr = fmt.Sprintf("Password for %s", username)
	}

	if password = passDefault; password == "" {
		password = os.Getenv("CLC_PASSWORD")
	}
	if password == "" {
		password = prompt.PasswordMasked(promptStr)
	}
	return username, password
}

// Parse time zone offset string @o
func ParseTimeZoneOffset(o string) (d time.Duration, err error) {
	if m := tzRegexp.FindStringSubmatch(o); m == nil {
		err = errors.Errorf("Invalid time zone offset format %q", o)
	} else {
		s := fmt.Sprintf("%s%sh%sm", m[1], m[2], m[3])
		if m[5] == "" {
			s += fmt.Sprintf("%ss", m[4])
		} else {
			s += fmt.Sprintf("%ss", m[5])
		}
		d, err = time.ParseDuration(s)
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
		panic(errors.Errorf("Expected a struct, got %T", in))
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoWrapText(false)

	table.SetHeader([]string{t.Name(), "Type", "Value"})
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		table.Append([]string{f.Name, f.Type.Name(), fmt.Sprintf("%v", v.Field(i))})
	}
	table.Render()
}
