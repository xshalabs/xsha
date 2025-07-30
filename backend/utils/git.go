package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type GitProtocolType string

const (
	GitProtocolHTTPS GitProtocolType = "https"
	GitProtocolSSH   GitProtocolType = "ssh"
)

type GitCredentialType string

const (
	GitCredentialTypePassword GitCredentialType = "password"
	GitCredentialTypeToken    GitCredentialType = "token"
	GitCredentialTypeSSHKey   GitCredentialType = "ssh_key"
)

type GitURLInfo struct {
	Protocol GitProtocolType `json:"protocol"`
	Host     string          `json:"host"`
	Owner    string          `json:"owner"`
	Repo     string          `json:"repo"`
	IsValid  bool            `json:"is_valid"`
}

type GitBranch struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	CommitID  string `json:"commit_id"`
}

type GitAccessResult struct {
	CanAccess    bool     `json:"can_access"`
	ErrorMessage string   `json:"error_message"`
	Branches     []string `json:"branches"`
}

type GitCredentialInfo struct {
	Type       GitCredentialType `json:"type"`
	Username   string            `json:"username"`
	Password   string            `json:"password"`
	PrivateKey string            `json:"private_key"`
	PublicKey  string            `json:"public_key"`
}

func DetectGitProtocol(repoURL string) GitProtocolType {
	repoURL = strings.TrimSpace(repoURL)

	if repoURL == "" {
		return GitProtocolHTTPS
	}

	if strings.HasPrefix(repoURL, "https://") {
		return GitProtocolHTTPS
	}

	if strings.HasPrefix(repoURL, "http://") {
		return GitProtocolHTTPS
	}

	if strings.HasPrefix(repoURL, "ssh://") {
		return GitProtocolSSH
	}

	sshPattern := regexp.MustCompile(`^[a-zA-Z0-9_.-]+@[a-zA-Z0-9.-]+:[a-zA-Z0-9/._-]+`)
	if sshPattern.MatchString(repoURL) {
		return GitProtocolSSH
	}

	return GitProtocolHTTPS
}

func ParseGitURL(repoURL string) *GitURLInfo {
	info := &GitURLInfo{
		IsValid: false,
	}

	repoURL = strings.TrimSpace(repoURL)

	info.Protocol = DetectGitProtocol(repoURL)

	if repoURL == "" {
		return info
	}

	switch info.Protocol {
	case GitProtocolHTTPS:
		return parseHTTPSURL(repoURL, info)
	case GitProtocolSSH:
		return parseSSHURL(repoURL, info)
	default:
		return info
	}
}

func parseHTTPSURL(repoURL string, info *GitURLInfo) *GitURLInfo {
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return info
	}

	info.Host = parsedURL.Host

	path := strings.Trim(parsedURL.Path, "/")

	if strings.HasSuffix(path, ".git") {
		path = strings.TrimSuffix(path, ".git")
	}

	parts := strings.Split(path, "/")
	if len(parts) >= 2 {
		info.Owner = parts[0]
		info.Repo = parts[1]
		info.IsValid = true
	}

	return info
}

func parseSSHURL(repoURL string, info *GitURLInfo) *GitURLInfo {
	if strings.HasPrefix(repoURL, "ssh://") {
		parsedURL, err := url.Parse(repoURL)
		if err != nil {
			return info
		}

		info.Host = parsedURL.Host

		path := strings.Trim(parsedURL.Path, "/")

		if strings.HasSuffix(path, ".git") {
			path = strings.TrimSuffix(path, ".git")
		}

		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			info.Owner = parts[0]
			info.Repo = parts[1]
			info.IsValid = true
		}

		return info
	}

	sshPattern := regexp.MustCompile(`^([a-zA-Z0-9_.-]+)@([a-zA-Z0-9.-]+):(.+)$`)
	matches := sshPattern.FindStringSubmatch(repoURL)
	if len(matches) == 4 {
		info.Host = matches[2]
		path := matches[3]

		if strings.HasSuffix(path, ".git") {
			path = strings.TrimSuffix(path, ".git")
		}

		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			info.Owner = parts[0]
			info.Repo = parts[1]
			info.IsValid = true
		}
	}

	return info
}

