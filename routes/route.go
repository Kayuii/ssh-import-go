package routes

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/go-cmd/cmd"
	"github.com/kayuii/ssh-import-go/logs"
	"github.com/kayuii/ssh-import-go/pkg/gh"
	"github.com/kayuii/ssh-import-go/pkg/lp"
	"github.com/kayuii/ssh-import-go/usecase"
	"github.com/kayuii/ssh-import-go/utils"
	"github.com/kayuii/ssh-import-go/version"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
)

type Route struct {
	logger    usecase.Logger
	ctx       *cli.Context
	output    string
	useragent string
}

func New(c *cli.Context) *Route {
	var logger usecase.Logger
	var output = ""
	var path = "~/.ssh/authorized_keys"
	var useragent = ""
	if c.IsSet("c") {
		logger = logs.NewZerologLoggerWithColor(os.Stdout)
	} else {
		logger = logs.NewZerologLogger(os.Stdout)
	}
	if c.IsSet("o") {
		path = c.String("o")
	}
	if c.IsSet("u") {
		useragent = c.String("u")
	}
	output, _ = homedir.Expand(path)
	return &Route{
		logger,
		c,
		output,
		useragent,
	}
}

func (r *Route) Exec() error {
	var username = ""
	var proto = usecase.DEFAULT_PROTO
	var errors = []string{}
	var keys = []string{}
	var action = ""

	if r.ctx.Args().Present() {
		for _, userid := range r.ctx.Args().Slice() {
			user_pieces := strings.Split(userid, ":")
			if len(user_pieces) == 2 {
				if !usecase.DEFAULT_PROTO.Match(user_pieces[0]) {
					r.logger.Errorf("ssh-import protocol %s: not found or cannot execute", user_pieces[0])
				}
				proto = usecase.DEFAULT_PROTO.Is(user_pieces[0])
				username = user_pieces[1]
			} else if len(user_pieces) == 1 {
				username = userid
			} else {
				r.logger.Errorf("Invalid user ID: [%s]", userid)
				errors = append(errors, userid)
				continue
			}
			var k = []string{}
			if r.ctx.IsSet("r") {
				k = r.RemoveKeys(proto, username)
				keys = append(keys, k...)
				action = "Removed"
			} else {
				k = r.ImportKeys(proto, username)
				keys = append(keys, k...)
				action = "Authorized"
			}
			if len(k) == 0 {
				errors = append(errors, userid)
			}
		}
		r.logger.Infof("[%d] SSH keys [%s]", len(keys), action)
	} else {
		cli.ShowAppHelpAndExit(r.ctx, 1)
	}

	if len(errors) > 0 {
		r.logger.Errorf("No matching keys found for [%s] ", strings.Join(errors, ","))
	}
	return nil
}

// Build a string that uniquely identifies a key
func (r *Route) fp_tuple(fields []string) string {
	tmp := []string{
		fields[0],
		fields[1],
		fields[len(fields)-1],
		"\n",
	}
	return strings.Join(tmp, " ")
}

// Return a list of uniquely identified keys
func (r *Route) KeyList(fields []string) []string {
	keys := []string{}
	for _, line := range fields {
		ssh_fp := r.KeyFingerprint(strings.Split(line, " "))
		if ssh_fp != "" {
			keys = append(keys, r.fp_tuple(strings.Split(ssh_fp, " ")))
		}
	}
	return keys
}

// Get the fingerprint for an SSH public key Returns None if not valid key material
func (r *Route) KeyFingerprint(fields []string) string {
	if len(fields) == 0 {
		return ""
	}
	if len(fields) < 3 {
		return ""
	}
	tempfd, err := ioutil.TempFile("", "ssh-auth-key-check.*.pub")
	if err != nil {
		r.logger.Error(err.Error())
		os.Exit(1)
	}
	defer func() {
		os.Remove(tempfd.Name())
		tempfd.Close()
	}()

	bw := bufio.NewWriter(tempfd)
	bw.WriteString(strings.Join(fields, " "))
	bw.Flush()

	_cmd := cmd.NewCmd("ssh-keygen", "-l", "-f", tempfd.Name())
	s := <-_cmd.Start()

	return s.Stdout[0]
}

