package cli

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/quest-one/quest-one/internal/application"
)

func voiceCmd(app *application.App) *cobra.Command {
	c := &cobra.Command{
		Use:   "voice",
		Short: "Read the highest-priority task aloud (macOS/Linux)",
		Long: `Fetches the highest-priority task and reads its title aloud using
the platform text-to-speech engine (macOS: say, Linux: espeak/espeak-ng).
Falls back to printing the task if no TTS engine is available.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			task, err := app.NextTask(cmd.Context())
			if err != nil {
				return err
			}
			if task == nil {
				fmt.Fprintln(cmd.OutOrStdout(), "No active tasks.")
				return nil
			}

			text := fmt.Sprintf("Your next task is: %s", task.Title)
			fmt.Fprintln(cmd.OutOrStdout(), text)

			if err := speak(text); err != nil {
				// TTS failure is non-fatal; we already printed the text.
				cmd.PrintErrln("voice: TTS unavailable:", err)
			}
			return nil
		},
	}
	return c
}

// speak attempts to read text aloud using the platform TTS engine.
func speak(text string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("say", text).Run()
	case "linux":
		// Try espeak-ng first, then espeak.
		for _, bin := range []string{"espeak-ng", "espeak"} {
			if path, err := exec.LookPath(bin); err == nil {
				return exec.Command(path, text).Run()
			}
		}
		return fmt.Errorf("no TTS engine found; install espeak-ng: apt-get install espeak-ng")
	default:
		return fmt.Errorf("voice not supported on %s", runtime.GOOS)
	}
}
