package utils

import (
	"encoding/json"
	"fmt"
	"loto-suite/backend/logging"
	"loto-suite/backend/models"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func ScanareBilet(gameId string, imageData []byte) (*models.ScanResult, error) {
	logging.InfoBe("Starting ticket scan")

	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("ticket_%d.jpg", time.Now().UnixNano()))

	if err := os.WriteFile(tempFile, imageData, 0644); err != nil {
		logging.ErrorBe(fmt.Sprintf("Failed to write temp file: %v", err))
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	defer os.Remove(tempFile)

	scanDir, err := getScannerScriptsDir()
	if err != nil {
		logging.ErrorBe(fmt.Sprintf("Failed to find scanner scripts directory: %v", err))
		return nil, err
	}

	scannerPath := filepath.Join(scanDir, fmt.Sprintf("scan_%s.py", gameId))
	cmd := exec.Command(getPythonCommand(), scannerPath, tempFile, "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := fmt.Sprintf("Scanner failed: %v, output: %s", err, string(output))
		logging.ErrorBe(errMsg)
		return nil, fmt.Errorf("scanner failed: %w, output: %s", err, string(output))
	}

	logging.DebugBe(fmt.Sprintf("Scanner output (%d bytes): %s", len(output), string(output)))

	var result models.ScanResult
	if err := json.Unmarshal(output, &result); err != nil {
		errMsg := fmt.Sprintf("Failed to parse scanner output: %v, raw output: %s", err, string(output))
		logging.ErrorBe(errMsg)
		return nil, fmt.Errorf("failed to parse scanner output: %w, raw output: %s", err, string(output))
	}

	// logging.InfoBe(fmt.Sprintf("Parsed result: game_id=%s, variants=%d", result.GameId, len(result.Variante)))

	return &result, nil
}

func getScannerScriptsDir() (string, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	possiblePaths := []string{
		filepath.Join(workDir, "backend", "scan-scripts"),
		filepath.Join(workDir, "..", "backend", "scan-scripts"),
		filepath.Join(workDir, "scan-scripts"),
		filepath.Join(workDir, "..", "scan-scripts"),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("could not find scanner scripts directory (checked from: %s)", workDir)
}

func getPythonCommand() string {
	workDir, err := os.Getwd()
	if err != nil {
		return "python3"
	}

	possibleVenvs := []string{
		filepath.Join(workDir, "venv", "bin", "python"),
		filepath.Join(workDir, "..", "venv", "bin", "python"),
		filepath.Join(workDir, "..", "..", "venv", "bin", "python"),
	}

	for _, venvPython := range possibleVenvs {
		if _, err := os.Stat(venvPython); err == nil {
			return venvPython
		}
	}

	return "python3"
}
