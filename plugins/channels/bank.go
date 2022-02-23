package channels

type BankChannel interface {
	Authorize ()
	PreAuthorize()
	Confirm()
	Reverse()
	Refund()
	Complete3DS()
}

var BankChannelType int = 1