func ValidateGitURL(repoURL string) error {
	info := ParseGitURL(repoURL)
	if !info.IsValid {
		return fmt.Errorf("invalid Git URL format")
	}
	return nil
}

func IsGitURL(str string) bool {
	str = strings.TrimSpace(str)

	if strings.HasPrefix(str, "https://") || strings.HasPrefix(str, "http://") {
		return true
	}

	if strings.HasPrefix(str, "ssh://") {
		return true
	}

	sshPattern := regexp.MustCompile(`^[a-zA-Z0-9_.-]+@[a-zA-Z0-9.-]+:[a-zA-Z0-9/._-]+`)
	return sshPattern.MatchString(str)
}

func FetchRepositoryBranchesWithConfig(repoURL string, credential *GitCredentialInfo, sslVerify bool) (*GitAccessResult, error) {
	tempDir, err := ioutil.TempDir("", "git-repo-*")
	if err != nil {
		return &GitAccessResult{
			CanAccess:    false,
			ErrorMessage: fmt.Sprintf("创建临时目录失败: %v", err),
		}, nil
	}
	defer os.RemoveAll(tempDir)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	var envVars []string

	if credential != nil {
		switch credential.Type {
		case GitCredentialTypePassword:
			if credential.Password == "" {
				return &GitAccessResult{
					CanAccess:    false,
					ErrorMessage: "password cannot be empty",
				}, nil
			}

			parsedURL, err := url.Parse(repoURL)
			if err != nil {
				return &GitAccessResult{
					CanAccess:    false,
					ErrorMessage: fmt.Sprintf("failed to parse URL: %v", err),
				}, nil
			}

			parsedURL.User = url.UserPassword(credential.Username, credential.Password)
			authenticatedURL := parsedURL.String()

			cmd = exec.CommandContext(ctx, "git", "ls-remote", "--heads", authenticatedURL)

		case GitCredentialTypeToken:
			if credential.Password == "" {
				return &GitAccessResult{
					CanAccess:    false,
					ErrorMessage: "token cannot be empty",
				}, nil
			}

			parsedURL, err := url.Parse(repoURL)
			if err != nil {
				return &GitAccessResult{
					CanAccess:    false,
					ErrorMessage: fmt.Sprintf("failed to parse URL: %v", err),
				}, nil
			}

			parsedURL.User = url.UserPassword(credential.Password, "x-oauth-basic")
			authenticatedURL := parsedURL.String()

			cmd = exec.CommandContext(ctx, "git", "ls-remote", "--heads", authenticatedURL)

		case GitCredentialTypeSSHKey:
			if credential.PrivateKey == "" {
				return &GitAccessResult{
					CanAccess:    false,
					ErrorMessage: "SSH private key cannot be empty",
				}, nil
			}

			keyFile := filepath.Join(tempDir, "ssh_key")
			if err := ioutil.WriteFile(keyFile, []byte(credential.PrivateKey), 0600); err != nil {
				return &GitAccessResult{
					CanAccess:    false,
					ErrorMessage: fmt.Sprintf("failed to create SSH key file: %v", err),
				}, nil
			}

			envVars = append(os.Environ(),
				fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no", keyFile),
			)

			cmd = exec.CommandContext(ctx, "git", "ls-remote", "--heads", repoURL)
			cmd.Env = envVars

		default:
			return &GitAccessResult{
				CanAccess:    false,
				ErrorMessage: "unsupported credential type",
			}, nil
		}
	} else {
		cmd = exec.CommandContext(ctx, "git", "ls-remote", "--heads", repoURL)
	}

	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}

	if !sslVerify {
		cmd.Env = append(cmd.Env, "GIT_SSL_NO_VERIFY=true")
	}

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			errorMessage := string(exitError.Stderr)
			if sslVerify && (strings.Contains(errorMessage, "SSL") || strings.Contains(errorMessage, "TLS") || strings.Contains(errorMessage, "certificate")) {
				return &GitAccessResult{
					CanAccess:    false,
					ErrorMessage: fmt.Sprintf("repository access validation failed: %s\n建议: try to set environment variable XSHA_GIT_SSL_VERIFY=false to disable SSL verification", errorMessage),
				}, nil
			}
			return &GitAccessResult{
				CanAccess:    false,
				ErrorMessage: fmt.Sprintf("repository access validation failed: %s", errorMessage),
			}, nil
		}
		return &GitAccessResult{
			CanAccess:    false,
			ErrorMessage: fmt.Sprintf("failed to execute git command: %v", err),
		}, nil
	}

	branches := parseBranchesFromLsRemote(string(output))

	return &GitAccessResult{
		CanAccess: true,
		Branches:  branches,
	}, nil
}

