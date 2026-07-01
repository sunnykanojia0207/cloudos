package oci

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// DockerEngine implements ContainerEngine via the `docker` CLI.
//
// This adapter shells out to the Docker CLI rather than using the Docker
// SDK directly. This keeps dependencies minimal, works with any Docker
// version, and makes it easy to swap to Podman or nerdctl by implementing
// the same ContainerEngine interface.
type DockerEngine struct{}

// NewDockerEngine creates a new Docker adapter.
func NewDockerEngine() *DockerEngine {
	return &DockerEngine{}
}

func (e *DockerEngine) Name() string { return "docker" }

// Available checks if docker is installed and the daemon is responsive.
func (e *DockerEngine) Available(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker", "info", "--format", "{{.ServerVersion}}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("docker daemon not available: %w", err)
	}
	if len(output) == 0 {
		return fmt.Errorf("docker daemon returned empty version")
	}
	return nil
}

// Pull pulls an OCI image from a registry.
func (e *DockerEngine) Pull(ctx context.Context, image string) error {
	cmd := exec.CommandContext(ctx, "docker", "pull", image)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker pull %q failed: %s: %w", image, truncateOutput(string(output)), err)
	}
	return nil
}

// Run creates and starts a container. Returns the container ID.
//
// Builds a `docker run -d` command with:
//   - volume mounts (bind mount artifact dir)
//   - port mappings (host:container)
//   - environment variables
//   - working directory
//   - labels for identification
//   - name (if configured)
func (e *DockerEngine) Run(ctx context.Context, config *ContainerConfig) (string, error) {
	args := []string{"run", "-d"}

	// Add labels for identification.
	for k, v := range config.Labels {
		args = append(args, "--label", fmt.Sprintf("%s=%s", k, v))
	}

	// Auto-remove on exit.
	if config.AutoRemove {
		args = append(args, "--rm")
	}

	// Container name.
	if config.Name != "" {
		args = append(args, "--name", config.Name)
	}

	// Working directory.
	if config.WorkDir != "" {
		args = append(args, "-w", config.WorkDir)
	}

	// Network mode.
	netMode := config.NetworkMode
	if netMode == "" {
		netMode = "bridge"
	}
	args = append(args, "--network", netMode)

	// Environment variables.
	for k, v := range config.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Port mappings.
	for hostPort, containerPort := range config.Ports {
		args = append(args, "-p", fmt.Sprintf("%d:%d", hostPort, containerPort))
	}

	// Volume mounts (bind mounts).
	for hostPath, containerPath := range config.Volumes {
		hostPath = convertPath(hostPath)
		args = append(args, "-v", fmt.Sprintf("%s:%s", hostPath, containerPath))
	}

	// Image.
	args = append(args, config.Image)

	// Command and arguments.
	if config.Command != "" {
		args = append(args, config.Command)
	}
	for _, arg := range config.Args {
		args = append(args, arg)
	}

	// Execute docker run.
	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker run failed: %s: %w", truncateOutput(string(output)), err)
	}

	containerID := strings.TrimSpace(string(output))
	if containerID == "" {
		return "", fmt.Errorf("docker run returned empty container ID")
	}

	return containerID, nil
}

// Stop stops a running container gracefully.
// If timeout is nil, defaults to 10 seconds.
func (e *DockerEngine) Stop(ctx context.Context, containerID string, timeout *time.Duration) error {
	args := []string{"stop"}
	if timeout != nil {
		t := int(timeout.Seconds())
		if t > 0 {
			args = append(args, "-t", strconv.Itoa(t))
		}
	}
	args = append(args, containerID)

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If container is already stopped, that's fine.
		if strings.Contains(string(output), "already being stopped") ||
			strings.Contains(string(output), "No such container") {
			return nil
		}
		return fmt.Errorf("docker stop %q failed: %s: %w", containerID, truncateOutput(string(output)), err)
	}
	return nil
}

// Remove removes a container.
func (e *DockerEngine) Remove(ctx context.Context, containerID string, force bool) error {
	args := []string{"rm"}
	if force {
		args = append(args, "-f")
	}
	args = append(args, containerID)

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If container doesn't exist, that's fine.
		if strings.Contains(string(output), "No such container") {
			return nil
		}
		return fmt.Errorf("docker rm %q failed: %s: %w", containerID, truncateOutput(string(output)), err)
	}
	return nil
}

