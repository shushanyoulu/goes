package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/czxichen/command/watchdog"
	conf "github.com/dlintw/goconf"
)

//+build windows,linux

const logDir = "./watchdog"

func newProc(svc *Service, null, pw *os.File) *os.ProcAttr {
	return &os.ProcAttr{Dir: svc.path, Files: []*os.File{null, pw, pw}}
}

func setPriority(pid, priority uintptr) syscall.Errno {
	return 0
}

var (
	logpath    = flag.String("log_path", "", "Specify log path")
	configFile = flag.String("config", "watchdog.ini", "Watchdog configuration file")
)

func cfgOpt(cfg *conf.ConfigFile, section, option string) string {
	if !cfg.HasOption(section, option) {
		return ""
	}
	s, err := cfg.GetString(section, option)
	if err != nil {
		log.Fatalf("Failed to get %s for %s: %v", option, section, err)
	}
	return s
}

func svcOpt(cfg *conf.ConfigFile, service, option string, required bool) string {
	opt := cfgOpt(cfg, service, option)
	if opt == "" && required {
		log.Fatalf("Service %s has missing %s option", service, option)
	}
	return opt
}

var signalNames = map[syscall.Signal]string{
	syscall.SIGINT:  "SIGINT",
	syscall.SIGQUIT: "SIGQUIT",
	syscall.SIGTERM: "SIGTERM",
}

func signalName(s syscall.Signal) string {
	if name, ok := signalNames[s]; ok {
		return name
	}
	return fmt.Sprintf("SIG %d", s)
}

type Shutdowner interface {
	Shutdown()
}

func shutdownHandler(server Shutdowner) {
	sigc := make(chan os.Signal, 3)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		for s := range sigc {
			name := s.String()
			if sig, ok := s.(syscall.Signal); ok {
				name = signalName(sig)
			}
			log.Printf("Received %v, initiating shutdown...", name)
			server.Shutdown()
		}
	}()
}

var (
	restartDelay      = 2 * time.Second
	restartBackoff    = 5 * time.Second
	restartBackoffMax = 60 * time.Second
)

type Watchdog struct {
	services map[string]*Service
	shutdown chan bool
}

func NewWatchdog() *Watchdog {
	return &Watchdog{
		services: make(map[string]*Service),
		shutdown: make(chan bool),
	}
}

//关闭服务
func (w *Watchdog) Shutdown() {
	select {
	case w.shutdown <- true:
	default:
	}
}

//添加服务,如果存在
func (w *Watchdog) AddService(name, binary string) (*Service, error) {
	if _, ok := w.services[name]; ok {
		return nil, fmt.Errorf("Service %q already exists", name)
	}

	svc := newService(name, binary)
	w.services[name] = svc

	return svc, nil
}

//启动服务
func (w *Watchdog) Walk() {
	log.Printf("Seesaw watchdog starting...")

	w.mapDependencies()

	for _, svc := range w.services {
		go svc.run()
	}
	<-w.shutdown
	for _, svc := range w.services {
		go svc.stop()
	}
	for _, svc := range w.services {
		stopped := <-svc.stopped
		svc.stopped <- stopped
	}
}

//设置依赖关系
func (w *Watchdog) mapDependencies() {
	for name := range w.services {
		svc := w.services[name]
		for depName := range svc.dependencies {
			dep, ok := w.services[depName]
			if !ok {
				log.Fatalf("Failed to find dependency %q for service %q", depName, name)
			}
			svc.dependencies[depName] = dep //依赖谁,依赖启动后才会启动自身
			dep.dependents[svc.name] = svc  //谁依赖它,依赖它的服务退出后,才退出本身
		}
	}
}

//默认的优先级为0
const prioProcess = 0

