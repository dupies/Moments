import {Suspense} from "react"
import {useFlowBalance} from "../hooks/use-flow-balance.hook"
import {useMomentsBalance} from "../hooks/use-moments-balance.hook"
import {useCurrentUser} from "../hooks/use-current-user.hook"
import {IDLE} from "../global/constants"
import {fmtMoments} from "../util/fmt-moments"
import {
  Box,
  Button,
  Table,
  Tbody,
  Tr,
  Td,
  Flex,
  Heading,
  Spinner,
  Center,
} from "@chakra-ui/react"
import {useInitialized} from "../hooks/use-initialized.hook"

export function BalanceCluster({address}) {
  const flow = useFlowBalance(address)
  const moments = useMomentsBalance(address)
  const init = useInitialized(address)
  return (
    <Box mb="4">
      <Box mb="2">
        <Flex>
          <Heading size="md" mr="4">
            Balances
          </Heading>
          {(flow.status !== IDLE || moments.status !== IDLE) && (
            <Center>
              <Spinner size="sm" />
            </Center>
          )}
        </Flex>
      </Box>
      <Box maxW="200px" borderWidth="1px" borderRadius="lg">
        <Table size="sm">
          <Tbody>
            <Tr>
              <Td>MOMENT</Td>
              {moments.status === IDLE ? (
                <Td isNumeric>{fmtMoments(moments.balance)}</Td>
              ) : (
                <Td isNumeric>
                  <Spinner size="sm" />
                </Td>
              )}
            </Tr>
          </Tbody>
        </Table>
      </Box>
      <Box mt="2">
        <Flex>
          <Button
            colorScheme="blue"
            disabled={moments.status !== IDLE || !init.isInitialized}
            onClick={moments.mint}
          >
            Request Meowr Moments
          </Button>
        </Flex>
      </Box>
    </Box>
  )
}

export default function WrappedBalanceCluster(props) {
  const [cu] = useCurrentUser()
  if (cu.addr !== props.address) return null

  return (
    <Suspense
      fallback={
        <Flex>
          <Heading size="md" mr="4">
            Balances
          </Heading>
          <Center>
            <Spinner size="sm" />
          </Center>
        </Flex>
      }
    >
      <BalanceCluster {...props} />
    </Suspense>
  )
}
