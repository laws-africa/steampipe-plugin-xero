connection "xero" {
  plugin = "local/xero"

  # oauth client id
  # client_id = "12345"

  # oauth client secret
  # client_secret = "ABC123"

  # oauth code for fetching tokens - this is provided in the URL after the oauth flow succeeds
  # oauth_code = "xxx"

  # The name of the Xero organisation (tenant) to fetch information for. If this is unspecified,
  # the first organisation provided by Xero is used.
  # tenant_name = "My Company"

  # Redirect URL for the oauth flow
  # redirect_url = "https://example.com/"

  # path to store oauth token information
  # oauth_token_path = "~/.steampipe/internal/xero-oauth-token.json"
}