#!/bin/bash

# Danh sách các topic cần tạo
TOPICS=("payment" "payment_events" "order_success" "user.created" "order_returned" "vendor_payment_processed" "vendor_account_updates" "vendor_payments" "bank_payouts" "product-events" "email-events" "user.created.dlq" "product_rating_updates")

# Tạo các topic
for TOPIC in "${TOPICS[@]}"; do
  kafka-topics --create \
    --bootstrap-server kafka:9092 \
    --if-not-exists \
    --partitions 1 \
    --replication-factor 1 \
    --topic "$TOPIC"
done

echo "All topics created successfully!"