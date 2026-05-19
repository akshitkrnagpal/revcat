package products

import "testing"

func TestNeedsTestStorePriceReminderForTitleOnlyProduct(t *testing.T) {
	body := map[string]any{
		"store_identifier": "com.test.monthly",
		"type":             "subscription",
		"title":            "Monthly",
		"app_id":           "app_xxx",
	}

	if !needsTestStorePriceReminder(body) {
		t.Fatal("want reminder for Test Store-style product body with title and no display_name")
	}
}

func TestNeedsTestStorePriceReminderSkipsDisplayNameProduct(t *testing.T) {
	body := map[string]any{
		"store_identifier": "app.monthly",
		"type":             "subscription",
		"display_name":     "Monthly",
		"app_id":           "app_xxx",
	}

	if needsTestStorePriceReminder(body) {
		t.Fatal("do not want reminder for regular store product body with display_name")
	}
}
