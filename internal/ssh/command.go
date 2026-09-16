package ssh

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"mytunnel/internal/config"
)

func Args(c config.Connection) []string {
	args := []string{"ssh", "-p", strconv.Itoa(c.SSHPort())}
	if c.Compression {
		args = append(args, "-C")
	}
	if c.ForwardX {
		args = append(args, "-X")
	}
	if c.ForwardAgent {
		args = append(args, "-A")
	}
	if c.GatewayPorts {
		args = append(args, "-g")
	}
	if custom := strings.TrimSpace(c.Custom); custom != "" {
		args = append(args, splitArgs(custom)...)
	}
	args = append(args, c.Target())
	return args
}

func CommandLine(c config.Connection) string {
	line := strings.Join(Args(c), " ")
	if p := strings.TrimSpace(c.Prepend); p != "" {
		line = p + " " + line
	}
	if a := strings.TrimSpace(c.Append); a != "" {
		line = line + " " + a
	}
	return line
}

func Command(c config.Connection) *exec.Cmd {
	args := Args(c)
	prepend := strings.TrimSpace(c.Prepend)
	appendCmd := strings.TrimSpace(c.Append)
	if prepend != "" || appendCmd != "" {
		return exec.Command("sh", "-c", CommandLine(c))
	}
	return exec.Command(args[0], args[1:]...)
}

func CopyCommand(c config.Connection) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(CommandLine(c))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return nil
}

func ReapConnection(c config.Connection) {
	for _, port := range LocalPorts(c.Custom) {
		pid, err := listenPID(port)
		if err != nil || !isSSH(pid) {
			continue
		}
		killProcess(pid)
	}
}

func ReapAll(cfg *config.Config) {
	if cfg == nil {
		return
	}
	for _, c := range cfg.Connections {
		ReapConnection(c)
	}
}

func LocalPorts(custom string) []int {
	fields := splitArgs(custom)
	var ports []int
	for i := 0; i < len(fields); i++ {
		spec := fields[i]
		if spec == "-L" && i+1 < len(fields) {
			i++
			spec = fields[i]
		} else if strings.HasPrefix(spec, "-L") && len(spec) > 2 {
			spec = spec[2:]
		} else {
			continue
		}
		if p, ok := parseLocalPort(spec); ok {
			ports = append(ports, p)
		}
	}
	return ports
}

func parseLocalPort(spec string) (int, bool) {
	parts := strings.Split(spec, ":")
	if len(parts) < 2 {
		return 0, false
	}
	idx := 0
	if len(parts) == 4 {
		idx = 1
	}
	p, err := strconv.Atoi(parts[idx])
	if err != nil || p <= 0 {
		return 0, false
	}
	return p, true
}

func splitArgs(s string) []string {
	var (
		out    []string
		cur    strings.Builder
		quote  rune
		escape bool
	)
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		out = append(out, cur.String())
		cur.Reset()
	}
	for _, r := range s {
		switch {
		case escape:
			cur.WriteRune(r)
			escape = false
		case r == '\\':
			escape = true
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
		case r == ' ' || r == '\t':
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	flush()
	return out
}

func listenPID(port int) (int, error) {
	cmd := exec.Command("lsof", "-nP", "-t", "-sTCP:LISTEN", fmt.Sprintf("-iTCP:%d", port))
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return 0, fmt.Errorf("nothing listening")
	}
	pid, err := strconv.Atoi(fields[0])
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func isSSH(pid int) bool {
	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=").Output()
	if err != nil {
		return false
	}
	name := strings.TrimSpace(string(out))
	return name == "ssh" || strings.HasSuffix(name, "/ssh")
}

func killProcess(pid int) {
	if pid <= 0 {
		return
	}
	_ = syscall.Kill(pid, syscall.SIGTERM)
	_ = syscall.Kill(-pid, syscall.SIGTERM)
}