func FetchRepositoryBranches(repoURL string, credential *GitCredentialInfo) (*GitAccessResult, error) {
	return FetchRepositoryBranchesWithConfig(repoURL, credential, false)
}

func parseBranchesFromLsRemote(output string) []string {
	var branches []string
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) == 2 && strings.HasPrefix(parts[1], "refs/heads/") {
			branchName := strings.TrimPrefix(parts[1], "refs/heads/")
			if branchName != "" {
				branches = append(branches, branchName)
			}
		}
	}

	return branches
}

func ValidateRepositoryAccess(repoURL string, credential *GitCredentialInfo) error {
	result, err := FetchRepositoryBranches(repoURL, credential)
	if err != nil {
		return err
	}

	if !result.CanAccess {
		return fmt.Errorf(result.ErrorMessage)
	}

	return nil
}

func GitResetToPreviousCommit(workspacePath, commitHash string) error {
	if workspacePath == "" {
		return fmt.Errorf("workspace path cannot be empty")
	}

	if commitHash == "" {
		return fmt.Errorf("commit hash cannot be empty")
	}

	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return fmt.Errorf("workspace directory does not exist: %s", workspacePath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	getPrevCmd := exec.CommandContext(ctx, "git", "rev-parse", commitHash+"^")
	getPrevCmd.Dir = workspacePath

	prevCommitOutput, err := getPrevCmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			errorMessage := string(exitError.Stderr)
			return fmt.Errorf("failed to get previous commit: %s", errorMessage)
		}
		return fmt.Errorf("failed to execute git rev-parse command: %v", err)
	}

	prevCommitHash := strings.TrimSpace(string(prevCommitOutput))
	if prevCommitHash == "" {
		return fmt.Errorf("failed to get previous commit hash")
	}

	resetCmd := exec.CommandContext(ctx, "git", "reset", "--hard", prevCommitHash)
	resetCmd.Dir = workspacePath

	_, err = resetCmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			errorMessage := string(exitError.Stderr)
			return fmt.Errorf("failed to reset to previous commit: %s", errorMessage)
		}
		return fmt.Errorf("failed to execute git reset command: %v", err)
	}

	return nil
}

type GitDiffFile struct {
	Path        string `json:"path"`
	Status      string `json:"status"`       // added, modified, deleted, renamed
	Additions   int    `json:"additions"`    // added lines
	Deletions   int    `json:"deletions"`    // deleted lines
	IsBinary    bool   `json:"is_binary"`    // is binary file
	OldPath     string `json:"old_path"`     // old path
	DiffContent string `json:"diff_content"` // diff content
}

type GitDiffSummary struct {
	TotalFiles     int           `json:"total_files"`
	TotalAdditions int           `json:"total_additions"`
	TotalDeletions int           `json:"total_deletions"`
	Files          []GitDiffFile `json:"files"`
	CommitsBehind  int           `json:"commits_behind"` // commits behind
	CommitsAhead   int           `json:"commits_ahead"`  // commits ahead
}

func GetBranchDiff(workspacePath, baseBranch, compareBranch string, includeContent bool) (*GitDiffSummary, error) {
	if workspacePath == "" {
		return nil, fmt.Errorf("workspace path cannot be empty")
	}

	if baseBranch == "" {
		return nil, fmt.Errorf("base branch cannot be empty")
	}

	if compareBranch == "" {
		return nil, fmt.Errorf("compare branch cannot be empty")
	}

	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace directory does not exist: %s", workspacePath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	summary := &GitDiffSummary{
		Files: []GitDiffFile{},
	}

	if err := getBranchCommitDiff(ctx, workspacePath, baseBranch, compareBranch, summary); err != nil {
		return nil, fmt.Errorf("failed to get commit diff: %v", err)
	}

	if err := getBranchFileDiff(ctx, workspacePath, baseBranch, compareBranch, summary, includeContent); err != nil {
		return nil, fmt.Errorf("failed to get file diff: %v", err)
	}

	return summary, nil
}

