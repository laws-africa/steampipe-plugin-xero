package xero

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
)

func tableXeroInvoice(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "xero_invoice",
		Description: "Invoices and bills.",
		List: &plugin.ListConfig{
			Hydrate:    listInvoices,
			KeyColumns: plugin.SingleColumn("invoice_id"),
		},
		Columns: []*plugin.Column{
			{Name: "invoice_id", Type: proto.ColumnType_STRING, Transform: transform.FromField("InvoiceID")},
			{Name: "type", Type: proto.ColumnType_STRING, Transform: transform.FromField("Type")},
		},
	}
}

func listInvoices(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	return nil, nil
}
