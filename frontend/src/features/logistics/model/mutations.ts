import { useMutation, useQueryClient } from "@tanstack/react-query"
import { createShipment } from "../api/shipments"
import { shipmentQueryKeys } from "./query-keys"

export function useCreateShipment(token: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (input: Parameters<typeof createShipment>[1]) =>
      createShipment(token, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: shipmentQueryKeys.all(),
      })
    },
  })
}
