# E2E Tests

End-to-end tests that run against a real Rollbar account. These tests verify the CLI works correctly with the actual Rollbar API.

## Setup

1. Get a read token from your Rollbar project:
   - Go to Project Settings > Access Tokens
   - Create or use an existing token with "read" scope

2. Set environment variables:
   ```bash
   export ROLLBAR_E2E_TOKEN="your-read-token"

   # Optional: specify an item counter for detailed tests
   export ROLLBAR_E2E_ITEM_COUNTER="123"
   ```

3. Run the tests:
   ```bash
   make test-e2e
   # or directly:
   go test -tags=e2e -v ./e2e/...
   ```

## Test Cases

- `TestE2E_ItemsList` - List items, verify response structure
- `TestE2E_ItemsWithFilters` - Test --level, --status, --env filters
- `TestE2E_ItemsTimeFilters` - Test --since, --from, --to filters
- `TestE2E_ItemsSearch` - Test --query text search
- `TestE2E_ItemsSorting` - Test --sort options
- `TestE2E_ItemDetail` - Get single item by counter
- `TestE2E_Occurrences` - List occurrences for item
- `TestE2E_OccurrenceDetail` - Get single occurrence
- `TestE2E_Context` - Generate context markdown
- `TestE2E_OutputFormats` - Test --output json/table/compact/markdown
- `TestE2E_InvalidToken` - Verify auth error handling

## Notes

- E2E tests are NOT run in CI/CD (via build tags)
- Run these manually before releases
- Tests require a Rollbar project with existing items/errors
