package gonf

import (
	"log/slog"
	"os"
)

func EnableDebugLogs() {
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(h))
}

func Resolve() (status Status) {
	for {
		// ----- Files -------------------------------------------------
		status = resolveFiles()
		switch status {
		case BROKEN:
			return BROKEN
		case REPAIRED:
			continue
		}
		// ----- Packages ----------------------------------------------
		status = resolvePackages()
		switch status {
		case BROKEN:
			return BROKEN
		case REPAIRED:
			packages_list_function()
			continue
		}
		// ----- Services ----------------------------------------------
		status = resolveServices()
		switch status {
		case BROKEN:
			return BROKEN
		case REPAIRED:
			continue
		}
		// ----- Commands ----------------------------------------------
		status = resolveCommands()
		switch status {
		case BROKEN:
			return BROKEN
		case REPAIRED:
			continue
		}
		return
	}
}
