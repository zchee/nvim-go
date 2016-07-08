package P

func F(x int) {
	var m map[string]bool
	func() { m = map[string]bool{"a": true}; return }()
	delete(m, "a")
	m["a"+"b"] = !true
IF:
	if x == 1 {
		go func(x int) {}(0)
	} else {
		for i := 0; i < 10; i++ {
			switch x := x.(type) {
			case nil:
			}
		}
	}
}
