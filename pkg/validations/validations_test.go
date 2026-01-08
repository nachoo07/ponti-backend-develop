package validations

import (
	"regexp"
	"strings"
	"testing"
)

func TestErr(t *testing.T) {
	err := Err("test_field", "test message")
	expected := "test_field: test message"
	if err.Error() != expected {
		t.Errorf("Err() = %v, want %v", err.Error(), expected)
	}
}

func TestJoinErrors(t *testing.T) {
	tests := []struct {
		name     string
		errors   []error
		expected string
		wantNil  bool
	}{
		{
			name:     "no errors",
			errors:   []error{},
			expected: "",
			wantNil:  true,
		},
		{
			name:     "all nil errors",
			errors:   []error{nil, nil, nil},
			expected: "",
			wantNil:  true,
		},
		{
			name:     "single error",
			errors:   []error{Err("field1", "msg1")},
			expected: "field1: msg1",
			wantNil:  false,
		},
		{
			name:     "multiple errors",
			errors:   []error{Err("field1", "msg1"), Err("field2", "msg2")},
			expected: "field1: msg1; field2: msg2",
			wantNil:  false,
		},
		{
			name:     "mixed nil and errors",
			errors:   []error{nil, Err("field1", "msg1"), nil, Err("field2", "msg2")},
			expected: "field1: msg1; field2: msg2",
			wantNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinErrors(tt.errors...)

			if tt.wantNil {
				if result != nil {
					t.Errorf("JoinErrors() = %v, want nil", result)
				}
			} else {
				if result == nil {
					t.Errorf("JoinErrors() = nil, want error")
				} else if result.Error() != tt.expected {
					t.Errorf("JoinErrors() = %v, want %v", result.Error(), tt.expected)
				}
			}
		})
	}
}