func getBranchCommitDiff(ctx context.Context, workspacePath, baseBranch, compareBranch string, summary *GitDiffSummary) error {
	aheadCmd := exec.CommandContext(ctx, "git", "rev-list", "--count", fmt.Sprintf("%s..%s", baseBranch, compareBranch))
	aheadCmd.Dir = workspacePath

	aheadOutput, err := aheadCmd.Output()
	if err != nil {
		summary.CommitsAhead = 0
	} else {
		if count, parseErr := fmt.Sscanf(strings.TrimSpace(string(aheadOutput)), "%d", &summary.CommitsAhead); parseErr != nil || count != 1 {
			summary.CommitsAhead = 0
		}
	}

	behindCmd := exec.CommandContext(ctx, "git", "rev-list", "--count", fmt.Sprintf("%s..%s", compareBranch, baseBranch))
	behindCmd.Dir = workspacePath

	behindOutput, err := behindCmd.Output()
	if err != nil {
		summary.CommitsBehind = 0
	} else {
		if count, parseErr := fmt.Sscanf(strings.TrimSpace(string(behindOutput)), "%d", &summary.CommitsBehind); parseErr != nil || count != 1 {
			summary.CommitsBehind = 0
		}
	}

	return nil
}

func getBranchFileDiff(ctx context.Context, workspacePath, baseBranch, compareBranch string, summary *GitDiffSummary, includeContent bool) error {
	statCmd := exec.CommandContext(ctx, "git", "-c", "core.quotepath=false", "diff", "--numstat", fmt.Sprintf("%s..%s", baseBranch, compareBranch))
	statCmd.Dir = workspacePath

	statOutput, err := statCmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			errorMessage := string(exitError.Stderr)
			return fmt.Errorf("git diff --numstat failed: %s", errorMessage)
		}
		return fmt.Errorf("failed to execute git diff --numstat: %v", err)
	}

	files, totalAdditions, totalDeletions := parseNumstat(string(statOutput))
	summary.Files = files
	summary.TotalFiles = len(files)
	summary.TotalAdditions = totalAdditions
	summary.TotalDeletions = totalDeletions

	if includeContent {
		for i := range summary.Files {
			content, err := getFileDiffContent(ctx, workspacePath, baseBranch, compareBranch, summary.Files[i].Path)
			if err != nil {
				Warn("Failed to get diff content for file", "file", summary.Files[i].Path, "error", err)
				continue
			}
			summary.Files[i].DiffContent = content
		}
	}

	return nil
}

func parseNumstat(output string) ([]GitDiffFile, int, int) {
	var files []GitDiffFile
	totalAdditions := 0
	totalDeletions := 0

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}

		file := GitDiffFile{
			Path: parts[2],
		}

		if parts[0] == "-" || parts[1] == "-" {
			file.IsBinary = true
		} else {
			if additions, err := fmt.Sscanf(parts[0], "%d", &file.Additions); err == nil && additions == 1 {
				totalAdditions += file.Additions
			}
			if deletions, err := fmt.Sscanf(parts[1], "%d", &file.Deletions); err == nil && deletions == 1 {
				totalDeletions += file.Deletions
			}
		}

		if file.Additions > 0 && file.Deletions == 0 {
			file.Status = "added"
		} else if file.Additions == 0 && file.Deletions > 0 {
			file.Status = "deleted"
		} else {
			file.Status = "modified"
		}

		files = append(files, file)
	}

	return files, totalAdditions, totalDeletions
}