// Inspect returns detailed information about a container.
func (e *DockerEngine) Inspect(ctx context.Context, containerID string) (*ContainerInfo, error) {
	cmd := exec.CommandContext(ctx, "docker", "inspect", containerID)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker inspect %q failed: %w", containerID, err)
	}

	// docker inspect returns an array.
	var inspections []map[string]interface{}
	if err := json.Unmarshal(output, &inspections); err != nil {
		return nil, fmt.Errorf("parse docker inspect output: %w", err)
	}
	if len(inspections) == 0 {
		return nil, fmt.Errorf("container %q not found", containerID)
	}

	info := inspections[0]

	// Extract state.
	state, _ := info["State"].(map[string]interface{})
	containerState := ContainerUnknown
	statusStr := ""
	exitCode := 0
	var startedAt, createdAt time.Time

	if state != nil {
		if s, ok := state["Status"].(string); ok {
			statusStr = s
			switch s {
			case "created":
				containerState = ContainerCreated
			case "running":
				containerState = ContainerRunning
			case "paused":
				containerState = ContainerPaused
			case "restarting":
				containerState = ContainerRestarting
			case "exited":
				containerState = ContainerExited
			case "removing":
				containerState = ContainerRemoving
			case "dead":
				containerState = ContainerDead
			}
		}
		if c, ok := state["ExitCode"].(float64); ok {
			exitCode = int(c)
		}
		if s, ok := state["StartedAt"].(string); ok {
			startedAt, _ = time.Parse(time.RFC3339Nano, s)
		}
	}

	// Extract config.
	cfg, _ := info["Config"].(map[string]interface{})
	imageName := ""
	if cfg != nil {
		if img, ok := cfg["Image"].(string); ok {
			imageName = img
		}
	}

	// Extract name.
	name, _ := info["Name"].(string)
	name = strings.TrimPrefix(name, "/")

	// Extract ports.
	ports := make(map[int]int)
	netSettings, _ := info["NetworkSettings"].(map[string]interface{})
	if netSettings != nil {
		portMap, _ := netSettings["Ports"].(map[string]interface{})
		for containerPortStr, bindings := range portMap {
			// Format: "8080/tcp"
			portParts := strings.Split(containerPortStr, "/")
			if len(portParts) == 0 {
				continue
			}
			containerPort, err := strconv.Atoi(portParts[0])
			if err != nil {
				continue
			}
			// Extract host port from bindings.
			bindingArr, ok := bindings.([]interface{})
			if ok && len(bindingArr) > 0 {
				binding, _ := bindingArr[0].(map[string]interface{})
				if binding != nil {
					if hp, ok := binding["HostPort"].(string); ok {
						hostPort, err := strconv.Atoi(hp)
						if err == nil {
							ports[hostPort] = containerPort
						}
					}
				}
			}
		}
	}

	return &ContainerInfo{
		ID:        containerID,
		Name:      name,
		Image:     imageName,
		State:     containerState,
		Status:    statusStr,
		Ports:     ports,
		CreatedAt: createdAt,
		StartedAt: startedAt,
		ExitCode:  exitCode,
	}, nil
}

// Logs retrieves container logs.
func (e *DockerEngine) Logs(ctx context.Context, containerID string, follow bool, tail int) ([]byte, error) {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	if tail > 0 {
		args = append(args, "--tail", strconv.Itoa(tail))
	}
	args = append(args, containerID)

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker logs %q failed: %w", containerID, err)
	}
	return output, nil
}

// LogStream returns a channel of log lines for streaming.
func (e *DockerEngine) LogStream(ctx context.Context, containerID string, follow bool, tail int) (<-chan string, <-chan error, error) {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	if tail > 0 {
		args = append(args, "--tail", strconv.Itoa(tail))
	}
	args = append(args, containerID)

	cmd := exec.CommandContext(ctx, "docker", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("create stdout pipe for docker logs: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("start docker logs: %w", err)
	}

	lines := make(chan string, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(lines)
		defer close(errs)
		defer cmd.Wait()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			select {
			case lines <- scanner.Text():
			case <-ctx.Done():
				cmd.Process.Kill()
				errs <- ctx.Err()
				return
			}
		}
		if err := scanner.Err(); err != nil {
			errs <- err
		}
	}()

	return lines, errs, nil
}

