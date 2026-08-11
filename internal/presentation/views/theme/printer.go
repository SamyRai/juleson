package theme

import (
	"fmt"
)

func PrintSuccess(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(SuccessStyle.Render("✅ " + msg))
}

func PrintError(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(ErrorStyle.Render("❌ Error: " + msg))
}

func PrintWarning(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(WarnStyle.Render("⚠️ " + msg))
}

func PrintInfo(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(InfoStyle.Render("ℹ️ " + msg))
}

func PrintHeader(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("\n%s\n", HeaderStyle.Render(msg))
}

func PrintStep(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(StepStyle.Render("🔨 " + msg))
}

func PrintMuted(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(MutedStyle.Render(msg))
}
