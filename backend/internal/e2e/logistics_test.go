//go:build integration

package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	logisticsdto "pinshaoshao/internal/feature/logistics/dto"
	orderdto "pinshaoshao/internal/feature/order/dto"
	paymentdto "pinshaoshao/internal/feature/payment/dto"

	"github.com/google/uuid"
)

func TestLogisticsHTTPFlow(t *testing.T) {
	resetDatabase(t)
	buyerID := insertBuyer(t)
	skuID, _ := insertSellableSKU(t)
	app := startTestApp(t, buyerID)

	var order orderdto.OrderResponse
	app.doJSON(t, http.MethodPost, "/api/v1/orders/direct", map[string]any{
		"checkout_key":     uuid.NewString(),
		"sku_id":           skuID,
		"quantity":         2,
		"receiver_name":    "Logistics Buyer",
		"receiver_phone":   "1234567890",
		"shipping_address": "Hangzhou",
	}, http.StatusCreated, &order)

	var payment paymentdto.OrderPaymentResponse
	app.doJSON(t, http.MethodPost, "/api/v1/payments", map[string]int{
		"order_id": order.ID,
	}, http.StatusCreated, &payment)
	var session paymentdto.OrderPaymentResponse
	app.doJSON(
		t,
		http.MethodPost,
		"/api/v1/payments/"+payment.PaymentNo+"/pay",
		nil,
		http.StatusOK,
		&session,
	)
	response, body := app.do(
		t,
		http.MethodPost,
		session.CheckoutURL+"/pay",
		nil,
		"",
	)
	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("checkout pay status = %d, body = %s", response.StatusCode, body)
	}
	waitFor(t, 10*time.Second, 50*time.Millisecond, func() bool {
		var paid orderdto.OrderResponse
		app.doJSON(
			t,
			http.MethodGet,
			fmt.Sprintf("/api/v1/orders/%d", order.ID),
			nil,
			http.StatusOK,
			&paid,
		)
		return paid.Status == "paid" &&
			paid.SubOrders[0].Status == "awaiting_shipment"
	})

	subOrderID := order.SubOrders[0].ID
	var shipment logisticsdto.ShipmentResponse
	app.doJSON(
		t,
		http.MethodPost,
		"/api/v1/merchant/shipments",
		map[string]any{
			"sub_order_id": subOrderID,
			"carrier":      "顺丰速运",
		},
		http.StatusCreated,
		&shipment,
	)
	if shipment.Status != "pending_pickup" ||
		shipment.Carrier != "顺丰速运" ||
		shipment.TrackingNo == "" ||
		shipment.SubOrderID != subOrderID {
		t.Fatalf("created shipment = %+v", shipment)
	}
	requireInventoryAfterShipment(t, skuID, 8, 0)

	var replayed logisticsdto.ShipmentResponse
	app.doJSON(
		t,
		http.MethodPost,
		"/api/v1/merchant/shipments",
		map[string]any{"sub_order_id": subOrderID},
		http.StatusOK,
		&replayed,
	)
	if replayed.ShipmentNo != shipment.ShipmentNo {
		t.Fatalf(
			"replayed shipment = %s, want %s",
			replayed.ShipmentNo,
			shipment.ShipmentNo,
		)
	}
	requireInventoryAfterShipment(t, skuID, 8, 0)

	var buyerShipments logisticsdto.BuyerShipmentListResponse
	app.doJSON(
		t,
		http.MethodGet,
		fmt.Sprintf("/api/v1/shipments?order_id=%d", order.ID),
		nil,
		http.StatusOK,
		&buyerShipments,
	)
	if len(buyerShipments.Items) != 1 ||
		buyerShipments.Items[0].ShipmentNo != shipment.ShipmentNo {
		t.Fatalf("buyer shipments = %+v", buyerShipments)
	}

	var merchantShipments logisticsdto.ShipmentListResponse
	app.doJSON(
		t,
		http.MethodGet,
		"/api/v1/merchant/shipments?page=1&page_size=10",
		nil,
		http.StatusOK,
		&merchantShipments,
	)
	if merchantShipments.Total != 1 ||
		len(merchantShipments.Items) != 1 {
		t.Fatalf("merchant shipments = %+v", merchantShipments)
	}

	var providerShipmentNo string
	if err := testPool.QueryRow(
		context.Background(),
		`SELECT provider_shipment_no::text FROM shipments WHERE shipment_no = $1::uuid`,
		shipment.ShipmentNo,
	).Scan(&providerShipmentNo); err != nil {
		t.Fatalf("read provider shipment number: %v", err)
	}
	for range 4 {
		response, body = app.do(
			t,
			http.MethodPost,
			"/mocklogistics/api/v1/shipments/"+providerShipmentNo+"/advance",
			nil,
			"",
		)
		if response.StatusCode != http.StatusOK {
			t.Fatalf("advance shipment status = %d, body = %s", response.StatusCode, body)
		}
	}
	response, body = app.do(
		t,
		http.MethodPost,
		"/mocklogistics/api/v1/shipments/"+providerShipmentNo+"/complete",
		nil,
		"",
	)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("complete shipment status = %d, body = %s", response.StatusCode, body)
	}

	var delivered logisticsdto.ShipmentResponse
	waitFor(t, 10*time.Second, 50*time.Millisecond, func() bool {
		app.doJSON(
			t,
			http.MethodGet,
			"/api/v1/shipments/"+shipment.ShipmentNo,
			nil,
			http.StatusOK,
			&delivered,
		)
		return delivered.Status == "delivered"
	})
	if delivered.DeliveredAt == nil || len(delivered.Events) < 6 {
		t.Fatalf("delivered shipment = %+v", delivered)
	}

	var completedOrder orderdto.OrderResponse
	waitFor(t, 5*time.Second, 50*time.Millisecond, func() bool {
		app.doJSON(
			t,
			http.MethodGet,
			fmt.Sprintf("/api/v1/orders/%d", order.ID),
			nil,
			http.StatusOK,
			&completedOrder,
		)
		return completedOrder.Status == "completed" &&
			completedOrder.SubOrders[0].Status == "completed"
	})
}

func requireInventoryAfterShipment(
	t *testing.T,
	skuID int,
	wantTotal int,
	wantFrozen int,
) {
	t.Helper()

	var total, frozen int
	if err := testPool.QueryRow(
		context.Background(),
		`SELECT total_stock, frozen_stock FROM inventory_items WHERE sku_id = $1`,
		skuID,
	).Scan(&total, &frozen); err != nil {
		t.Fatalf("read inventory: %v", err)
	}
	if total != wantTotal || frozen != wantFrozen {
		t.Fatalf(
			"inventory = %d/%d, want %d/%d",
			total,
			frozen,
			wantTotal,
			wantFrozen,
		)
	}
}
