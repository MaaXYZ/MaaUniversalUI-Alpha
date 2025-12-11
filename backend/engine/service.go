package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"muu-alpha/backend/pi"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/MaaXYZ/maa-framework-go/v3"
	"github.com/MaaXYZ/maa-framework-go/v3/controller/adb"
	"github.com/MaaXYZ/maa-framework-go/v3/controller/win32"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	EventEngineRunning = "engine:running"
	EventAppError      = "app:error"
)

type service struct {
	ctx       context.Context
	mu        sync.RWMutex
	isRunning bool
	tasker    *maa.Tasker
	res       *maa.Resource
	ctrl      *maa.Controller
	agent     *maa.AgentClient
	agentCmd  *exec.Cmd
}

func (s *service) GetMaaVersion() string {
	return maa.Version()
}

func (s *service) GetIsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

func (s *service) Start() {
	// Use write lock to prevent race condition (TOCTOU)
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		log.Println("engine is already running")
		return
	}
	s.isRunning = true
	s.mu.Unlock()

	log.Println("engine starting...")
	runtime.EventsEmit(s.ctx, EventEngineRunning, true)

	// helper to handle initialization errors safely
	handleInitError := func(err error, cleanupFunc func()) {
		log.Println("engine start failed:", err)
		runtime.EventsEmit(s.ctx, EventAppError, err.Error())

		if cleanupFunc != nil {
			cleanupFunc()
		}
		s.mu.Lock()
		s.isRunning = false
		s.mu.Unlock()
		runtime.EventsEmit(s.ctx, EventEngineRunning, false)
	}

	tasker := maa.NewTasker()

	var (
		res      *maa.Resource
		ctrl     *maa.Controller
		agent    *maa.AgentClient
		agentCmd *exec.Cmd
	)

	localCleanup := func() {
		if tasker != nil {
			tasker.Destroy()
		}
		if agent != nil {
			agent.Destroy()
		}
		if res != nil {
			res.Destroy()
		}
		if ctrl != nil {
			ctrl.Destroy()
		}
		if agentCmd != nil && agentCmd.Process != nil {
			_ = agentCmd.Process.Kill()
			_ = agentCmd.Wait()
		}
	}

	piSrv := pi.PI()
	v2Loaded := piSrv.V2Loaded()
	piConf := piSrv.GetConfig()
	if v2Loaded == nil || v2Loaded.Interface == nil || piConf == nil {
		handleInitError(errors.New("v2 loaded or interface or config is nil"), localCleanup)
		return
	}
	iface := v2Loaded.Interface

	var err error

	// init res
	res, err = s.createRes(iface, piConf)
	if err != nil {
		handleInitError(fmt.Errorf("failed to create resource: %w", err), localCleanup)
		return
	}

	// init ctrl
	ctrl, err = s.createCtrl(iface, piConf)
	if err != nil {
		handleInitError(fmt.Errorf("failed to create controller: %w", err), localCleanup)
		return
	}

	// init agent
	if iface.Agent != nil {
		agent, agentCmd, err = s.createAgent(iface, res)
		if err != nil {
			handleInitError(fmt.Errorf("failed to create agent: %w", err), localCleanup)
			return
		}
	}

	if !tasker.BindResource(res) {
		handleInitError(errors.New("failed to bind resource to tasker"), localCleanup)
		return
	}
	if !tasker.BindController(ctrl) {
		handleInitError(errors.New("failed to bind controller to tasker"), localCleanup)
		return
	}

	s.mu.Lock()
	// Double-check if Stop() was called during initialization
	if !s.isRunning {
		s.mu.Unlock()
		log.Println("engine start aborted (stopped during init)")
		localCleanup() // Clean up the local objects we just created
		return
	}
	s.tasker = tasker
	s.res = res
	s.ctrl = ctrl
	s.agent = agent
	s.agentCmd = agentCmd
	s.mu.Unlock()

	taskList := GetTaskList()

	go func() {
		defer s.Stop()
		for _, task := range taskList {
			// Hold read lock while checking and getting tasker reference
			s.mu.RLock()
			running := s.isRunning
			tasker := s.tasker
			s.mu.RUnlock()

			if !running || tasker == nil {
				return
			}

			task.StartedAt = time.Now()
			job := tasker.PostTask(task.Entry, task.PipelineOverride).Wait()
			task.Status = job.Status()
			task.FinishedAt = time.Now()
		}
	}()
}

func (s *service) createRes(iface *pi.V2Interface, piConf *pi.InterfaceConfig) (*maa.Resource, error) {
	bundles := make([]string, 0)
	for _, res := range iface.Resource {
		if res.Name == piConf.Resource {
			bundles = append(bundles, res.Path...)
			break
		}
	}

	if len(bundles) == 0 {
		log.Printf("warning: no resource bundles found for resource name: %s", piConf.Resource)
	}

	res := maa.NewResource()
	if res == nil {
		return nil, errors.New("failed to create resource instance")
	}

	exeDir, err := s.getExecutableDir()
	if err != nil {
		res.Destroy()
		return nil, err
	}

	for _, bundle := range bundles {
		bundlePath := filepath.Join(exeDir, bundle)
		if !res.PostBundle(bundlePath).Wait().Success() {
			res.Destroy()
			return nil, fmt.Errorf("failed to post bundle: %s", bundlePath)
		}
	}

	return res, nil
}

func (s *service) createCtrl(iface *pi.V2Interface, piConf *pi.InterfaceConfig) (*maa.Controller, error) {
	switch piConf.Controller.Type {
	case "Adb":
		return s.createAdbCtrl(piConf)
	case "Win32":
		return s.createWin32Ctrl(iface)
	default:
		return nil, errors.New("unsupported controller type: " + piConf.Controller.Type)
	}
}

