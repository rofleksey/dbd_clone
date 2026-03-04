package docker

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Manager struct {
	mu            sync.Mutex
	cli           *client.Client
	gameImage     string
	portMin       int
	portMax       int
	usedPorts     map[int]bool
	hostIP        string
	masterPort    string
	containers    map[int]string // gameID -> containerID
	ports         map[int]int    // gameID -> port
	logVolumeName string         // Docker volume name for game server logs
}

func NewManager() *Manager {
	portMin, _ := strconv.Atoi(getEnvDefault("GAME_PORT_MIN", "10000"))
	portMax, _ := strconv.Atoi(getEnvDefault("GAME_PORT_MAX", "10100"))

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("WARNING: Failed to create Docker client: %v", err)
		log.Printf("Game server containers will not be available")
	}

	m := &Manager{
		cli:           cli,
		gameImage:     getEnvDefault("GAME_SERVER_IMAGE", "dbd-game-server"),
		portMin:       portMin,
		portMax:       portMax,
		usedPorts:     make(map[int]bool),
		hostIP:        getEnvDefault("HOST_IP", "host.docker.internal"),
		masterPort:    getEnvDefault("MASTER_PORT", "8080"),
		containers:    make(map[int]string),
		ports:         make(map[int]int),
		logVolumeName: getEnvDefault("GAME_LOG_DIR", "dbd-logs-data"),
	}
	if cli != nil {
		m.stopAllGameContainers(context.Background())
	}
	return m
}

func getEnvDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// stopAllGameContainers stops and removes all existing dbd-game-* containers on master server launch.
// This avoids "port already allocated" and ensures a clean state after master restarts.
func (m *Manager) stopAllGameContainers(ctx context.Context) {
	listOpts := container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", "dbd-game")),
	}
	containers, err := m.cli.ContainerList(ctx, listOpts)
	if err != nil {
		log.Printf("WARNING: could not list containers to stop: %v", err)
		return
	}
	timeout := 5
	for _, c := range containers {
		names := strings.Join(c.Names, ", ")
		if err := m.cli.ContainerStop(ctx, c.ID, container.StopOptions{Timeout: &timeout}); err != nil {
			log.Printf("WARNING: failed to stop container %s (%s): %v", c.ID[:12], names, err)
		} else {
			log.Printf("Stopped game container %s (%s)", c.ID[:12], names)
		}
		if err := m.cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true}); err != nil {
			log.Printf("WARNING: failed to remove container %s (%s): %v", c.ID[:12], names, err)
		}
	}
}

func (m *Manager) allocatePort() (int, error) {
	for p := m.portMin; p <= m.portMax; p++ {
		if !m.usedPorts[p] {
			m.usedPorts[p] = true
			return p, nil
		}
	}
	return 0, fmt.Errorf("no available ports")
}

func (m *Manager) StartGameServer(ctx context.Context, gameID int, killerID int, maxPlayers int) (string, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cli == nil {
		return "", 0, fmt.Errorf("docker client not available")
	}

	port, err := m.allocatePort()
	if err != nil {
		return "", 0, err
	}

	containerName := fmt.Sprintf("dbd-game-%d", gameID)
	masterURL := fmt.Sprintf("http://%s:%s", m.hostIP, m.masterPort)
	portStr := fmt.Sprintf("%d", port)

	containerConfig := &container.Config{
		Image: m.gameImage,
		Env: []string{
			fmt.Sprintf("GAME_ID=%d", gameID),
			fmt.Sprintf("GAME_PORT=%d", port),
			fmt.Sprintf("MASTER_URL=%s", masterURL),
			fmt.Sprintf("MAX_PLAYERS=%d", maxPlayers),
			fmt.Sprintf("KILLER_ID=%d", killerID),
			"LOG_DIR=/logs",
		},
		ExposedPorts: nat.PortSet{
			nat.Port(portStr + "/tcp"): struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(portStr + "/tcp"): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: portStr,
				},
			},
		},
		ExtraHosts: []string{"host.docker.internal:host-gateway"},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: m.logVolumeName,
				Target: "/logs",
			},
		},
	}

	resp, err := m.cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		m.usedPorts[port] = false
		return "", 0, fmt.Errorf("failed to create container: %w", err)
	}

	if err := m.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		m.usedPorts[port] = false
		m.cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		return "", 0, fmt.Errorf("failed to start container: %w", err)
	}

	m.containers[gameID] = resp.ID
	m.ports[gameID] = port

	log.Printf("Started game server container %s on port %d for game %d", resp.ID[:12], port, gameID)

	return resp.ID, port, nil
}

func (m *Manager) StopGameServer(ctx context.Context, gameID int) error {
	m.mu.Lock()
	containerID, exists := m.containers[gameID]
	port := m.ports[gameID]
	if !exists {
		m.mu.Unlock()
		return nil
	}
	delete(m.containers, gameID)
	delete(m.ports, gameID)
	if port > 0 {
		m.usedPorts[port] = false
	}
	m.mu.Unlock()

	if m.cli == nil {
		return fmt.Errorf("docker client not available")
	}

	timeout := 5
	stopOptions := container.StopOptions{Timeout: &timeout}
	if err := m.cli.ContainerStop(ctx, containerID, stopOptions); err != nil {
		log.Printf("Warning: failed to stop container %s: %v", containerID[:12], err)
	}

	if err := m.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
		log.Printf("Warning: failed to remove container %s: %v", containerID[:12], err)
	}

	log.Printf("Stopped game server container %s for game %d", containerID[:12], gameID)
	return nil
}

func (m *Manager) FreePort(port int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.usedPorts[port] = false
}

func (m *Manager) GetContainerID(gameID int) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.containers[gameID]
}
