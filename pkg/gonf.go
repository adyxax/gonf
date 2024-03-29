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
		// ----- Users -------------------------------------------------
		status = resolveUsers()
		switch status {
		case BROKEN:
			return BROKEN
		case REPAIRED:
			continue
		}
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
		// ----- CustomPromises ----------------------------------------
		status = resolveCustomPromises()
		switch status {
		case BROKEN:
			return BROKEN
		case REPAIRED:
			continue
		}
		return
	}
}
