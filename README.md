# steampipe-plugin-xero

This is a plugin for [steampipe.io](https://steampipe.io/) for reading the [Xero API](https://developer.xero.com/).

Currently, it provides only a `xero_invoice` table which lists all Xero invoices.

## Configuration

Specify configuration details either in the `~/.steampipe/config/xero.spc` file or using the appropriate environment variables.

* `client_id` (`XERO_CLIENT_ID`): the oauth client id of the application registered with Xero
* `client_secret` (`XERO_CLIENT_SECRET`): the oauth client secret of the application registered with Xero
* `oauth_code` (`XERO_OAUTH_CODE`): the code provided in the URL after the oauth flow succeeds
* `tenant_name` (`XERO_TENANTE_NAME`): the organisation name to fetch information for (optional)

## Setting up oauth

1. Create a new Xero app and set up the [oauth code flow configuration](https://developer.xero.com/documentation/guides/oauth2/auth-flow)
2. Provide the plugin with the `client_id` and `client_secret` as per the configuration described above
3. Run steampipe and run a query against xero. This will give you an error and provide a URL to visit. `steampipe query "select * from xero_invoice"`
4. Visit the URL and authorise the plugin access to your Xero organisation.
5. Set the code provided once the authorisation flow completes as the `oauth_code` configuration described above

The plugin will fetch new oauth tokens from Xero as necessary. The tokens will expire after 30 days if unused. In that case, re-run the process above to provide a new oauth code.