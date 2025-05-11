package system

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"golang.org/x/text/encoding/simplifiedchinese"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

func initOS() (o Os) {
	o.GOOS = runtime.GOOS
	o.NumCPU = runtime.NumCPU()
	o.Compiler = runtime.Compiler
	o.GoVersion = runtime.Version()
	o.NumGoroutine = runtime.NumGoroutine()
	return o
}

func initCPU() (c Cpu, err error) {
	if cores, err := cpu.Counts(false); err != nil {
		return c, err
	} else {
		c.Cores = cores
	}
	if cpus, err := cpu.Percent(time.Duration(200)*time.Millisecond, true); err != nil {
		return c, err
	} else {
		c.Cpus = cpus
	}
	return c, nil
}

func initRAM() (r Rrm, err error) {
	if u, err := mem.VirtualMemory(); err != nil {
		return r, err
	} else {
		r.UsedMB = int(u.Used) / MB
		r.TotalMB = int(u.Total) / MB
		r.UsedPercent = int(u.UsedPercent)
	}
	return r, nil
}

func initDisk() (d Disk, err error) {
	if u, err := disk.Usage("/"); err != nil {
		return d, err
	} else {
		d.UsedMB = int(u.Used) / MB
		d.UsedGB = int(u.Used) / GB
		d.TotalMB = int(u.Total) / MB
		d.TotalGB = int(u.Total) / GB
		d.UsedPercent = int(u.UsedPercent)
	}
	return d, nil
}

func byteToString(data []byte, charset string) string {
	var result string

	switch charset {
	case "GBK":
		decodeBytes, _ := simplifiedchinese.GBK.NewDecoder().Bytes(data)
		result = string(decodeBytes)
	case "GB18030":
		decodeBytes, _ := simplifiedchinese.GB18030.NewDecoder().Bytes(data)
		result = string(decodeBytes)
	case "UTF8":
		fallthrough
	default:
		result = string(data)
	}

	return result
}

func parseProcessInfo(output []byte, pid int) (*ProcessInfo, error) {
	lines := strings.Split(string(output), "\n")

	if len(lines) < 2 {
		return nil, fmt.Errorf("no process found with PID %d", pid)
	}

	var processInfo ProcessInfo
	if runtime.GOOS == "windows" {
		fields := strings.Split(lines[1], "\",\"")
		if len(fields) < 9 {
			return nil, fmt.Errorf("unexpected tasklist output format")
		}

		processInfo = ProcessInfo{
			PID:    pid,
			CPU:    "N/A",
			Memory: fields[4], // Memory usage in K
			State:  fields[5],
			User:   "N/A",
			Cmd:    fields[8],
		}
	} else {
		fields := strings.Fields(lines[1])
		if len(fields) < 6 {
			return nil, fmt.Errorf("unexpected ps output format")
		}

		processInfo = ProcessInfo{
			PID:    pid,
			CPU:    fields[1],
			Memory: fields[2],
			State:  fields[3],
			User:   fields[4],
			Cmd:    fields[5],
		}
	}

	return &processInfo, nil
}

func getThreadsInfo(pid int) ([]string, error) {
	cmd := exec.Command("ps", "-T", "-p", strconv.Itoa(pid))
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")

	var threads []string
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) != "" {
			threads = append(threads, line)
		}
	}

	return threads, nil
}

func getIOStats(pid int) (string, error) {
	filePath := fmt.Sprintf("/proc/%d/io", pid)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func getProcessStartTime(pid int) (string, error) {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "lstart=")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func getParentProcess(pid int) (int, error) {
	cmd := exec.Command("ps", "-o", "ppid=", "-p", strconv.Itoa(pid))
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	ppid, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, err
	}

	return ppid, nil
}

func getNetworkConnections(pid int) (string, error) {
	cmd := exec.Command("lsof", "-p", strconv.Itoa(pid), "-i")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