//定义服务的类型.
type Service struct {
	name   string
	binary string
	path   string
	args   []string

	uid      uint32
	gid      uint32
	priority int

	dependencies map[string]*Service
	dependents   map[string]*Service

	termTimeout time.Duration

	lock    sync.Mutex
	process *os.Process

	done     chan bool
	shutdown chan bool
	started  chan bool
	stopped  chan bool

	failures uint64
	restarts uint64

	lastFailure time.Time
	lastRestart time.Time
}

//初始化一个Service.
func newService(name, binary string) *Service {
	return &Service{
		name:         name,
		binary:       binary,
		args:         make([]string, 0),
		dependencies: make(map[string]*Service),
		dependents:   make(map[string]*Service),

		done:     make(chan bool),
		shutdown: make(chan bool, 1),
		started:  make(chan bool, 1),
		stopped:  make(chan bool, 1),

		termTimeout: 5 * time.Second,
	}
}

//给这个服务添加依赖.
func (svc *Service) AddDependency(name string) {
	svc.dependencies[name] = nil
}

//为服务添加启动参数.
func (svc *Service) AddArgs(args string) {
	svc.args = strings.Fields(args)
}

//为进程设置优先级,Windows下面无效.
func (svc *Service) SetPriority(priority int) error {
	if priority < -20 || priority > 19 {
		return fmt.Errorf("Invalid priority %d - must be between -20 and 19", priority)
	}
	svc.priority = priority
	return nil
}

func (svc *Service) SetTermTimeout(tt time.Duration) {
	svc.termTimeout = tt
}

func (svc *Service) SetUser(username string) error {
	u, err := user.Lookup(username)
	if err != nil {
		return err
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return err
	}
	svc.uid = uint32(uid)
	svc.gid = uint32(gid)
	return nil
}

func (svc *Service) run() {
	//如果存在依赖,要等依赖全部启动完毕之后才会自动自身.
	for _, dep := range svc.dependencies {
		log.Printf("Service %s waiting for %s to start", svc.name, dep.name)
		select {
		case started := <-dep.started:
			dep.started <- started
		case <-svc.shutdown:
			goto done
		}
	}

	for {
		//如果启动失败,怎等待时间会延长,最大不超过restartBackoffMax时间
		//程序启动必须是阻塞的,不然会重复运行
		if svc.failures > 0 {
			delay := time.Duration(svc.failures) * restartBackoff
			if delay > restartBackoffMax {
				delay = restartBackoffMax
			}
			log.Printf("Service %s has failed %d times - delaying %s before restart",
				svc.name, svc.failures, delay)

			select {
			case <-time.After(delay):
			case <-svc.shutdown:
				goto done
			}
		}

		svc.restarts++
		svc.lastRestart = time.Now()
		svc.runOnce()

		select {
		case <-time.After(restartDelay):
		case <-svc.shutdown:
			goto done
		}
	}
done:
	svc.done <- true
}

//为服务创建日志文件
func (svc *Service) logFile() (*os.File, error) {
	logName := svc.name + ".log"

	if err := os.MkdirAll(logDir, 0666); err != nil {
		if !os.IsExist(err) {
			return nil, err
		}
	}
	f, err := os.Create(path.Join(logDir, logName))
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(f, "Log file for %s (stdout/stderr)\n", svc.name)
	fmt.Fprintf(f, "Created at: %s\n", time.Now().Format("2006/01/02 15:04:05"))
	return f, nil
}

