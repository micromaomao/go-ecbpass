package main

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var options = struct {
	algorithm      string
	urls           []string
	putToClipboard bool
	password       []byte
}{"scrypt", make([]string, 0, 1), false, nil}

var xclipPath string = ""

// Parse arguments and do stuff.
func main() {
	args := os.Args
	stdoutIsTty := terminal.IsTerminal(syscall.Stdin)
	options.putToClipboard = stdoutIsTty
	dashdash := false
	for i := 1; i < len(args); i++ {
		c := args[i]

		if dashdash {
			options.urls = append(options.urls, c)
			continue
		}

		if c == "-a" {
			i++
			if i == len(args) {
				argerror("Expected argument after -a")
				return
			}
			options.algorithm = args[i]
			continue
		}

		const pf = "--clipboard="
		if strings.HasPrefix(c, pf) {
			switch val := c[len(pf):]; val {
			case "always":
				options.putToClipboard = true
			case "no":
				options.putToClipboard = false
			case "auto":
				options.putToClipboard = stdoutIsTty
			default:
				argerror(fmt.Sprintf("Invalid value %v for --clipboard. Expected always/no/auto.\n", val))
				return
			}
			continue
		}

		if c == "-p" {
			i++
			if i == len(args) {
				argerror("Expected argument after -p")
				return
			}
			options.password = []byte(args[i])
			continue
		}

		if c == "--help" {
			fmt.Fprint(os.Stdout, "Usage: go-ecbpass [-a <pbkdf2|scrypt>] [--clipboard=<always|no|auto>] [-p <password>] [urls..]\n"+
				"For more info, see man go-ecbpass\n")
			os.Exit(0)
		}

		if c == "--" {
			dashdash = true
			continue
		}

		if strings.HasPrefix(c, "-") {
			if strings.HasPrefix(c, "--") {
				argerror("Unknown option " + c[2:])
				return
			}
			argerror("Unknown option " + c[1:])
			return
		}

		options.urls = append(options.urls, c)
	}

	if options.putToClipboard && len(options.urls) > 1 {
		options.putToClipboard = false
		fmt.Fprint(os.Stderr, "Generated password will not be put to clipboard, because multiplt urls are provided on the command line.\n")
		fmt.Fprint(os.Stderr, "Specify --clipboard=no to slience this warning.\n")
	}

	if options.putToClipboard {
		_xclipPath, err := exec.LookPath("xclip")
		xclipPath = _xclipPath
		if err != nil {
			xclipPath = ""
			options.putToClipboard = false
			if options.password == nil {
				fmt.Fprint(os.Stderr, "xclip command not found. Unable to put stuff to clipboard. \033[1;31mGenerated password will be printed out\033[0m.\n")
			} else {
				fmt.Fprint(os.Stderr, "xclip command not found. Unable to put stuff to clipboard. Generated password will be printed out.\n")
			}
		}
	}

	if options.password == nil {
		if !stdoutIsTty {
			fmt.Fprintf(os.Stderr, "Password need to be provided on the command line, because stdin is not a terminal.\n")
			os.Exit(1)
			return
		}
		fmt.Fprintf(os.Stderr, "[go-ecbpass] master password: ")
		os.Stderr.Sync()
		pw, err := terminal.ReadPassword(syscall.Stdin)
		os.Stderr.WriteString("\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read password: %v.\n", err.Error())
			os.Exit(1)
			return
		}
		options.password = pw
		hint := Hashhint(pw)
		fmt.Fprintf(os.Stderr, "  \033[32mPassword hint: %v\033[0m\n", hint)
	}

	if len(options.urls) == 0 {
		fmt.Fprintf(os.Stderr, "Enter url.\n")
		for {
			url, err, eof := readLine(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v", err.Error())
				os.Exit(1)
				return
			}
			if url != "" {
				doUrl(url)
			}
			if eof {
				return
			}
		}
	} else {
		for _, url := range options.urls {
			doUrl(url)
		}
	}
}

func argerror(err string) {
	fmt.Fprintf(os.Stderr, "go-ecbpass: invalid argument: %v.\n", err)
	os.Exit(1)
}

// Read a entire line, and no more, from `from`, and convert the result into string.
// If the stream ended without newline, the string until it ends will be returned.
func readLine(from *os.File) (line string, err error, eof bool) {
	var stepSize int = 20
	buf := make([]byte, stepSize)
	var off int = 0
	for {
		readbuf := buf[off:]
		n, err := from.Read(readbuf)
		if err != nil && err != io.EOF {
			return "", err, false
		}
		if n < len(readbuf) || (n == len(readbuf) && readbuf[n-1] == '\n') {
			return strings.TrimRight(string(buf[0:off+n]), "\n"), nil, err == io.EOF
		} else {
			off += stepSize
			newBuf := make([]byte, len(buf)+stepSize)
			copy(newBuf, buf)
			buf = newBuf
		}
	}
}

func doUrl(url string) {
	salt, err := UrlToSalt(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing url: %v... Using the provided string verbatim as salt.\n", err.Error())
		salt = []byte(url)
	}
	var cryptFunc (func(password []byte, salt []byte) (result string)) = nil
	switch options.algorithm {
	case "pbkdf2":
		cryptFunc = PBKDF2
	case "scrypt":
		cryptFunc = Scrypt
	default:
		fmt.Fprintf(os.Stderr, "Unknow algorithm %v.\n", options.algorithm)
		os.Exit(1)
		return
	}
	if len(options.urls) == 0 {
		// interactive
		fmt.Fprint(os.Stderr, "Calculating, please wait...")
	}
	result := cryptFunc(options.password, salt)
	fmt.Fprint(os.Stderr, "\033[2K\r")
	if options.putToClipboard {
		err := putStringToClipboard(result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to run xclip: %v.", err.Error())
			os.Exit(1)
		} else {
			fmt.Fprint(os.Stderr, "Password copied.\n")
		}
	} else {
		os.Stdout.WriteString(result + "\n")
	}
}

func putStringToClipboard(str string) (err error) {
	cmd := exec.Command(xclipPath, "-selection", "clipboard", "-in")
	cmd.Stdout = nil
	cmd.Stderr = nil
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	cmd.Start()
	_, err = io.WriteString(stdin, str)
	closeErr := stdin.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return cmd.Wait()
}
