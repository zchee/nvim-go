nvim-go contributing memo
=========================

Command
-------

### command completion

Registering the command completion customlist are `Complete` variable in the `plugin.CommandOptions`.
Here is example.

```go
p.HandleCommand(&plugin.CommandOptions{Name: "GoFoo", NArgs: "?", Eval: "expand('%:p')", Complete: "customlist,GoFooCompletion"}, cmdFoo)
```

Neovim's customlist is need list type. See `:help :command-completion-customlist`.
That function should return the completion candidates `[]string` list and `error`.
Note that **must be `error` in return second variable**. If `[]string` variable only, does not works. example is

```go
func cmdFooComplete() ([]string, error) {
	var dir interface{}
	err := c.v.Eval("expand('%:p:h')", &dir)
	if err != nil {
		log.Print("dir")
		return nil, err
	}

	var filelist []string
	files, err := ioutil.ReadDir(dir.(string))
	if err != nil {
		log.Print("ioutil.ReadDir")
		return nil, err
	}
	for _, f := range files {
		filelist = append(filelist, f.Name())
	}

	return filelist, nil
}
```
