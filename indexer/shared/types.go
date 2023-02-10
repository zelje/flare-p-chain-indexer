package shared

type Output interface {
	Tx() string    // transaction id of this output
	Index() uint32 // output index
	Addr() string  // address
}

type Input interface {
	OutTx() string    // output transaction id of the input
	OutIndex() uint32 // index of output transaction
	Addr() string     // address

	UpdateAddr(string)
}