func getFileDiffContent(ctx context.Context, workspacePath, baseBranch, compareBranch, filePath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-c", "core.quotepath=false", "diff", fmt.Sprintf("%s..%s", baseBranch, compareBranch), "--", filePath)
	cmd.Dir = workspacePath

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			errorMessage := string(exitError.Stderr)
			return "", fmt.Errorf("git diff failed for file %s: %s", filePath, errorMessage)
		}
		return "", fmt.Errorf("failed to execute git diff for file %s: %v", filePath, err)
	}

	return string(output), nil
}

func ValidateBranchExists(workspacePath, branchName string) error {
	if workspacePath == "" {
		return fmt.Errorf("workspace path cannot be empty")
	}

	if branchName == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return fmt.Errorf("workspace directory does not exist: %s", workspacePath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", branchName))
	cmd.Dir = workspacePath

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("branch '%s' does not exist", branchName)
	}

	return nil
}

func GetCommitDiff(workspacePath, commitHash string, includeContent bool) (*GitDiffSummary, error) {
	if workspacePath == "" {
		return nil, fmt.Errorf("workspace path cannot be empty")
	}

	if commitHash == "" {
		return nil, fmt.Errorf("commit hash cannot be empty")
	}

	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace directory does not exist: %s", workspacePath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := validateCommitExists(ctx, workspacePath, commitHash); err != nil {
		return nil, fmt.Errorf("commit validation failed: %v", err)
	}

	summary := &GitDiffSummary{
		Files: []GitDiffFile{},
	}

	if err := getCommitFileDiff(ctx, workspacePath, commitHash, summary, includeContent); err != nil {
		return nil, fmt.Errorf("failed to get commit diff: %v", err)
	}

	return summary, nil
}

func validateCommitExists(ctx context.Context, workspacePath, commitHash string) error {
	cmd := exec.CommandContext(ctx, "git", "cat-file", "-t", commitHash)
	cmd.Dir = workspacePath

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("commit '%s' does not exist", commitHash)
	}

	if strings.TrimSpace(string(output)) != "commit" {
		return fmt.Errorf("'%s' is not a valid commit", commitHash)
	}

	return nil
}

func getCommitFileDiff(ctx context.Context, workspacePath, commitHash string, summary *GitDiffSummary, includeContent bool) error {
	statCmd := exec.CommandContext(ctx, "git", "-c", "core.quotepath=false", "diff", "--numstat", commitHash+"^", commitHash)
	statCmd.Dir = workspacePath

	statOutput, err := statCmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			errorMessage := string(exitError.Stderr)
			return fmt.Errorf("git diff --numstat failed: %s", errorMessage)
		}
		return fmt.Errorf("failed to execute git diff --numstat: %v", err)
	}

	files, totalAdditions, totalDeletions := parseNumstat(string(statOutput))
	summary.Files = files
	summary.TotalFiles = len(files)
	summary.TotalAdditions = totalAdditions
	summary.TotalDeletions = totalDeletions

	if includeContent {
		for i := range summary.Files {
			content, err := getCommitFileDiffContent(ctx, workspacePath, commitHash, summary.Files[i].Path)
			if err != nil {
				Warn("Failed to get diff content for file", "file", summary.Files[i].Path, "error", err)
				continue
			}
			summary.Files[i].DiffContent = content
		}
	}

	return nil
}

func getCommitFileDiffContent(ctx context.Context, workspacePath, commitHash, filePath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-c", "core.quotepath=false", "diff", commitHash+"^", commitHash, "--", filePath)
	cmd.Dir = workspacePath

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			errorMessage := string(exitError.Stderr)
			return "", fmt.Errorf("git diff failed for file %s: %s", filePath, errorMessage)
		}
		return "", fmt.Errorf("failed to execute git diff for file %s: %v", filePath, err)
	}

	return string(output), nil
}

func GetCommitFileDiff(workspacePath, commitHash, filePath string) (string, error) {
	if workspacePath == "" {
		return "", fmt.Errorf("workspace path cannot be empty")
	}

	if commitHash == "" {
		return "", fmt.Errorf("commit hash cannot be empty")
	}

	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := validateCommitExists(ctx, workspacePath, commitHash); err != nil {
		return "", err
	}

	return getCommitFileDiffContent(ctx, workspacePath, commitHash, filePath)
}
