package env

func (sf SourceFunc) LookupEnv(s string) (string, bool) {
	return sf(s)
}
