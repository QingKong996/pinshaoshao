-- +goose Up
DROP TABLE IF EXISTS mock_payment_notifications CASCADE;
DROP TABLE IF EXISTS mock_payment_orders CASCADE;
DROP TABLE IF EXISTS mock_logistics_notifications CASCADE;
DROP TABLE IF EXISTS mock_logistics_events CASCADE;
DROP TABLE IF EXISTS mock_shipments CASCADE;
DROP TYPE IF EXISTS mock_payment_notification_status;
DROP TYPE IF EXISTS mock_payment_status;
DROP TYPE IF EXISTS mock_logistics_notification_status;
DROP TYPE IF EXISTS mock_shipment_status;

-- +goose Down
SELECT 1;
