package system

import (
	"bytes"
	"fmt"
	"github.com/wind959/ko-utils/validator"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"unicode/utf8"
)

//获取服务器cpu,硬盘内存等信息

type Server struct {
	Os   Os   `json:"os"`
	Cpu  Cpu  `json:"cpu"`
	Rrm  Rrm  `json:"ram"`
	Disk Disk `json:"disk"`
}

type Os struct {
	GOOS         string `json:"goos"`
	NumCPU       int    `json:"numCpu"`
	Compiler     string `json:"compiler"`
	GoVersion    string `json:"goVersion"`
	NumGoroutine int    `json:"numGoroutine"`
}

type Cpu struct {
	Cpus  []float64 `json:"cpus"`
	Cores int       `json:"cores"`
}

type Rrm struct {
	UsedMB      int `json:"usedMb"`
	TotalMB     int `json:"totalMb"`
	UsedPercent int `json:"usedPercent"`
}

type Disk struct {
	UsedMB      int `json:"usedMb"`
	UsedGB      int `json:"usedGb"`
	TotalMB     int `json:"totalMb"`
	TotalGB     int `json:"totalGb"`
	UsedPercent int `json:"usedPercent"`
}

type ProcessInfo struct {
	PID                int
	CPU                string
	Memory             string
	State              string
	User               string
	Cmd                string
	Threads            []string
	IOStats            string
	StartTime          string
	ParentPID          int
	NetworkConnections string
}

type (
	Option func(*exec.Cmd)
)

// GetServerInfo 获取机器详情
func GetServerInfo() (server *Server, err error) {
	var s Server
	s.Os = initOS()
	if s.Cpu, err = initCPU(); err != nil {
		return &s, err
	}
	if s.Rrm, err = initRAM(); err != nil {

		return &s, err
	}
	if s.Disk, err = initDisk(); err != nil {
		return &s, err
	}
	return &s, nil
}

// IsWindows 检查当前操作系统是否是windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsLinux 检查当前操作系统是否是linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsMac 检查当前操作系统是否是macos
func IsMac() bool {
	return runtime.GOOS == "darwin"
}

// SetOsEnv 设置由key命名的环境变量的值
func SetOsEnv(key, value string) error {
	return os.Setenv(key, value)
}

// GetOsEnv 获取key命名的环境变量的值
func GetOsEnv(key string) string {
	return os.Getenv(key)
}

// RemoveOsEnv 删除单个环境变量
func RemoveOsEnv(key string) error {
	return os.Unsetenv(key)
}

// ExecCommand 执行shell命令，返回命令的stdout和stderr字符串，如果出现错误，则返回错误。
// 参数`command`是一个完整的命令字符串，如ls-a（linux），dir（windows），ping 127.0.0.1。
// 在linux中，使用/bin/bash-c执行命令，在windows中，使用powershell.exe执行命令。
// 函数的第二个参数是cmd选项控制参数，类型是func(*exec.Cmd)，可以通过这个参数设置cmd属性。
func ExecCommand(command string, opts ...Option) (stdout, stderr string, err error) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	cmd := exec.Command("/bin/bash", "-c", command)
	if IsWindows() {
		cmd = exec.Command("powershell.exe", command)
	}

	for _, opt := range opts {
		if opt != nil {
			opt(cmd)
		}
	}
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err = cmd.Run()

	if err != nil {
		if utf8.Valid(errOut.Bytes()) {
			stderr = byteToString(errOut.Bytes(), "UTF8")
		} else if validator.IsGBK(errOut.Bytes()) {
			stderr = byteToString(errOut.Bytes(), "GBK")
		}
		return
	}

	data := out.Bytes()
	if utf8.Valid(data) {
		stdout = byteToString(data, "UTF8")
	} else if validator.IsGBK(data) {
		stdout = byteToString(data, "GBK")
	}
	return
}

// GetOsBits 获取当前操作系统位数，返回32或64
func GetOsBits() int {
	return 32 << (^uint(0) >> 63)
}

// StartProcess 创建进程
func StartProcess(command string, args ...string) (int, error) {
	cmd := exec.Command(command, args...)

	if err := cmd.Start(); err != nil {
		return 0, err
	}

	return cmd.Process.Pid, nil
}

// StopProcess 停止进程
func StopProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Signal(os.Kill)
}

// KillProcess 强制停止进程
func KillProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Kill()
}

// GetProcessInfo 根据进程id获取进程信息
func GetProcessInfo(pid int) (*ProcessInfo, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "CSV", "/V")
	} else {
		cmd = exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "pid,%cpu,%mem,state,user,comm")
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	processInfo, err := parseProcessInfo(output, pid)
	if err != nil {
		return nil, err
	}

	if runtime.GOOS != "windows" {
		processInfo.Threads, _ = getThreadsInfo(pid)
		processInfo.IOStats, _ = getIOStats(pid)
		processInfo.StartTime, _ = getProcessStartTime(pid)
		processInfo.ParentPID, _ = getParentProcess(pid)
		processInfo.NetworkConnections, _ = getNetworkConnections(pid)
	}

	return processInfo, nil
}
