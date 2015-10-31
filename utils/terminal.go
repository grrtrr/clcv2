/*
 * Low-level terminal functions
 */
package utils

import (
	"golang.org/x/crypto/ssh/terminal"
	"os/signal"
	"syscall"
	"strings"
	"errors"
	"bytes"
	"time"
	"fmt"
	"os"
)

/* Get the terminal descriptor for stdin */
func getTerminalFd() (fd int, err error) {
	fd = int(os.Stdin.Fd())

	if !terminal.IsTerminal(fd) {
		err = fmt.Errorf("Input is not from a terminal")
	}
	return
}

/* Read non-empty password from terminal */
func GetPass(prompt string) (pass string, err error) {
	var resp []byte

	fd, err := getTerminalFd()
	if err != nil {
		return
	}

	/* Store current terminal state in case the call gets interrupted by signal */
	oldState, err := terminal.GetState(fd)
	if err != nil {
		err = fmt.Errorf("Failed to get terminal state: %s\n", err)
		return
	}

	/*
         * Install signal handler
         * Unlike the ReadLine function, using a raw terminal state here does not help.
         * If the prompt gets killed by a signal, the terminal state is not restored.
         * Hence restore it in a signal handler
         */
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGINT)
	go func() {
		<-c
		terminal.Restore(fd, oldState)
		fmt.Fprintln(os.Stderr, "aborting")
		os.Exit(1)
	}()


	for i := 0; len(resp) == 0; i++ {
		if i > 0 {
			fmt.Printf("\rInvalid response - try again")
			time.Sleep(500 * time.Millisecond)
		}
		/* Clear line - see https://en.wikipedia.org/wiki/ANSI_escape_code */
		fmt.Printf("\r\033[2K%s: ", prompt)

		/* This function internally takes care of restoring terminal state. */
		resp, err = terminal.ReadPassword(fd)
		if err != nil {
			return
		}
		resp = bytes.TrimSpace(resp)
	}
	/* Restore signal handling */
	signal.Stop(c)
	signal.Reset(syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGINT)
	return string(resp), nil
}

/*
 * Prompt for a value, allowing defaults and limiting choices:
 * a) prompt <msg>                     - prompt for a non-empty answer
 * b) prompt <msg> <default>           - prompt using @default in case of empty answer
 * c) prompt <msg> <v1> <v2> [.. <vn>] - allow only v1 .. vn as valid answer
 *    Special Case: if <vn> repeats an element v1 .. v<n-1>, it is used as a default
 *    The case of using a single repeated element twice is the same as (b).
 */
func PromptInput(msg string, choices ...string) (resp string, err error) {
	var defval  string
	var allowed []string = choices[:]

	switch len(allowed) {
	case 0:	/* Do nothing */
	case 1: defval = allowed[0]
		msg = fmt.Sprintf("%s [%s]", msg, defval)
	default:
		if strings.Index(strings.Join(allowed[:len(allowed)-1], " "),
			         allowed[len(allowed)-1]) != -1 {
			/* Choice with default present */
			defval  = allowed[len(allowed)-1]
			allowed = allowed[:len(allowed)-1]
			for i, e := range allowed {
				if e == defval {
					allowed[i] = fmt.Sprintf("*%s*", e)
				}
			}
		}
		msg = fmt.Sprintf("%s [%s]", msg, strings.Join(allowed, "/"))
	}

	fd, err := getTerminalFd()
	if err != nil {
		return
	}

	oldState, err := terminal.MakeRaw(fd)
	if err != nil {
		err = fmt.Errorf("Failed to set terminal into raw mode: %s", err)
		return
	}
	defer terminal.Restore(fd, oldState)

	/* Save Cursor position */
	fmt.Print("\r\033[s")

	/* Prompt returns to saved cursor (\033[u) and clears to end of line (\033[2K) */
	term := terminal.NewTerminal(os.Stdin, fmt.Sprintf("\033[u\033[2K%s: ", msg))
	if term == nil {
		panic(errors.New("could not create terminal"))
	}

	for i := 0; len(resp) == 0; i++ {
		if i > 0 {
			fmt.Printf("\033[u\033[2KInvalid response - try again")
			time.Sleep(500 * time.Millisecond)
		}

		resp, err = term.ReadLine()
		if err != nil { /* e.g. io.EOF */
			return
		}
		resp = strings.TrimSpace(resp)
		if resp == "" && defval != "" {
			resp = defval
		} else if resp != "" && len(allowed) > 1 {
			if !isArrayMember(resp, allowed) {
				resp = ""
			}
		}
	}
	return
}

/* Return true if @e is contained in @stringArray */
func isArrayMember(e string, stringArray []string) bool {
	for _, c := range stringArray {
		if e == c {
			return true
		}
	}
	return false
}
