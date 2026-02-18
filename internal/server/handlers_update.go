package server

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/thinkscotty/kibble/internal/updater"
)

func (s *Server) handleUpdateCheck(w http.ResponseWriter, r *http.Request) {
	info, err := updater.CheckForUpdate(r.Context(), s.version)
	if err != nil {
		slog.Error("Update check failed", "error", err)
		fmt.Fprintf(w, `<span class="text-error">Update check failed: %s</span>`,
			template.HTMLEscapeString(err.Error()))
		return
	}

	if info == nil {
		fmt.Fprint(w, `<span class="text-success">You are running the latest version!</span>`)
		return
	}

	// Truncate release notes for display
	notes := info.Body
	if len(notes) > 500 {
		notes = notes[:500] + "..."
	}

	fmt.Fprintf(w, `<div id="update-result">
		<div style="margin-bottom: 0.75rem;">
			<span class="badge badge-active">Update Available</span>
			<strong style="margin-left: 0.5rem;">%s</strong>
		</div>
		<div class="text-muted text-sm" style="margin-bottom: 0.75rem; white-space: pre-line;">%s</div>
		<p class="text-muted text-sm" style="margin-bottom: 0.75rem;">Binary: %s (%s)</p>
		<button type="button" class="btn btn-primary"
				hx-post="/settings/update/install"
				hx-target="#update-result"
				hx-swap="innerHTML"
				hx-confirm="Install update %s? The service will restart automatically.">
			Install Update
		</button>
	</div>`,
		template.HTMLEscapeString(info.TagName),
		template.HTMLEscapeString(notes),
		template.HTMLEscapeString(info.AssetName),
		updater.FormatBytes(info.AssetSize),
		template.HTMLEscapeString(info.TagName),
	)
}

func (s *Server) handleUpdateInstall(w http.ResponseWriter, r *http.Request) {
	// Use a background context for the download (not tied to HTTP request timeout)
	dlCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Re-check for update to get fresh download URL
	info, err := updater.CheckForUpdate(dlCtx, s.version)
	if err != nil {
		slog.Error("Update check failed during install", "error", err)
		fmt.Fprintf(w, `<span class="text-error">Update check failed: %s</span>`,
			template.HTMLEscapeString(err.Error()))
		return
	}
	if info == nil {
		fmt.Fprint(w, `<span class="text-success">Already up to date!</span>`)
		return
	}

	// Download and install
	result, err := updater.DownloadAndInstall(dlCtx, info, s.version)
	if err != nil {
		slog.Error("Update install failed", "error", err)
		fmt.Fprintf(w, `<span class="text-error">Installation failed: %s</span>`,
			template.HTMLEscapeString(err.Error()))
		return
	}

	slog.Info("Update installed successfully",
		"old_version", result.OldVersion,
		"new_version", result.NewVersion)

	fmt.Fprintf(w, `<div>
		<span class="text-success">Update installed successfully!</span>
		<p class="text-muted text-sm" style="margin-top: 0.5rem;">
			Updated from %s to %s. Restarting service...
		</p>
		<p class="text-muted text-sm" id="restart-status">Waiting for restart...</p>
		<script>
		(function() {
			var status = document.getElementById('restart-status');
			var dots = 0;
			var timer = setInterval(function() {
				dots = (dots + 1) %% 4;
				status.textContent = 'Waiting for restart' + '.'.repeat(dots);
			}, 500);
			setTimeout(function poll() {
				fetch(window.location.href, {method: 'HEAD', cache: 'no-store'})
					.then(function(r) {
						if (r.ok) { clearInterval(timer); window.location.reload(); }
						else { setTimeout(poll, 2000); }
					})
					.catch(function() { setTimeout(poll, 2000); });
			}, 5000);
		})();
		</script>
	</div>`,
		template.HTMLEscapeString(result.OldVersion),
		template.HTMLEscapeString(result.NewVersion),
	)

	// Schedule restart after response is sent
	// Give the response time to be sent and received before restarting
	go func() {
		time.Sleep(3 * time.Second)
		slog.Info("Initiating service restart after update")
		if err := updater.RestartService(); err != nil {
			slog.Error("Failed to restart service", "error", err)
		}
	}()
}