//运行程序
func (svc *Service) runOnce() {
	args := make([]string, len(svc.args)+1)
	args[0] = svc.name
	copy(args[1:], svc.args)

	fmt.Println("Args:", args)
	null, err := os.Open(os.DevNull)
	if err != nil {
		log.Printf("Service %s - failed to open %s: %v", svc.name, os.DevNull, err)
		return
	}

	lfile, err := svc.logFile()
	if err != nil {
		log.Printf("Service %s - failed to create log file: %v", svc.name, err)
		null.Close()
		return
	}

	attr := newProc(svc, null, lfile)

	log.Printf("Starting service %s...", svc.name)
	proc, err := os.StartProcess(svc.binary, args, attr)
	if err != nil {
		log.Printf("Service %s failed to start: %v", svc.name, err)
		svc.lastFailure = time.Now()
		svc.failures++
		null.Close()
		return
	}

	null.Close()
	lfile.Close()
	svc.lock.Lock()
	svc.process = proc
	svc.lock.Unlock()

	if err := setPriority(uintptr(proc.Pid), uintptr(svc.priority)); err != 0 {
		log.Printf("Failed to set priority to %d for service %s: %v", svc.priority, svc.name, err)
	}
	select {
	case svc.started <- true:
	default:
	}

	state, err := svc.process.Wait()
	if err != nil {
		log.Printf("Service %s wait failed with %v", svc.name, err)
		svc.lastFailure = time.Now()
		svc.failures++
		return
	}
	if !state.Success() {
		log.Printf("Service %s exited with %v", svc.name, state)
		svc.lastFailure = time.Now()
		svc.failures++
		return
	}

	svc.failures = 0
	log.Printf("Service %s exited normally.", svc.name)
}

//给进程发送信号
func (svc *Service) signal(sig os.Signal) error {
	svc.lock.Lock()
	defer svc.lock.Unlock()
	if svc.process == nil {
		return nil
	}
	return svc.process.Signal(sig)
}

//停止服务
func (svc *Service) stop() {
	log.Printf("Stopping service %s...", svc.name)
	//等待依赖它的进程退出完毕之后再退出自己.
	for _, dep := range svc.dependents {
		log.Printf("Service %s waiting for %s to stop", svc.name, dep.name)
		stopped := <-dep.stopped
		dep.stopped <- stopped
	}

	svc.shutdown <- true
	//首先给进程发送退出信号,如果超时没有退出,则直接发送Kill信号.
	svc.signal(syscall.SIGTERM)
	select {
	case <-svc.done:
	case <-time.After(svc.termTimeout):
		svc.signal(syscall.SIGKILL)
		<-svc.done
	}
	log.Printf("Service %s stopped", svc.name)
	svc.stopped <- true
}

func main() {
	flag.Parse()
	if *logpath == "" {
		*logpath = os.Args[0] + ".log"
	}
	logFile, err := os.Create(*logpath)
	if err != nil {
		log.Fatalf("Create log file error:%s\n", err.Error())
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	cfg, err := conf.ReadConfigFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config file %q: %v", *configFile, err)
	}

	fido := watchdog.NewWatchdog()

	shutdownHandler(fido)
	for _, name := range cfg.GetSections() {
		if name == "default" {
			continue
		}

		binary := svcOpt(cfg, name, "binary", true)
		args := svcOpt(cfg, name, "args", false)

		svc, err := fido.AddService(name, binary)
		if err != nil {
			log.Fatalf("Failed to add service %q: %v", name, err)
		}
		svc.AddArgs(args)
		if dep := svcOpt(cfg, name, "dependency", false); dep != "" {
			svc.AddDependency(dep)
		}
		if opt := svcOpt(cfg, name, "priority", false); opt != "" {
			prio, err := strconv.Atoi(opt)
			if err != nil {
				log.Fatalf("Service %s has invalid priority %q: %v", name, opt, err)
			}
			if err := svc.SetPriority(prio); err != nil {
				log.Fatalf("Failed to set priority for service %s: %v", name, err)
			}
		}
		if opt := svcOpt(cfg, name, "term_timeout", false); opt != "" {
			tt, err := time.ParseDuration(opt)
			if err != nil {
				log.Fatalf("Service %s has invalid term_timeout %q: %v", name, opt, err)
			}
			svc.SetTermTimeout(tt)
		}

		if user := svcOpt(cfg, name, "user", false); user != "" {
			if err := svc.SetUser(user); err != nil {
				log.Fatalf("Failed to set user for service %s: %v", name, err)
			}
		}
	}

	fido.Walk()
}
