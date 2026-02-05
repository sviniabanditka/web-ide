package builtin

import (
	"github.com/webide/ide/backend/internal/ai/tools"
)

func init() {
	tools.GlobalRegistry.Register(ListDir())
	tools.GlobalRegistry.Register(ReadFile())
	tools.GlobalRegistry.Register(SearchInFiles())
	tools.GlobalRegistry.Register(ApplyPatch())
	tools.GlobalRegistry.Register(RunCommand())
	tools.GlobalRegistry.Register(GetCommandOutput())
	tools.GlobalRegistry.Register(CancelCommand())
}
