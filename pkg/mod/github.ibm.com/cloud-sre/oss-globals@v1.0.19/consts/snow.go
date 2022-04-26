package consts

const (
	// PotentialCIE SN incident status "potential-cie" potential custumer impact event
	PotentialCIE = "potential-cie"
	// ConfirmedCIE SN incident status "confirmed-cie" confirmed custumer impact event
	ConfirmedCIE = "confirmed-cie"
	NormalIncSts = "normal"
	// CreateReq SN create message type
	CreateReq = "create.request"
	// CloseReq SN close message type
	CloseReq = "close.request"
	// UpdReq SN update message type
	UpdReq = "update.request"
	// CustImpactingDefVal SN customer impacting default value
	CustImpactingDefVal = "false"
)

var (
	//CustImpactVal SN valid custumer impacting options
	CustImpactVal = map[string]bool{
		"true":       true,
		"false":      true,
		PotentialCIE: true,
		"0":          true,
		"1":          true,
	}

	//SNstatusVal PnP  to SN status map
	SNstatusVal = map[string]string{
		NormalIncSts: "22",
		PotentialCIE: "20",
		ConfirmedCIE: "21",
	}
)
