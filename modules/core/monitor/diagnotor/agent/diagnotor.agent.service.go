package diagnotor

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/erda-project/erda-proto-go/core/monitor/diagnotor/pb"
	"github.com/erda-project/erda/pkg/common/errors"
	"github.com/shirou/gopsutil/process"
)

type diagnotorAgentService struct {
	p *provider

	lock       sync.Mutex
	lastIOStat map[int32]*ioCountersStatEntry
}

type ioCountersStatEntry struct {
	stat     process.IOCountersStat
	lastTime time.Time
}

func (s *diagnotorAgentService) ListTargetProcesses(ctx context.Context, req *pb.ListTargetProcessesRequest) (*pb.ListTargetProcessesResponse, error) {
	lastIOStat := make(map[int32]*ioCountersStatEntry)
	status := &pb.HostProcessStatus{}
	procs, err := process.Processes()
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}

	for _, proc := range procs {
		ps := &pb.Process{
			Pid: strconv.Itoa(int(proc.Pid)),
		}
		ps.Name, _ = proc.Name()
		ps.Cmdline, _ = proc.Cmdline()
		// user
		user, _ := proc.Username()
		if len(user) <= 0 {
			uids, _ := proc.Uids()
			if len(uids) > 0 {
				user = strconv.Itoa(int(uids[0]))
			}
		}
		ps.User = user
		ps.CreateTime, _ = proc.CreateTime()
		numFDs, _ := proc.NumFDs()
		ps.NumFDs = int64(numFDs)
		numThreads, _ := proc.NumThreads()
		ps.NumThreads = int64(numThreads)
		// app type
		ps.AppType = getAppType(ps)

		// io
		now := time.Now()
		io, err := proc.IOCounters()
		if err == nil {
			ps.Io = &pb.ProcessIOStatus{
				ReadCount:  int64(io.ReadCount),
				ReadBytes:  int64(io.ReadBytes),
				WriteCount: int64(io.WriteCount),
				WriteBytes: int64(io.WriteBytes),
			}
			lastIOStat[proc.Pid] = &ioCountersStatEntry{
				stat:     *io,
				lastTime: now,
			}
			s.lock.Lock()
			if len(s.lastIOStat) > 0 {
				if lastStat, ok := s.lastIOStat[proc.Pid]; ok {
					seconds := now.Sub(lastStat.lastTime).Seconds()
					if seconds > 0 {
						ps.Io.ReadRate = float64(io.ReadBytes-lastStat.stat.ReadBytes) / seconds
						ps.Io.WriteRate = float64(io.WriteBytes-lastStat.stat.WriteBytes) / seconds
					}
				}
			}
			s.lock.Unlock()
		}

		// cpu
		cpuTimes, err := proc.Times()
		if err == nil {
			ps.Cpu = &pb.ProcessCPUStatus{
				User:      cpuTimes.User,
				System:    cpuTimes.System,
				Idle:      cpuTimes.Idle,
				Nice:      cpuTimes.Nice,
				IoWait:    cpuTimes.Iowait,
				Irq:       cpuTimes.Irq,
				SoftIrq:   cpuTimes.Softirq,
				Steal:     cpuTimes.Steal,
				Guest:     cpuTimes.Guest,
				GuestNice: cpuTimes.GuestNice,
			}
		}
		cpuPercent, err := proc.Percent(time.Duration(0))
		if err == nil {
			if ps.Cpu == nil {
				ps.Cpu = &pb.ProcessCPUStatus{}
			}
			ps.Cpu.Usage = cpuPercent
		}

		// memory
		mem, err := proc.MemoryInfo()
		if err == nil {
			ps.Memory = &pb.ProcessMemoryStatus{
				Rss:    int64(mem.RSS),
				Vms:    int64(mem.VMS),
				Swap:   int64(mem.Swap),
				Data:   int64(mem.Data),
				Stack:  int64(mem.Stack),
				Locked: int64(mem.Locked),
			}

		}
		memPerc, err := proc.MemoryPercent()
		if err == nil {
			if ps.Memory == nil {
				ps.Memory = &pb.ProcessMemoryStatus{}
			}
			ps.Memory.Usage = float64(memPerc)
		}

		// rlimit
		rlims, err := proc.RlimitUsage(true)
		if err == nil {
			ps.Rlimit = &pb.ProcessRLimit{}
			for _, rlim := range rlims {
				stat := &pb.ProcessRLimitStatus{
					Soft: int64(rlim.Soft),
					Hard: int64(rlim.Hard),
					Used: int64(rlim.Used),
				}
				switch rlim.Resource {
				case process.RLIMIT_CPU:
					ps.Rlimit.CpuTime = stat
				case process.RLIMIT_DATA:
					ps.Rlimit.MemoryData = stat
				case process.RLIMIT_STACK:
					ps.Rlimit.MemoryStack = stat
				case process.RLIMIT_RSS:
					ps.Rlimit.MemoryRss = stat
				case process.RLIMIT_NOFILE:
					ps.Rlimit.NumFDs = stat
				case process.RLIMIT_MEMLOCK:
					ps.Rlimit.MemoryLocked = stat
				case process.RLIMIT_AS:
					ps.Rlimit.MemoryVms = stat
				case process.RLIMIT_LOCKS:
					stat.Used = -1
					ps.Rlimit.FileLocks = stat
				case process.RLIMIT_SIGPENDING:
					ps.Rlimit.SignalsPending = stat
				case process.RLIMIT_NICE:
					ps.Rlimit.NicePriority = stat
				case process.RLIMIT_RTPRIO:
					ps.Rlimit.RealtimePriority = stat
				}
			}
		}

		// context switches
		ctxSwitches, err := proc.NumCtxSwitches()
		if err == nil {
			ps.ContextSwitches = &pb.ProcessContextSwitches{
				Voluntary:   ctxSwitches.Voluntary,
				Involuntary: ctxSwitches.Involuntary,
			}
		}

		status.Processes = append(status.Processes, ps)
	}

	// save last io status
	s.lock.Lock()
	s.lastIOStat = lastIOStat
	s.lock.Unlock()

	return &pb.ListTargetProcessesResponse{
		Data: status,
	}, nil
}

func getAppType(proc *pb.Process) string {
	switch proc.Name {
	case "java":
		return "java"
	case "sh", "bash", "zsh":
		return "shell"
	}
	return "unknown"
}
