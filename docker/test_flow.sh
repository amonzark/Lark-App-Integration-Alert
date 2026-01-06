#!/bin/bash

ALERTMANAGER_URL="http://localhost:9093/api/v2/alerts" #port forwarding alertmanager 
ALERTMANAGER_URL2="http://localhost:9092/api/v2/alerts" #localhost alertmanager
ALERT_NAME="TestAlertFromKatulampaLarkAppSelectionDuration"
INSTANCE="lark-app-test-selection-duration"
SEVERITY="critical"

ALERT_PAYLOAD=$(cat <<EOF
[
    {
        "labels": {
            "alertname": "$ALERT_NAME",
            "instance": "$INSTANCE",
            "severity": "$SEVERITY"
        },
        "annotations": {
            "summary": "This is just a test alert."
        },
        "status": "firing"
    }
]
EOF
)

echo "Sending initial alert to Alertmanager..."
curl -X POST -H "Content-Type: application/json" -d "$ALERT_PAYLOAD" "$ALERTMANAGER_URL"

echo "Sending initial alert to Alertmanager..."
curl -X POST -H "Content-Type: application/json" -d "$ALERT_PAYLOAD" "$ALERTMANAGER_URL2"
