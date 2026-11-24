package logistics

import (
	"testing"
	"time"
)

func TestDeliveryDurationConfig(t *testing.T) {
	t.Setenv("MOCKLOGISTICS_DELIVERY_DURATION", "2m")
	if got := ConfigFromEnv().DeliveryDuration; got != 2*time.Minute {
		t.Fatalf("delivery duration = %s, want 2m", got)
	}

	t.Setenv("MOCKLOGISTICS_DELIVERY_DURATION", "10m")
	if got := ConfigFromEnv().DeliveryDuration; got != maxDeliveryDuration {
		t.Fatalf("clamped delivery duration = %s, want %s", got, maxDeliveryDuration)
	}

	t.Setenv("MOCKLOGISTICS_DELIVERY_DURATION", "")
	if got := (Config{}).withDefaults().DeliveryDuration; got != defaultDeliveryDuration {
		t.Fatalf("default delivery duration = %s, want %s", got, defaultDeliveryDuration)
	}
}
