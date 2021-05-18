import Moment from "../../contracts/Moment.cdc"

// This script returns the total amount of Moment currently in existence.

pub fun main(): UFix64 {

    let supply = Moment.totalSupply

    log(supply)

    return supply
}
