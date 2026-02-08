package inbox

import (
	"fmt"
	"os/exec"
	"runtime"
)

// SendNotification sends an OS notification with the given title and message.
//
// On macOS, it tries terminal-notifier first, then falls back to osascript.
// On Linux, it tries notify-send.
// On other platforms, it is a no-op and returns nil.
func SendNotification(title string, message string) error {
	switch runtime.GOOS {
	case "darwin":
		return sendMacOSNotification(title, message)
	case "linux":
		return sendLinuxNotification(title, message)
	default:
		// No-op for unsupported platforms
		return nil
	}
}

// sendMacOSNotification tries terminal-notifier first, then falls back to osascript.
func sendMacOSNotification(title, message string) error {
	// Try terminal-notifier first
	if path, err := exec.LookPath("terminal-notifier"); err == nil {
		cmd := exec.Command(path, "-title", title, "-message", message)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Fall back to osascript
	script := fmt.Sprintf(`display notification %q with title %q`, message, title)
	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send macOS notification: %w", err)
	}
	return nil
}

// sendLinuxNotification uses notify-send.
func sendLinuxNotification(title, message string) error {
	path, err := exec.LookPath("notify-send")
	if err != nil {
		// notify-send not available, no-op
		return nil
	}

	cmd := exec.Command(path, title, message)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send Linux notification: %w", err)
	}
	return nil
}
