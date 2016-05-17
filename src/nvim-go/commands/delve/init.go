package delve

import (
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	// Launch
	plugin.HandleCommand("DlvStartServer", &plugin.CommandOptions{NArgs: "*", Eval: "[getcwd(), expand('%:p:h')]", Complete: "file"}, cmdStartServer)
	plugin.HandleCommand("DlvStartClient", &plugin.CommandOptions{Eval: "[getcwd(), expand('%:p:h')]"}, delveStartClient)

	// Command
	plugin.HandleCommand("DlvContinue", &plugin.CommandOptions{}, cmdContinue)
	plugin.HandleCommand("DlvNext", &plugin.CommandOptions{}, cmdNext)
	plugin.HandleCommand("DlvStep", &plugin.CommandOptions{}, cmdStep)
	plugin.HandleCommand("DlvStepInstruction", &plugin.CommandOptions{}, cmdStepInstruction)
	plugin.HandleCommand("DlvRestart", &plugin.CommandOptions{}, cmdRestart)
	plugin.HandleCommand("DlvDisassemble", &plugin.CommandOptions{}, disassemble)
	plugin.HandleCommand("DlvStdin", &plugin.CommandOptions{NArgs: "+"}, cmdStdin)

	// Breokpoint
	plugin.HandleCommand("DlvBreakpoint", &plugin.CommandOptions{NArgs: "+", Complete: "customlist,DelveFunctionList"}, setBreakpoint)
	plugin.HandleFunction("DelveFunctionList", &plugin.FunctionOptions{}, functionList)

	// RPC export
	plugin.Handle("DlvContinue", cmdContinue)
	plugin.Handle("DlvNext", cmdNext)
	plugin.Handle("DlvStep", cmdStep)
	plugin.Handle("DlvStepInstruction", cmdStepInstruction)
	plugin.Handle("DelveStdin", cmdStdin)
	plugin.Handle("DlvRestart", cmdRestart)
	plugin.Handle("DlvDetach", CmdDetach)

	// Exit
	plugin.HandleCommand("DlvDetach", &plugin.CommandOptions{}, CmdDetach)
	plugin.HandleCommand("DlvKill", &plugin.CommandOptions{}, CmdKill)
}

// cmdBuildEval represent a Dlv commands Eval args.
type cmdDelveEval struct {
	Cwd string `msgpack:",array"`
	Dir string
}

// Wrapper function for commands using goroutine.
//
// The advantage is do not freeze the neovim user interface even if any command resulting the busy state.
// Note may become multistage concurrency processing.
//
//  Neovim rpc call (asynchronous)
//    -> Wrapper function (goroutine)
//      -> Remote plugin internal (goroutine)
//        -> neovim-go/vim.Pipeline (goroutine & chan)
func cmdStartServer(v *vim.Vim, args []string, eval cmdDelveEval) {
	go delveStartServer(v, args, eval)
}
func cmdStdin(v *vim.Vim) {
	go stdin(v)
}
func cmdContinue(v *vim.Vim) {
	go cont(v)
}
func cmdNext(v *vim.Vim) {
	go next(v)
}
func cmdStep(v *vim.Vim) {
	go step(v)
}
func cmdStepInstruction(v *vim.Vim) {
	go stepInstruction(v)
}
func cmdRestart(v *vim.Vim) {
	go restart(v)
}
func CmdDetach(v *vim.Vim) {
	go detach(v)
}
func CmdKill(v *vim.Vim) {
	go kill()
}
