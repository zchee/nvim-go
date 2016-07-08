package P

func F() {
	foo       // <1> beginning of cun
	if true { // <2> if block
		bar                // <3> beginning of then block
		baz                // <4> second line of then block
		for x := range y { // <5> for block
			quux // <6> inside for
		}
	} else {
		blah // <7> else block
	}
}
