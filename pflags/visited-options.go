/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pflags

import "flag"

type VisitedOptions struct {
	m map[string]bool
	s []string
}

func NewVisitedOptions() (o *VisitedOptions) { return &VisitedOptions{} }

func (o *VisitedOptions) Map() (m map[string]bool) {
	o.m = make(map[string]bool)
	flag.Visit(o.flagVisitFunc)
	m = o.m
	o.m = nil
	return
}

func (o *VisitedOptions) Slice() (s []string) {
	flag.Visit(o.flagVisitFunc)
	s = o.s
	o.s = nil
	return
}

func (o *VisitedOptions) flagVisitFunc(flagFlag *flag.Flag) {
	if m := o.m; m != nil {
		m[flagFlag.Name] = true
	} else {
		o.s = append(o.s, flagFlag.Name)
	}
}
