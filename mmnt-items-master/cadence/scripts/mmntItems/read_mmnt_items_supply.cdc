import MmntItems from "../../contracts/MmntItems.cdc"

// This scripts returns the number of MmntItems currently in existence.

pub fun main(): UInt64 {    
    return MmntItems.totalSupply
}