// Call out to a subcommand to handle the specified protocol and username
func (r *Route) FetchKeys(proto usecase.GitProto, username string) []string {

	var output = []string{}

	switch proto {
	case usecase.GIT_GITHUB:
		res, err := gh.FetchKeys(username, r.UserAgent())
		if err != nil {
			r.logger.Error(err.Error())
			os.Exit(1)
		}
		output = append(output, res...)
	case usecase.GIT_LAUNCHPAD:
		res, err := lp.FetchKeys(username, r.UserAgent())
		if err != nil {
			r.logger.Error(err.Error())
			os.Exit(1)
		}
		output = append(output, res...)
	default:
		break
	}
	return output
}

// Import keys from service at 'proto' for 'username', appending to output file
func (r *Route) ImportKeys(proto usecase.GitProto, username string) []string {
	// Map out which keys we already have, so we don't keep appending the same ones
	var local_keys = strings.Join(r.KeyList(r.ReadKeyfile()), "")
	// Protocol handler should output SSH keys, one per line
	var comment_string = fmt.Sprintf("# ssh-import-id %s:%s\n", proto, username)
	var result = []string{}
	var keyfile_lines = []string{}
	for _, line := range r.FetchKeys(proto, username) {
		fields := strings.Split(line, " ")
		fields = append(fields, comment_string)
		ssh_fp := strings.Split(r.KeyFingerprint(fields), " ")
		if strings.Contains(local_keys, r.fp_tuple(ssh_fp)) {
			r.logger.Infof(" Already authorized ['%s' '%s' '%s' '%s']", ssh_fp[0], ssh_fp[1], ssh_fp[2], ssh_fp[len(ssh_fp)-1])
			result = append(result, strings.Join(fields, " "))
		} else {
			keyfile_lines = append(keyfile_lines, strings.Join(fields, " "))
			result = append(result, strings.Join(fields, " "))
			r.logger.Infof(" Authorized key ['%s' '%s' '%s' '%s']", ssh_fp[0], ssh_fp[1], ssh_fp[2], ssh_fp[len(ssh_fp)-1])
		}
	}
	if len(keyfile_lines) > 0 {
		r.WriteKeyfile(keyfile_lines, os.O_WRONLY|os.O_APPEND)
	}
	return result
}

// Remove keys from the output file, if they were inserted by this tool
func (r *Route) RemoveKeys(proto usecase.GitProto, username string) []string {
	// Only remove keys labeled with our comment string
	var comment_string = fmt.Sprintf("# ssh-import-id %s:%s\n", proto, username)
	var update_lines = []string{}
	var removed = []string{}
	for _, line := range r.ReadKeyfile() {
		if strings.HasSuffix(line, comment_string) {
			removed = append(removed, line)
		} else {
			update_lines = append(update_lines, line)
		}
	}
	if len(removed) > 0 {
		r.WriteKeyfile(update_lines, os.O_WRONLY|os.O_TRUNC)
	}
	return removed
}

// Locate key file, read the current state, return lines in a list
func (r *Route) ReadKeyfile() []string {
	var lines = []string{}
	if !utils.IsExist(r.output) {
		r.logger.Errorf("Could not read authorized key file [%s]", r.output)
		os.Exit(1)
	}
	f, err := os.Open(r.output)
	if err != nil {
		r.logger.Error(err.Error())
		os.Exit(1)
	}
	defer f.Close()
	bs := bufio.NewScanner(f)
	bs.Split(bufio.ScanLines)
	for bs.Scan() {
		if len(bs.Bytes()) > 0 {
			lines = append(lines, fmt.Sprintf("%s\n", bs.Text()))
		}
	}
	return lines
}

// Locate key file, write lines to it
func (r *Route) WriteKeyfile(keyfile_lines []string, flag int) {
	// var lines = []string{}
	if !utils.IsExist(r.output) {
		r.logger.Errorf("Could not read authorized key file [%s]", r.output)
		os.Exit(1)
	}
	f, err := os.OpenFile(r.output, flag, os.ModePerm)
	if err != nil {
		r.logger.Error(err.Error())
		os.Exit(1)
	}
	defer f.Close()
	bw := bufio.NewWriter(f)
	for _, line := range keyfile_lines {
		if line != "" {
			bw.WriteString(line)
		}
	}
	bw.Flush()
}

// Construct a useful user agent string
func (r *Route) UserAgent() string {
	ssh_import_id := fmt.Sprintf("ssh-import-id/%s", version.APP_VERSION)
	golang := fmt.Sprintf("golang/%s", version.GOVERSION)
	distro := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	return fmt.Sprintf("%s %s %s %s", ssh_import_id, golang, distro, r.useragent)
}