func TestValidateRequiredString(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   string
		wantErr bool
	}{
		{"valid string", "name", "hello", false},
		{"empty string", "name", "", true},
		{"whitespace only", "name", "   ", true},
		{"tabs only", "name", "\t\t", true},
		{"mixed whitespace", "name", "  \t  ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequiredString(tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequiredString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStringLen(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   string
		min     int
		max     int
		wantErr bool
	}{
		{"valid length", "name", "hello", 3, 10, false},
		{"exact min", "name", "hi", 2, 10, false},
		{"exact max", "name", "hello world", 3, 11, false},
		{"too short", "name", "hi", 3, 10, true},
		{"too long", "name", "hello world", 3, 10, true},
		{"empty with min 0", "name", "", 0, 10, false},
		{"whitespace trimmed", "name", "  hello  ", 3, 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStringLen(tt.field, tt.value, tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStringLen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSafeString(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   string
		regex   *regexp.Regexp
		wantErr bool
	}{
		{"valid string", "name", "hello123", nil, false},
		{"valid with spaces", "name", "hello world", nil, false},
		{"valid with dots", "name", "hello.world", nil, false},
		{"valid with underscores", "name", "hello_world", nil, false},
		{"valid with hyphens", "name", "hello-world", nil, false},
		{"invalid with special chars", "name", "hello@world", nil, true},
		{"invalid with brackets", "name", "hello[world]", nil, true},
		{"custom regex valid", "name", "HELLO", regexp.MustCompile(`^[A-Z]+$`), false},
		{"custom regex invalid", "name", "hello", regexp.MustCompile(`^[A-Z]+$`), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSafeString(tt.field, tt.value, tt.regex)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSafeString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		email   string
		wantErr bool
	}{
		{"valid email", "email", "test@example.com", false},
		{"valid with dots", "email", "test.name@example.com", false},
		{"valid with hyphens", "email", "test-name@example.com", false},
		{"valid with plus", "email", "test+name@example.com", false},
		{"valid with numbers", "email", "test123@example.com", false},
		{"empty email", "email", "", true},
		{"no @ symbol", "email", "testexample.com", true},
		{"multiple @ symbols", "email", "test@@example.com", true},
		{"with spaces", "email", "test @example.com", true},
		{"invalid format", "email", "test@", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.field, tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		url     string
		wantErr bool
	}{
		{"valid http", "url", "http://example.com", false},
		{"valid https", "url", "https://example.com", false},
		{"valid with path", "url", "https://example.com/path", false},
		{"valid with query", "url", "https://example.com?q=test", false},
		{"valid with port", "url", "https://example.com:8080", false},
		{"empty url", "url", "", true},
		{"no scheme", "url", "example.com", true},
		{"invalid scheme", "url", "ftp://example.com", true},
		{"no host", "url", "https://", true},
		{"malformed", "url", "not-a-url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.field, tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNumeric(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   string
		wantErr bool
	}{
		{"valid numeric", "code", "12345", false},
		{"valid numeric with spaces", "code", "  12345  ", false},
		{"empty string", "code", "", true},
		{"whitespace only", "code", "   ", true},
		{"contains letters", "code", "123abc", true},
		{"contains special chars", "code", "123!@#", true},
		{"contains spaces in middle", "code", "123 45", true},
		{"zero", "code", "0", false},
		{"single digit", "code", "5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNumeric(tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNumeric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateIntRange(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   int64
		min     int64
		max     int64
		wantErr bool
	}{
		{"valid in range", "age", 25, 18, 65, false},
		{"at min", "age", 18, 18, 65, false},
		{"at max", "age", 65, 18, 65, false},
		{"below min", "age", 16, 18, 65, true},
		{"above max", "age", 70, 18, 65, true},
		{"negative", "age", -5, 18, 65, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIntRange(tt.field, tt.value, tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIntRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateFloatRange(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   float64
		min     float64
		max     float64
		wantErr bool
	}{
		{"valid in range", "price", 25.50, 10.0, 100.0, false},
		{"at min", "price", 10.0, 10.0, 100.0, false},
		{"at max", "price", 100.0, 10.0, 100.0, false},
		{"below min", "price", 5.0, 10.0, 100.0, true},
		{"above max", "price", 150.0, 10.0, 100.0, true},
		{"negative", "price", -5.0, 10.0, 100.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFloatRange(tt.field, tt.value, tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFloatRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNonNegative(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   int64
		wantErr bool
	}{
		{"positive", "count", 5, false},
		{"zero", "count", 0, false},
		{"negative", "count", -5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNonNegative(tt.field, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNonNegative() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUUID4(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		uuid    string
		wantErr bool
	}{
		{"valid uuid4", "id", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid uuid4 uppercase", "id", "550E8400-E29B-41D4-A716-446655440000", false},
		{"invalid format", "id", "550e8400-e29b-41d4-a716-44665544000", true},
		{"wrong version", "id", "550e8400-e29b-11d4-a716-446655440000", true},
		{"wrong variant", "id", "550e8400-e29b-41d4-a716-446655440000", false}, // This should pass
		{"empty", "id", "", true},
		{"malformed", "id", "not-a-uuid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUUID4(tt.field, tt.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUUID4() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateULID(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		ulid    string
		wantErr bool
	}{
		{"valid ulid", "id", "01ARZ3NDEKTSV4RRFFQ69G5FAV", false},
		{"valid lowercase", "id", "01arz3ndektsv4rrffq69g5fav", false},
		{"too short", "id", "01ARZ3NDEKTSV4RRFFQ69G5FA", true},
		{"too long", "id", "01ARZ3NDEKTSV4RRFFQ69G5FAVX", true},
		{"invalid chars", "id", "01ARZ3NDEKTSV4RRFFQ69G5FA!", true},
		{"empty", "id", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateULID(tt.field, tt.ulid)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateULID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEnumString(t *testing.T) {
	allowed := []string{"red", "green", "blue"}

	tests := []struct {
		name    string
		field   string
		value   string
		allowed []string
		wantErr bool
	}{
		{"valid value", "color", "red", allowed, false},
		{"case sensitive", "color", "Red", allowed, true},
		{"not in list", "color", "yellow", allowed, true},
		{"empty value", "color", "", allowed, true},
		{"empty allowed list", "color", "red", []string{}, true},
		{"nil allowed list", "color", "red", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnumString(tt.field, tt.value, tt.allowed...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEnumString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateStringSliceNotEmpty(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		slice   []string
		wantErr bool
	}{
		{"valid slice", "tags", []string{"tag1", "tag2"}, false},
		{"single item", "tags", []string{"tag1"}, false},
		{"empty slice", "tags", []string{}, true},
		{"nil slice", "tags", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStringSliceNotEmpty(tt.field, tt.slice)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStringSliceNotEmpty() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUniqueStrings(t *testing.T) {
	tests := []struct {
		name            string
		field           string
		slice           []string
		caseInsensitive bool
		wantErr         bool
	}{
		{"unique strings", "tags", []string{"tag1", "tag2", "tag3"}, false, false},
		{"duplicate strings", "tags", []string{"tag1", "tag2", "tag1"}, false, true},
		{"case sensitive unique", "tags", []string{"Tag1", "tag1"}, false, false},
		{"case insensitive duplicate", "tags", []string{"Tag1", "tag1"}, true, true},
		{"single item", "tags", []string{"tag1"}, false, false},
		{"nil slice", "tags", nil, false, false},
		{"empty slice", "tags", []string{}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUniqueStrings(tt.field, tt.slice, tt.caseInsensitive)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUniqueStrings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSliceLen(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		length  int
		min     int
		max     int
		wantErr bool
	}{
		{"valid length", "items", 5, 1, 10, false},
		{"at min", "items", 1, 1, 10, false},
		{"at max", "items", 10, 1, 10, false},
		{"below min", "items", 0, 1, 10, true},
		{"above max", "items", 15, 1, 10, true},
		{"negative length", "items", -1, 1, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSliceLen(tt.field, tt.length, tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSliceLen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateISODate(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		date    string
		wantErr bool
	}{
		{"valid date", "birth_date", "1990-01-15", false},
		{"valid date with zeros", "birth_date", "2000-12-31", false},
		{"invalid format", "birth_date", "1990/01/15", true},
		{"invalid date", "birth_date", "1990-13-45", true},
		{"empty string", "birth_date", "", true},
		{"wrong separator", "birth_date", "1990.01.15", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateISODate(tt.field, tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateISODate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateISOTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		timestamp string
		wantErr   bool
	}{
		{"valid RFC3339", "created_at", "2023-01-15T10:30:00Z", false},
		{"valid RFC3339Nano", "created_at", "2023-01-15T10:30:00.123456789Z", false},
		{"invalid format", "created_at", "2023-01-15 10:30:00", true},
		{"empty string", "created_at", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateISOTimestamp(tt.field, tt.timestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateISOTimestamp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDateRange(t *testing.T) {
	tests := []struct {
		name       string
		startField string
		start      string
		endField   string
		end        string
		wantErr    bool
	}{
		{"valid range", "start_date", "2023-01-01", "end_date", "2023-01-31", false},
		{"same date", "start_date", "2023-01-15", "end_date", "2023-01-15", false},
		{"invalid start after end", "start_date", "2023-01-31", "end_date", "2023-01-01", true},
		{"invalid start date", "start_date", "invalid", "end_date", "2023-01-31", true},
		{"invalid end date", "start_date", "2023-01-01", "end_date", "invalid", true},
		{"mixed date and timestamp", "start_date", "2023-01-01", "end_date", "2023-01-31T23:59:59Z", false},
		{"mixed timestamp and date", "start_date", "2023-01-01T00:00:00Z", "end_date", "2023-01-31", false},
		{"both timestamps", "start_date", "2023-01-01T00:00:00Z", "end_date", "2023-01-31T23:59:59Z", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDateRange(tt.startField, tt.start, tt.endField, tt.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDateRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNotFuture(t *testing.T) {
	// This test is time-dependent, so we'll test with a known past date
	pastDate := "2020-01-01T00:00:00Z"

	tests := []struct {
		name    string
		field   string
		date    string
		wantErr bool
	}{
		{"past date", "created_at", pastDate, false},
		{"invalid format", "created_at", "invalid", true},
		{"empty string", "created_at", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotFuture(tt.field, tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNotFuture() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateContentType(t *testing.T) {
	allowed := []string{"application/json", "text/plain"}

	tests := []struct {
		name    string
		field   string
		ct      string
		allowed []string
		wantErr bool
	}{
		{"valid content type", "content_type", "application/json", allowed, false},
		{"valid with parameters", "content_type", "application/json; charset=utf-8", allowed, false},
		{"case insensitive", "content_type", "APPLICATION/JSON", allowed, false},
		{"not in allowed list", "content_type", "text/html", allowed, true},
		{"empty content type", "content_type", "", allowed, true},
		{"no restrictions", "content_type", "any/type", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContentType(tt.field, tt.ct, tt.allowed...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateContentType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateBearerFormat(t *testing.T) {
	tests := []struct {
		name          string
		field         string
		auth          string
		wantErr       bool
		expectedToken string
	}{
		{"valid bearer", "authorization", "Bearer token123", false, "token123"},
		{"valid bearer with spaces", "authorization", "Bearer  token123  ", false, "token123"},
		{"wrong scheme", "authorization", "Basic token123", true, ""},
		{"no space", "authorization", "Bearertoken123", true, ""},
		{"empty token", "authorization", "Bearer ", true, ""},
		{"empty auth", "authorization", "", true, ""},
		{"case insensitive", "authorization", "bearer token123", false, "token123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ValidateBearerFormat(tt.field, tt.auth)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBearerFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && token != tt.expectedToken {
				t.Errorf("ValidateBearerFormat() token = %v, want %v", token, tt.expectedToken)
			}
		})
	}
}

func TestValidateIdempotencyKey(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		key     string
		wantErr bool
	}{
		{"valid key", "idempotency_key", "abc123def", false},
		{"valid key at min length", "idempotency_key", "12345678", false},
		{"valid key at max length", "idempotency_key", strings.Repeat("a", 128), false},
		{"too short", "idempotency_key", "1234567", true},
		{"too long", "idempotency_key", strings.Repeat("a", 129), true},
		{"with spaces", "idempotency_key", "abc 123", true},
		{"with tabs", "idempotency_key", "abc\t123", true},
		{"empty", "idempotency_key", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIdempotencyKey(tt.field, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIdempotencyKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCurrencyISO4217(t *testing.T) {
	allowed := map[string]struct{}{"USD": {}, "EUR": {}, "GBP": {}}

	tests := []struct {
		name    string
		field   string
		code    string
		allowed map[string]struct{}
		wantErr bool
	}{
		{"valid USD", "currency", "USD", allowed, false},
		{"valid EUR", "currency", "EUR", allowed, false},
		{"not in allowed list", "currency", "JPY", allowed, true},
		{"no restrictions", "currency", "JPY", nil, false},
		{"too short", "currency", "US", allowed, true},
		{"too long", "currency", "USDD", allowed, true},
		{"lowercase", "currency", "usd", allowed, true},
		{"with numbers", "currency", "US1", allowed, true},
		{"empty", "currency", "", allowed, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCurrencyISO4217(tt.field, tt.code, tt.allowed)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCurrencyISO4217() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMonetaryCents(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		cents   int64
		wantErr bool
	}{
		{"positive", "amount", 1000, false},
		{"zero", "amount", 0, false},
		{"negative", "amount", -100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMonetaryCents(tt.field, tt.cents)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMonetaryCents() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRequireNotNil(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		ptr     *string
		wantErr bool
	}{
		{"not nil", "name", stringPtr("hello"), false},
		{"nil", "name", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequireNotNil(tt.field, tt.ptr)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequireNotNil() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateOptional(t *testing.T) {
	tests := []struct {
		name      string
		ptr       *string
		validator func(string) error
		wantErr   bool
	}{
		{"nil pointer", nil, func(s string) error { return Err("test", "should not run") }, false},
		{"not nil valid", stringPtr("hello"), func(s string) error { return nil }, false},
		{"not nil invalid", stringPtr("hello"), func(s string) error { return Err("test", "validation failed") }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOptional(tt.ptr, tt.validator)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOptional() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizePagination(t *testing.T) {
	tests := []struct {
		name      string
		page      int
		limit     int
		wantPage  int
		wantLimit int
	}{
		{"valid values", 5, 25, 5, 25},
		{"zero page", 0, 25, 1, 25},
		{"zero limit", 5, 0, 5, 20},
		{"both zero", 0, 0, 1, 20},
		{"negative page", -5, 25, 1, 25},
		{"negative limit", 5, -25, 5, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, limit := NormalizePagination(tt.page, tt.limit)
			if page != tt.wantPage {
				t.Errorf("NormalizePagination() page = %v, want %v", page, tt.wantPage)
			}
			if limit != tt.wantLimit {
				t.Errorf("NormalizePagination() limit = %v, want %v", limit, tt.wantLimit)
			}
		})
	}
}

func TestValidatePagination(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		limit    int
		maxLimit int
		wantErr  bool
	}{
		{"valid pagination", 1, 20, 100, false},
		{"page at minimum", 1, 20, 100, false},
		{"limit at maximum", 5, 100, 100, false},
		{"page below minimum", 0, 20, 100, true},
		{"page below minimum", -1, 20, 100, true},
		{"limit below minimum", 1, 0, 100, true},
		{"limit below minimum", 1, -5, 100, true},
		{"limit above maximum", 1, 150, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePagination(tt.page, tt.limit, tt.maxLimit)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePagination() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSort(t *testing.T) {
	whitelist := map[string]struct{}{"name": {}, "age": {}, "created_at": {}}

	tests := []struct {
		name      string
		field     string
		sort      string
		whitelist map[string]struct{}
		wantErr   bool
	}{
		{"valid sort field", "sort", "name", whitelist, false},
		{"not in whitelist", "sort", "invalid", whitelist, true},
		{"no restrictions", "sort", "any_field", nil, false},
		{"empty sort", "sort", "", whitelist, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSort(tt.field, tt.sort, tt.whitelist)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSort() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSortDir(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		dir     string
		wantErr bool
	}{
		{"valid asc", "direction", "asc", false},
		{"valid desc", "direction", "desc", false},
		{"case insensitive asc", "direction", "ASC", false},
		{"case insensitive desc", "direction", "DESC", false},
		{"invalid direction", "direction", "invalid", true},
		{"empty direction", "direction", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSortDir(tt.field, tt.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSortDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}