// Stats returns resource usage statistics for a container.
func (e *DockerEngine) Stats(ctx context.Context, containerID string) (*ContainerStats, error) {
	// Run docker stats as a one-shot (--no-stream).
	cmd := exec.CommandContext(ctx, "docker", "stats", "--no-stream", "--format", "{{json .}}", containerID)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker stats %q failed: %w", containerID, err)
	}

	// Parse JSON output.
	var stat struct {
		CPUPerc    string `json:"CPUPerc"`
		MemUsage   string `json:"MemUsage"`
		MemPerc    string `json:"MemPerc"`
		NetIO      string `json:"NetIO"`
		BlockIO    string `json:"BlockIO"`
		PIDs       string `json:"PIDs"`
		MemLimit   string `json:"MemLimit"`
	}
	if err := json.Unmarshal(output, &stat); err != nil {
		return nil, fmt.Errorf("parse docker stats output: %w", err)
	}

	// Parse CPU percentage (e.g. "2.50%").
	cpuPerc := 0.0
	if strings.HasSuffix(stat.CPUPerc, "%") {
		cpuPerc, _ = strconv.ParseFloat(strings.TrimSuffix(stat.CPUPerc, "%"), 64)
	}

	// Parse memory usage (e.g. "15.4MiB / 7.5GiB").
	var memUsage, memLimit uint64
	if stat.MemUsage != "" {
		parts := strings.Split(stat.MemUsage, " / ")
		if len(parts) >= 1 {
			memUsage = parseBytes(parts[0])
		}
		if len(parts) >= 2 {
			memLimit = parseBytes(parts[1])
		}
	}

	// Parse network I/O (e.g. "10.5kB / 2.3MB").
	var netRx, netTx uint64
	if stat.NetIO != "" {
		parts := strings.Split(stat.NetIO, " / ")
		if len(parts) >= 1 {
			netRx = parseBytes(strings.TrimSpace(parts[0]))
		}
		if len(parts) >= 2 {
			netTx = parseBytes(strings.TrimSpace(parts[1]))
		}
	}

	// Parse block I/O (e.g. "0B / 0B").
	var blockRead, blockWrite uint64
	if stat.BlockIO != "" {
		parts := strings.Split(stat.BlockIO, " / ")
		if len(parts) >= 1 {
			blockRead = parseBytes(strings.TrimSpace(parts[0]))
		}
		if len(parts) >= 2 {
			blockWrite = parseBytes(strings.TrimSpace(parts[1]))
		}
	}

	// Parse PIDs.
	var pids uint64
	if stat.PIDs != "" && stat.PIDs != "-" {
		pids, _ = strconv.ParseUint(stat.PIDs, 10, 64)
	}

	return &ContainerStats{
		CPUPercent:  cpuPerc,
		MemoryUsage: memUsage,
		MemoryLimit: memLimit,
		NetworkRx:   netRx,
		NetworkTx:   netTx,
		BlockRead:   blockRead,
		BlockWrite:  blockWrite,
		PIDs:        pids,
		Timestamp:   time.Now(),
	}, nil
}

// List returns all containers matching the optional label filter.
func (e *DockerEngine) List(ctx context.Context, labelFilter map[string]string) ([]ContainerInfo, error) {
	args := []string{"ps", "-a", "--format", "{{json .}}"}

	// Add label filters.
	for k, v := range labelFilter {
		args = append(args, "--filter", fmt.Sprintf("label=%s=%s", k, v))
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker ps failed: %w", err)
	}

	// Parse JSON lines (docker returns one JSON object per line).
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var containers []ContainerInfo
	for _, line := range lines {
		if line == "" {
			continue
		}
		var entry struct {
			ID      string `json:"ID"`
			Names   string `json:"Names"`
			Image   string `json:"Image"`
			State   string `json:"State"`
			Status  string `json:"Status"`
			Ports   string `json:"Ports"`
			Created string `json:"CreatedAt"`
		}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		containers = append(containers, ContainerInfo{
			ID:     entry.ID,
			Name:   entry.Names,
			Image:  entry.Image,
			State:  ContainerState(entry.State),
			Status: entry.Status,
		})
	}

	return containers, nil
}

// ── Helpers ─────────────────────────────────────────────────────────────────

// convertPath converts a Windows path to a format Docker understands.
// On Windows, C:\path\to\dir becomes /c/path/to/dir for Git Bash/Mingw,
// or /host_mnt/c/path/to/dir for Docker Desktop.
// On Linux/macOS, returns the path unchanged.
func convertPath(path string) string {
	if runtime.GOOS != "windows" {
		return path
	}
	// Docker Desktop for Windows expects Unix-style paths with /host_mnt/ prefix.
	if len(path) >= 2 && path[1] == ':' {
		drive := strings.ToLower(string(path[0]))
		rest := strings.ReplaceAll(path[2:], "\\", "/")
		return fmt.Sprintf("/host_mnt/%s%s", drive, rest)
	}
	return strings.ReplaceAll(path, "\\", "/")
}

// parseBytes converts a human-readable byte string (e.g. "15.4MiB") to bytes.
func parseBytes(s string) uint64 {
	s = strings.TrimSpace(s)
	multiplier := uint64(1)

	switch {
	case strings.HasSuffix(s, "TiB"):
		multiplier = 1 << 40
		s = strings.TrimSuffix(s, "TiB")
	case strings.HasSuffix(s, "GiB"):
		multiplier = 1 << 30
		s = strings.TrimSuffix(s, "GiB")
	case strings.HasSuffix(s, "MiB"):
		multiplier = 1 << 20
		s = strings.TrimSuffix(s, "MiB")
	case strings.HasSuffix(s, "KiB"):
		multiplier = 1 << 10
		s = strings.TrimSuffix(s, "KiB")
	case strings.HasSuffix(s, "TB"):
		multiplier = 1e12
		s = strings.TrimSuffix(s, "TB")
	case strings.HasSuffix(s, "GB"):
		multiplier = 1e9
		s = strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "MB"):
		multiplier = 1e6
		s = strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "KB"):
		multiplier = 1e3
		s = strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "B"):
		s = strings.TrimSuffix(s, "B")
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return uint64(val * float64(multiplier))
}

// truncateOutput truncates command output for error messages.
func truncateOutput(output string) string {
	if len(output) > 500 {
		return output[:500] + "..."
	}
	return output
}
