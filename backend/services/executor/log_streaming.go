package executor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
	"xsha-backend/database"
	"xsha-backend/repository"
	"xsha-backend/utils"
)

// LogStreamingService provides real-time log streaming capabilities
type LogStreamingService interface {
	// StreamConversationLogs streams logs for a conversation
	// Returns a channel that emits log lines and an error channel
	StreamConversationLogs(ctx context.Context, conversationID uint) (<-chan string, <-chan error, error)

	// GetHistoricalLogs gets historical logs for a completed conversation
	GetHistoricalLogs(conversationID uint) (string, error)

	// IsConversationRunning checks if a conversation is currently running
	IsConversationRunning(conversationID uint) (bool, error)
}

type logStreamingService struct {
	conversationRepo repository.TaskConversationRepository
	execLogRepo      repository.TaskExecutionLogRepository
	execManager      *ExecutionManager
}

func NewLogStreamingService(
	conversationRepo repository.TaskConversationRepository,
	execLogRepo repository.TaskExecutionLogRepository,
	execManager *ExecutionManager,
) LogStreamingService {
	return &logStreamingService{
		conversationRepo: conversationRepo,
		execLogRepo:      execLogRepo,
		execManager:      execManager,
	}
}

func (s *logStreamingService) StreamConversationLogs(ctx context.Context, conversationID uint) (<-chan string, <-chan error, error) {
	// First check if conversation exists
	conv, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return nil, nil, fmt.Errorf("conversation not found: %v", err)
	}

	logChan := make(chan string, 100) // Buffer to prevent blocking
	errChan := make(chan error, 1)

	go func() {
		defer close(logChan)
		defer close(errChan)

		// Check conversation status
		isRunning := s.execManager.IsRunning(conversationID)

		if isRunning {
			// Get container ID for real-time logs
			containerID := s.execManager.GetContainerID(conversationID)
			if containerID != "" {
				// Send existing logs first
				if existingLogs, err := s.getExistingLogs(conversationID); err == nil && existingLogs != "" {
					for _, line := range strings.Split(existingLogs, "\n") {
						if line != "" {
							select {
							case logChan <- line:
							case <-ctx.Done():
								return
							}
						}
					}
				}

				// Stream real-time logs from container
				if err := s.streamContainerLogs(ctx, containerID, logChan); err != nil {
					errChan <- err
				}
			} else {
				// Fallback to polling database logs
				utils.Warn("No container ID found for running conversation, using database polling", "conversationID", conversationID)
				s.pollDatabaseLogs(ctx, conversationID, logChan, errChan)
			}
		} else {
			historicalLogs, err := s.GetHistoricalLogs(conversationID)
			if err != nil {
				errChan <- err
				return
			}

			// Send historical logs line by line
			if historicalLogs != "" {
				for _, line := range strings.Split(historicalLogs, "\n") {
					if line != "" {
						select {
						case logChan <- line:
						case <-ctx.Done():
							return
						}
					}
				}
			}

			// Send completion message
			statusMsg := fmt.Sprintf("=== Conversation completed with status: %s ===", conv.Status)
			select {
			case logChan <- statusMsg:
			case <-ctx.Done():
				return
			}
		}
	}()

	return logChan, errChan, nil
}

func (s *logStreamingService) streamContainerLogs(ctx context.Context, containerID string, logChan chan<- string) error {
	// Use docker logs with follow flag to get real-time logs
	cmd := exec.CommandContext(ctx, "docker", "logs", "-f", "--timestamps", containerID)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start docker logs command: %v", err)
	}

	// Read both stdout and stderr
	go s.readLogStream(ctx, stdout, "STDOUT", logChan)
	go s.readLogStream(ctx, stderr, "STDERR", logChan)

	// Wait for command to finish or context cancellation
	go func() {
		cmd.Wait()
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Kill the process if it's still running
	if cmd.Process != nil {
		cmd.Process.Kill()
	}

	return nil
}

func (s *logStreamingService) readLogStream(ctx context.Context, reader io.Reader, prefix string, logChan chan<- string) {
	scanner := bufio.NewScanner(reader)
	// Set larger buffer to handle large log outputs (1MB buffer)
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, 0, 64*1024) // Start with 64KB
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()

		// Protect against extremely large log lines
		if len(line) > maxCapacity {
			line = line[:maxCapacity-3] + "..."
			utils.Warn("Truncated extremely large log line in streaming", "prefix", prefix, "original_length", len(scanner.Text()))
		}

		select {
		case <-ctx.Done():
			return
		case logChan <- line:
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		utils.Error("Log streaming scanner failed", "prefix", prefix, "error", err)
		select {
		case <-ctx.Done():
			return
		case logChan <- fmt.Sprintf("ERROR - Scanner failed: %v", err):
		}
	}
}

func (s *logStreamingService) pollDatabaseLogs(ctx context.Context, conversationID uint, logChan chan<- string, errChan chan<- error) {
	var lastLogLength int
	ticker := time.NewTicker(1 * time.Second) // Poll every second
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check if conversation is still running
			if !s.execManager.IsRunning(conversationID) {
				// Conversation finished, send final logs and exit
				if finalLogs, err := s.getExistingLogs(conversationID); err == nil {
					newContent := ""
					if len(finalLogs) > lastLogLength {
						newContent = finalLogs[lastLogLength:]
					}

					if newContent != "" {
						for _, line := range strings.Split(newContent, "\n") {
							if line != "" {
								select {
								case logChan <- line:
								case <-ctx.Done():
									return
								}
							}
						}
					}
				}

				// Send completion message
				conv, _ := s.conversationRepo.GetByID(conversationID)
				statusMsg := fmt.Sprintf("=== Conversation completed with status: %s ===", conv.Status)
				select {
				case logChan <- statusMsg:
				case <-ctx.Done():
					return
				}
				return
			}

			// Get current logs
			currentLogs, err := s.getExistingLogs(conversationID)
			if err != nil {
				select {
				case errChan <- err:
				case <-ctx.Done():
				}
				return
			}

			// Check for new log content
			if len(currentLogs) > lastLogLength {
				newContent := currentLogs[lastLogLength:]
				lastLogLength = len(currentLogs)

				// Send new log lines
				for _, line := range strings.Split(newContent, "\n") {
					if line != "" {
						select {
						case logChan <- line:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}
	}
}

func (s *logStreamingService) getExistingLogs(conversationID uint) (string, error) {
	execLog, err := s.execLogRepo.GetByConversationID(conversationID)
	if err != nil {
		return "", err
	}
	return execLog.ExecutionLogs, nil
}

func (s *logStreamingService) GetHistoricalLogs(conversationID uint) (string, error) {
	execLog, err := s.execLogRepo.GetByConversationID(conversationID)
	if err != nil {
		return "", fmt.Errorf("failed to get execution log: %v", err)
	}
	return execLog.ExecutionLogs, nil
}

func (s *logStreamingService) IsConversationRunning(conversationID uint) (bool, error) {
	conv, err := s.conversationRepo.GetByID(conversationID)
	if err != nil {
		return false, fmt.Errorf("conversation not found: %v", err)
	}

	return conv.Status == database.ConversationStatusRunning || s.execManager.IsRunning(conversationID), nil
}
