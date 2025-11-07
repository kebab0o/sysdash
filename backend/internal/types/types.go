package types

import "time"

type CPUPoint struct {
	At time.Time `json:"t"`
	V  float64   `json:"v"`
}

type MemPoint struct {
	At time.Time `json:"t"`
	V  float64   `json:"v"`
}

type DiskPoint struct {
	At      time.Time `json:"t"`
	Mount   string    `json:"mount"`
	UsedPct float64   `json:"usedPct"`
	UsedGB  float64   `json:"usedGB"`
	TotalGB float64   `json:"totalGB"`
}

type DiskIOPoint struct {
	At       time.Time `json:"t"`
	ReadMBs  float64   `json:"readMBs"`
	WriteMBs float64   `json:"writeMBs"`
}

type NetPoint struct {
	At    time.Time `json:"t"`
	RxKBs float64   `json:"rxKBs"`
	TxKBs float64   `json:"txKBs"`
}

type Item struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
