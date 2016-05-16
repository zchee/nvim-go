package delve

import (
	"github.com/garyburd/neovim-go/vim"
	"github.com/garyburd/neovim-go/vim/plugin"
)

func init() {
	// Launch
	plugin.HandleCommand("DlvStartServer", &plugin.CommandOptions{NArgs: "*", Eval: "[getcwd(), expand('%:p:h')]", Complete: "file"}, cmdDelveStartServer)
	plugin.HandleCommand("DlvStartClient", &plugin.CommandOptions{Eval: "[getcwd(), expand('%:p:h')]"}, delveStartClient)

	// Command
	plugin.HandleCommand("DlvContinue", &plugin.CommandOptions{}, cmdDelveContinue)
	plugin.HandleCommand("DlvNext", &plugin.CommandOptions{}, cmdDelveNext)
	plugin.HandleCommand("DlvRestart", &plugin.CommandOptions{}, cmdDelveRestart)
	plugin.HandleCommand("DlvDisassemble", &plugin.CommandOptions{}, delveDisassemble)
	plugin.HandleCommand("DlvCommand", &plugin.CommandOptions{NArgs: "+"}, cmdDelveCommand)

	// Breokpoint
	plugin.HandleCommand("DlvBreakpoint", &plugin.CommandOptions{NArgs: "+", Complete: "customlist,DelveFunctionList"}, delveSetBreakpoint)
	plugin.HandleFunction("DelveFunctionList", &plugin.FunctionOptions{}, delveFunctionList)

	// RPC export
	plugin.Handle("DlvContinue", cmdDelveContinue)
	plugin.Handle("DlvNext", cmdDelveNext)
	plugin.Handle("DlvRestart", cmdDelveRestart)
	plugin.Handle("DlvDetach", CmdDelveDetach)

	// Exit
	plugin.HandleCommand("DlvDetach", &plugin.CommandOptions{}, CmdDelveDetach)
	plugin.HandleCommand("DlvKill", &plugin.CommandOptions{}, CmdDelveKill)
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
func cmdDelveStartServer(v *vim.Vim, args []string, eval cmdDelveEval) {
	go delveStartServer(v, args, eval)
}
func cmdDelveCommand(v *vim.Vim, args []string) {
	go delveCommand(v, args)
}
func cmdDelveContinue(v *vim.Vim) {
	go delveContinue(v)
}
func cmdDelveNext(v *vim.Vim) {
	go delveNext(v)
}
func cmdDelveRestart(v *vim.Vim) {
	go delveRestart(v)
}
func CmdDelveDetach(v *vim.Vim) {
	go delveDetach(v)
}
func CmdDelveKill(v *vim.Vim) {
	go delveKill()
}
