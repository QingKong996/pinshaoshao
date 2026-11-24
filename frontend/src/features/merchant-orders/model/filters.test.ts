import { describe, expect, it } from "vitest"
import { ALL, buildMerchantOrderListQuery } from "./filters"

describe("buildMerchantOrderListQuery", () => {
  it("omits the status when all orders are selected", () => {
    expect(
      buildMerchantOrderListQuery({
        page: 2,
        pageSize: 20,
        statusFilter: ALL,
      }),
    ).toEqual({ page: 2, page_size: 20 })
  })

  it("includes the selected status", () => {
    expect(
      buildMerchantOrderListQuery({
        page: 1,
        pageSize: 10,
        statusFilter: "awaiting_shipment",
      }),
    ).toEqual({
      page: 1,
      page_size: 10,
      status: "awaiting_shipment",
    })
  })
})
