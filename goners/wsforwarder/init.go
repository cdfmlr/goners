package wsforwarder

import "golang.org/x/exp/slog"

var logger *slog.Logger

func init() {
	logger = slog.Default().WithGroup("wsforwarder")
}
