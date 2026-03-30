package instance

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// Instance 表示一个 PicoClaw 实例
type Instance struct {
	ID            string `yaml:"id"`
	Name          string `yaml:"name"`
	Workspace     string `yaml:"workspace"`
	Port          int    `yaml:"port"`
	Model         string `yaml:"model"`
	APIKeyEnv     string `yaml:"api_key_env"`
	MemoryLimitMB int    `yaml:"memory_limit_mb"`
	AutoStart     bool   `yaml:"auto_start"`

	// 运行时状态
	PID        int       `json:"pid"`
	Status     string    `json:"status"` // running, stopped, crashed
	StartTime  time.Time `json:"start_time"`
	RestartCount int     `json:"restart_count"`

	// 内部字段
	cmd    *exec.Cmd
	mutex  sync.RWMutex
}

// Manager 管理所有实例
type Manager struct {
	Instances    map[string]*Instance
	ConfigPath   string
	mutex        sync.RWMutex
}

// NewManager 创建实例管理器
func NewManager(configPath string) (*Manager, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	instances := make(map[string]*Instance)
	for _, cfg := range config.Instances {
		instances[cfg.ID] = &Instance{
			ID:            cfg.ID,
			Name:          cfg.Name,
			Workspace:     cfg.Workspace,
			Port:          cfg.Port,
			Model:         cfg.Model,
			APIKeyEnv:     cfg.APIKeyEnv,
			MemoryLimitMB: cfg.MemoryLimitMB,
			AutoStart:     cfg.AutoStart,
			Status:        "stopped",
		}
	}

	return &Manager{
		Instances:  instances,
		ConfigPath: configPath,
	}, nil
}

// Start 启动单个实例
func (m *Manager) Start(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	inst, exists := m.Instances[id]
	if !exists {
		return fmt.Errorf("实例 %s 不存在", id)
	}

	if inst.Status == "running" {
		return fmt.Errorf("实例 %s 已在运行", id)
	}

	// 确保 workspace 目录存在
	if err := os.MkdirAll(inst.Workspace, 0755); err != nil {
		return fmt.Errorf("创建工作目录失败：%v", err)
	}

	// 构建 PicoClaw 启动命令
	// 假设 PicoClaw 二进制在 /usr/local/bin/picoclaw
	cmd := exec.Command("/usr/local/bin/picoclaw", "serve", "--port", fmt.Sprintf("%d", inst.Port))
	cmd.Dir = inst.Workspace
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PORT=%d", inst.Port),
		fmt.Sprintf("WORKSPACE=%s", inst.Workspace),
	)

	// 设置进程组 (便于后续杀死整个进程树)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// 启动进程
	if err := cmd.Start(); err != nil {
		inst.Status = "crashed"
		return fmt.Errorf("启动失败：%v", err)
	}

	inst.cmd = cmd
	inst.PID = cmd.Process.Pid
	inst.Status = "running"
	inst.StartTime = time.Now()

	fmt.Printf("✅ 实例 %s 已启动 (PID: %d, Port: %d)\n", inst.Name, inst.PID, inst.Port)
	return nil
}

// Stop 停止单个实例
func (m *Manager) Stop(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	inst, exists := m.Instances[id]
	if !exists {
		return fmt.Errorf("实例 %s 不存在", id)
	}

	if inst.Status != "running" {
		return nil
	}

	// 杀死进程组
	if inst.cmd != nil && inst.cmd.Process != nil {
		syscall.Kill(-inst.cmd.Process.Pid, syscall.SIGTERM)
		time.Sleep(2 * time.Second)
		// 如果还没死，强制杀死
		if inst.cmd.ProcessState == nil {
			syscall.Kill(-inst.cmd.Process.Pid, syscall.SIGKILL)
		}
	}

	inst.Status = "stopped"
	inst.PID = 0
	inst.cmd = nil

	fmt.Printf("⏹️  实例 %s 已停止\n", inst.Name)
	return nil
}

// Restart 重启实例
func (m *Manager) Restart(id string) error {
	if err := m.Stop(id); err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	return m.Start(id)
}

// AutoStart 自动启动所有标记为 auto_start 的实例
func (m *Manager) AutoStart() {
	for id := range m.Instances {
		inst := m.Instances[id]
		if inst.AutoStart {
			go func(id string) {
				if err := m.Start(id); err != nil {
					fmt.Printf("❌ 自动启动 %s 失败：%v\n", id, err)
				}
			}(id)
			time.Sleep(100 * time.Millisecond) // 避免同时启动冲击
		}
	}
}

// StopAll 停止所有实例
func (m *Manager) StopAll() {
	for id := range m.Instances {
		m.Stop(id)
	}
}

// GetStatus 获取实例状态
func (m *Manager) GetStatus(id string) (map[string]interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	inst, exists := m.Instances[id]
	if !exists {
		return nil, fmt.Errorf("实例 %s 不存在", id)
	}

	inst.mutex.RLock()
	defer inst.mutex.RUnlock()

	return map[string]interface{}{
		"id":             inst.ID,
		"name":           inst.Name,
		"status":         inst.Status,
		"pid":            inst.PID,
		"port":           inst.Port,
		"workspace":      inst.Workspace,
		"start_time":     inst.StartTime,
		"restart_count":  inst.RestartCount,
		"uptime_seconds": time.Since(inst.StartTime).Seconds(),
	}, nil
}

// GetAllStatus 获取所有实例状态
func (m *Manager) GetAllStatus() []map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	statuses := make([]map[string]interface{}, 0, len(m.Instances))
	for id := range m.Instances {
		if status, err := m.GetStatus(id); err == nil {
			statuses = append(statuses, status)
		}
	}
	return statuses
}

// CheckHealth 检查实例健康状态 (被 monitor 调用)
func (m *Manager) CheckHealth(id string) bool {
	m.mutex.RLock()
	inst, exists := m.Instances[id]
	m.mutex.RUnlock()

	if !exists || inst.Status != "running" {
		return false
	}

	// 检查进程是否还活着
	if inst.cmd != nil && inst.cmd.ProcessState != nil {
		// 进程已退出
		inst.Status = "crashed"
		inst.RestartCount++
		return false
	}

	return true
}
