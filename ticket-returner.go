/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

// TicketReturner is functional interface returned by [ModeratorCore.Ticket]
type TicketReturner interface {
	// returnTicket returns a ticket obtained from [ModeratorCore.Ticket]
	ReturnTicket()
}
