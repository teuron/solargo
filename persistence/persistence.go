package persistence

import "solargo/inverter"

//GenericDatabase provides an abstraction over a specific database
type GenericDatabase interface {
	//SendData of the inverter to the database
	SendData(data inverter.Data)
}