func (s *service) createAdbCtrl(piConf *pi.InterfaceConfig) (*maa.Controller, error) {
	adbPath := piConf.Adb.AdbPath
	address := piConf.Adb.Address
	config := piConf.Adb.Config

	configJson, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal adb config: %w", err)
	}

	exeDir, err := s.getExecutableDir()
	if err != nil {
		return nil, err
	}
	agentPath := filepath.Join(exeDir, "share", "MaaAgentBinary")

	ctrl := maa.NewAdbController(adbPath, address, adb.ScreencapDefault, adb.InputDefault, string(configJson), agentPath)
	if ctrl == nil {
		return nil, errors.New("failed to create adb controller instance")
	}

	if !ctrl.PostConnect().Wait().Success() {
		ctrl.Destroy()
		return nil, errors.New("failed to connect to adb")
	}

	return ctrl, nil
}

func (s *service) createWin32Ctrl(iface *pi.V2Interface) (*maa.Controller, error) {
	if len(iface.Controller) == 0 {
		return nil, errors.New("pi config has no controller definitions")
	}

	var win32Ctrl *pi.V2Controller
	for _, c := range iface.Controller {
		if c.Type == "Win32" {
			win32Ctrl = &c
			break
		}
	}

	if win32Ctrl == nil {
		return nil, errors.New("no Win32 controller configuration found in project interface")
	}

	windows := maa.FindDesktopWindows()
	if len(windows) == 0 {
		return nil, errors.New("failed to find any desktop windows")
	}

	var targetWnd *maa.DesktopWindow
	classRe, err := regexp.Compile(win32Ctrl.Win32.ClassRegex)
	if err != nil {
		return nil, fmt.Errorf("invalid class regex: %w", err)
	}
	windowRe, err := regexp.Compile(win32Ctrl.Win32.WindowRegex)
	if err != nil {
		return nil, fmt.Errorf("invalid window regex: %w", err)
	}
	for _, window := range windows {
		if classRe.MatchString(window.ClassName) && windowRe.MatchString(window.WindowName) {
			targetWnd = window
			break
		}
	}

	if targetWnd == nil {
		return nil, errors.New("failed to find target window matching regex")
	}

	screencap, err := win32.ParseScreencapMethod(win32Ctrl.Win32.Screencap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse screencap method: %w", err)
	}

	mouse, err := win32.ParseInputMethod(win32Ctrl.Win32.Mouse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mouse input method: %w", err)
	}

	keyboard, err := win32.ParseInputMethod(win32Ctrl.Win32.Keyboard)
	if err != nil {
		return nil, fmt.Errorf("failed to parse keyboard input method: %w", err)
	}

	ctrl := maa.NewWin32Controller(targetWnd.Handle, screencap, mouse, keyboard)
	if ctrl == nil {
		return nil, errors.New("failed to create win32 controller instance")
	}

	if !ctrl.PostConnect().Wait().Success() {
		ctrl.Destroy()
		return nil, errors.New("failed to connect to win32 window")
	}

	return ctrl, nil
}

func (s *service) createAgent(iface *pi.V2Interface, res *maa.Resource) (*maa.AgentClient, *exec.Cmd, error) {
	identifier := iface.Agent.Identifier

	agent := maa.NewAgentClient(identifier)
	if agent == nil {
		return nil, nil, errors.New("failed to create agent client instance")
	}

	// Clean up agent if subsequent steps fail
	cleanup := func() {
		agent.Destroy()
	}

	if !agent.BindResource(res) {
		cleanup()
		return nil, nil, errors.New("failed to bind resource to agent client")
	}

	id, _ := agent.Identifier()
	cmd := exec.Command(iface.Agent.ChildExec, append(iface.Agent.ChildArgs, id)...)

	exeDir, err := s.getExecutableDir()
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	cmd.Dir = exeDir

	if err := cmd.Start(); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to start agent child process: %w", err)
	}

	if !agent.Connect() {
		cleanup()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
		return nil, nil, errors.New("failed to connect to agent server")
	}

	return agent, cmd, nil
}

func (s *service) Stop() {
	// Use write lock to prevent race condition and safely swap resources
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = false

	// Capture references under lock, then clear fields
	tasker := s.tasker
	res := s.res
	ctrl := s.ctrl
	agent := s.agent
	agentCmd := s.agentCmd

	s.tasker = nil
	s.res = nil
	s.ctrl = nil
	s.agent = nil
	s.agentCmd = nil
	s.mu.Unlock()

	// Perform cleanup outside the lock to avoid blocking other operations
	if tasker != nil {
		if tasker.Running() {
			tasker.PostStop().Wait()
		}
		tasker.Destroy()
	}

	if agent != nil {
		agent.Destroy()
	}

	if res != nil {
		res.Destroy()
	}

	if ctrl != nil {
		ctrl.Destroy()
	}

	if agentCmd != nil && agentCmd.Process != nil {
		// Kill the process directly without checking ProcessState,
		// since ProcessState is only set after Wait() is called
		if err := agentCmd.Process.Kill(); err != nil {
			log.Println("failed to kill agent child process", err)
		}
		// Wait for the process to release resources
		_ = agentCmd.Wait()
	}

	runtime.EventsEmit(s.ctx, EventEngineRunning, false)
	log.Println("engine stopped")
}

func (s *service) getExecutableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", errors.New("failed to get executable path")
	}
	return filepath.Dir(exePath), nil
}
