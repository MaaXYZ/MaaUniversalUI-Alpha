package engine

import (
	"context"
	"encoding/json"
	"errors"
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

// setRunning sets the running state and emits event to frontend
func (s *service) setRunning(running bool) {
	s.mu.Lock()
	s.isRunning = running
	s.mu.Unlock()
	runtime.EventsEmit(s.ctx, "engine:running", running)
}

func (s *service) Start() {
	s.mu.RLock()
	running := s.isRunning
	s.mu.RUnlock()
	if running {
		log.Println("engine is already running")
		return
	}

	log.Println("engine starting...")

	s.setRunning(true)

	s.tasker = maa.NewTasker()

	piSrv := pi.PI()
	v2Loaded := piSrv.V2Loaded()
	piConf := piSrv.GetConfig()
	if v2Loaded == nil || v2Loaded.Interface == nil || piConf == nil {
		runtime.EventsEmit(s.ctx, "app:error", "v2 loaded or interface or config is nil")
		log.Println("v2 loaded or interface or config is nil")
		s.Stop()
		return
	}
	iface := v2Loaded.Interface

	err := s.initRes(iface, piConf)
	if err != nil {
		runtime.EventsEmit(s.ctx, "app:error", "failed to initialize resource: "+err.Error())
		log.Println("failed to initialize resource: ", err)
		s.Stop()
		return
	}

	err = s.initCtrl(iface, piConf)
	if err != nil {
		runtime.EventsEmit(s.ctx, "app:error", "failed to initialize controller: "+err.Error())
		log.Println("failed to initialize controller: ", err)
		s.Stop()
		return
	}

	if iface.Agent != nil {
		err = s.initAgent(iface)
		if err != nil {
			runtime.EventsEmit(s.ctx, "app:error", "failed to initialize agent: "+err.Error())
			log.Println("failed to initialize agent: ", err)
			s.Stop()
			return
		}
	}

	taskList := GetTaskList()

	go func() {
		for _, task := range taskList {
			s.mu.RLock()
			running := s.isRunning
			s.mu.RUnlock()
			if !running || s.tasker == nil {
				runtime.EventsEmit(s.ctx, "app:error", "engine is being stopped")
				log.Println("engine is being stopped")
				return
			}

			task.StartedAt = time.Now()
			task.Status = s.tasker.PostTask(task.Entry, task.PipelineOverride).Wait().Status()
			task.FinishedAt = time.Now()
		}

		s.Stop()
	}()
}

func (s *service) initRes(iface *pi.V2Interface, piConf *pi.InterfaceConfig) error {
	bundles := make([]string, 0)
	for _, res := range iface.Resource {
		if res.Name == piConf.Resource {
			bundles = append(bundles, res.Path...)
			break
		}
	}

	s.res = maa.NewResource()
	if s.res == nil {
		return errors.New("failed to create resource")
	}

	exePath, err := os.Executable()
	if err != nil {
		return errors.New("failed to get executable path")
	}
	exeDir := filepath.Dir(exePath)

	for _, bundle := range bundles {
		bundlePath := filepath.Join(exeDir, bundle)
		if !s.res.PostBundle(bundlePath).Wait().Success() {
			return errors.New("failed to post bundle to resource")
		}
	}

	if !s.tasker.BindResource(s.res) {
		return errors.New("failed to bind resource to tasker")
	}

	return nil
}

func (s *service) initCtrl(iface *pi.V2Interface, piConf *pi.InterfaceConfig) error {

	var err error
	switch piConf.Controller.Type {
	case "Adb":
		err = s.initAdbCtrl(piConf)
	case "Win32":
		err = s.initWin32Ctrl(iface)
	default:
		return errors.New("unsupported controller type")
	}
	if err != nil {
		return err
	}

	if !s.tasker.BindController(s.ctrl) {
		return errors.New("failed to bind controller to tasker")
	}

	return nil
}

func (s *service) initAdbCtrl(piConf *pi.InterfaceConfig) error {
	adbPath := piConf.Adb.AdbPath
	address := piConf.Adb.Address
	config := piConf.Adb.Config

	configJson, err := json.Marshal(config)
	if err != nil {
		return errors.New("failed to marshal adb config")
	}

	exePath, err := os.Executable()
	if err != nil {
		return errors.New("failed to get executable path")
	}
	exeDir := filepath.Dir(exePath)
	agentPath := filepath.Join(exeDir, "share", "MaaAgentBinary")

	s.ctrl = maa.NewAdbController(adbPath, address, adb.ScreencapDefault, adb.InputDefault, string(configJson), agentPath)
	if s.ctrl == nil {
		return errors.New("failed to create adb controller")
	}

	if !s.ctrl.PostConnect().Wait().Success() {
		return errors.New("failed to post connect to adb controller")
	}

	return nil
}

func (s *service) initWin32Ctrl(iface *pi.V2Interface) error {

	if len(iface.Controller) == 0 {
		return errors.New("pi no controller")
	}

	var win32Ctrl *pi.V2Controller
	for _, ctrl := range iface.Controller {
		if ctrl.Type == "Win32" {
			win32Ctrl = &ctrl
			break
		}
	}

	if win32Ctrl == nil {
		return errors.New("pi not win32 controller")
	}

	windows := maa.FindDesktopWindows()
	if len(windows) == 0 {
		return errors.New("failed to find any windows")
	}

	var targetWnd *maa.DesktopWindow
	classRe, err := regexp.Compile(win32Ctrl.Win32.ClassRegex)
	if err != nil {
		return err
	}

	windowRe, err := regexp.Compile(win32Ctrl.Win32.WindowRegex)
	if err != nil {
		return err
	}
	for _, window := range windows {
		match := classRe.MatchString(window.ClassName)
		if !match {
			continue
		}
		match = windowRe.MatchString(window.WindowName)
		if !match {
			continue
		}
		targetWnd = window
		break
	}

	if targetWnd == nil {
		return errors.New("failed to find target window")
	}

	screencap, err := win32.ParseScreencapMethod(win32Ctrl.Win32.Screencap)
	if err != nil {
		return errors.New("failed to parse screencap method")
	}

	mouse, err := win32.ParseInputMethod(win32Ctrl.Win32.Mouse)
	if err != nil {
		return errors.New("failed to parse mouse input method")
	}

	keyboard, err := win32.ParseInputMethod(win32Ctrl.Win32.Keyboard)
	if err != nil {
		return errors.New("failed to parse keyboard input method")
	}

	s.ctrl = maa.NewWin32Controller(targetWnd.Handle, screencap, mouse, keyboard)
	if s.ctrl == nil {
		return errors.New("failed to create win32 controller")
	}

	if !s.ctrl.PostConnect().Wait().Success() {
		return errors.New("failed to post connect to win32 controller")
	}

	return nil
}

func (s *service) initAgent(iface *pi.V2Interface) error {

	identifier := iface.Agent.Identifier

	s.agent = maa.NewAgentClient(identifier)
	if s.agent == nil {
		return errors.New("failed to create agent client")
	}

	if !s.agent.BindResource(s.res) {
		return errors.New("failed to bind resource to agent client")
	}

	exePath, err := os.Executable()
	if err != nil {
		return errors.New("failed to get executable path")
	}
	exeDir := filepath.Dir(exePath)

	var ChildExec string
	if filepath.IsAbs(iface.Agent.ChildExec) {
		ChildExec = iface.Agent.ChildExec
	} else {
		ChildExec = filepath.Join(exeDir, iface.Agent.ChildExec)
	}

	if identifier == "" {
		identifier, _ = s.agent.Identifier()
		s.agentCmd = exec.Command(ChildExec, append(iface.Agent.ChildArgs, identifier)...)
	} else {
		s.agentCmd = exec.Command(ChildExec, iface.Agent.ChildArgs...)
	}

	if err := s.agentCmd.Start(); err != nil {
		return errors.New("failed to start agent child process: " + err.Error())
	}

	if !s.agent.Connect() {
		return errors.New("failed to connect to agent server")
	}

	return nil
}

func (s *service) Stop() {
	s.mu.RLock()
	running := s.isRunning
	s.mu.RUnlock()
	if !running {
		log.Println("engine is not running")
		return
	}

	if s.tasker != nil {
		if s.tasker.Running() {
			s.tasker.PostStop().Wait()
		}

		s.tasker.Destroy()
		s.tasker = nil
	}

	if s.res != nil {
		s.res.Destroy()
		s.res = nil
	}

	if s.ctrl != nil {
		s.ctrl.Destroy()
		s.ctrl = nil
	}

	if s.agent != nil {
		s.agent.Destroy()
		s.agent = nil
	}

	if s.agentCmd != nil && s.agentCmd.Process != nil {
		if s.agentCmd.ProcessState == nil {
			if err := s.agentCmd.Process.Kill(); err != nil {
				log.Println("failed to kill agent child process", err)
			}
		}
		s.agentCmd = nil
	}

	s.setRunning(false)
	log.Println("engine stopped")
}
