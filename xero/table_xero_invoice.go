package xero

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
)

func tableXeroInvoice(ctx context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "xero_invoice",
		Description: "Invoices and bills.",
		List: &plugin.ListConfig{
			Hydrate: listInvoices,
			// KeyColumns: plugin.SingleColumn("invoice_id"),
		},
		Columns: []*plugin.Column{
			{Name: "type", Type: proto.ColumnType_STRING},
			{Name: "invoice_id", Type: proto.ColumnType_STRING},
			{Name: "invoice_number", Type: proto.ColumnType_STRING},
			{Name: "reference", Type: proto.ColumnType_STRING},
			{Name: "amount_paid", Type: proto.ColumnType_DOUBLE},
			{Name: "amount_due", Type: proto.ColumnType_DOUBLE},
			{Name: "amount_credited", Type: proto.ColumnType_DOUBLE},
			{Name: "url", Type: proto.ColumnType_STRING},
			{Name: "currency_rate", Type: proto.ColumnType_DOUBLE},
			{Name: "is_discounted", Type: proto.ColumnType_BOOL},
			{Name: "has_attachments", Type: proto.ColumnType_BOOL},
			{Name: "has_errors", Type: proto.ColumnType_BOOL},
			{Name: "date_string", Type: proto.ColumnType_STRING},
			{Name: "due_date_string", Type: proto.ColumnType_STRING},
			{Name: "status", Type: proto.ColumnType_STRING},
			{Name: "line_amount_types", Type: proto.ColumnType_STRING},
			{Name: "sub_total", Type: proto.ColumnType_DOUBLE},
			{Name: "total_tax", Type: proto.ColumnType_DOUBLE},
			{Name: "total", Type: proto.ColumnType_DOUBLE},
			{Name: "updated_date_utc", Type: proto.ColumnType_STRING},
			{Name: "currency_code", Type: proto.ColumnType_STRING},
			{Name: "contact_id", Type: proto.ColumnType_STRING, Transform: transform.FromField("Contact.ContactID")},
			{Name: "contact_name", Type: proto.ColumnType_STRING, Transform: transform.FromField("Contact.Name")},
		},
	}
}

func listInvoices(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	client, err := connect(ctx, d)
	if err != nil {
		plugin.Logger(ctx).Error("xero.listInvoices", "connection_error", err)
		return nil, err
	}

	// TODO: config
	tenantId, err := getTenantId(ctx, "Laws.Africa")
	if err != nil {
		plugin.Logger(ctx).Error("xero.listInvoices", "getTenantId", err)
		return nil, err
	}

	url := "https://api.xero.com/api.xro/2.0/invoices"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("xero-tenant-id", tenantId)
	resp, err := client.Do(req)

	if err != nil {
		plugin.Logger(ctx).Error("xero.listInvoices", "invoices", err)
		return nil, err
	}

	if resp.StatusCode == 200 {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)

		response := Invoices{}
		err = json.Unmarshal(body, &response)
		if err != nil {
			plugin.Logger(ctx).Error("xero.listInvoices", "unmarshal", err)
			return nil, err
		}

		for _, t := range response.Invoices {
			d.StreamListItem(ctx, t)
		}
	} else {
		plugin.Logger(ctx).Error("xero.listInvoices", "unmarshal", err)
		return nil, fmt.Errorf("error querying xero api: %v", resp.Status)
	}

	return nil, nil
}

type Invoice struct {
	Type           string        `json:"Type"`
	InvoiceID      string        `json:"InvoiceID"`
	InvoiceNumber  string        `json:"InvoiceNumber"`
	Reference      string        `json:"Reference"`
	Payments       []interface{} `json:"Payments"`
	CreditNotes    []interface{} `json:"CreditNotes"`
	Prepayments    []interface{} `json:"Prepayments"`
	Overpayments   []interface{} `json:"Overpayments"`
	AmountDue      float64       `json:"AmountDue"`
	AmountPaid     float64       `json:"AmountPaid"`
	AmountCredited float64       `json:"AmountCredited"`
	URL            string        `json:"Url"`
	CurrencyRate   float64       `json:"CurrencyRate"`
	IsDiscounted   bool          `json:"IsDiscounted"`
	HasAttachments bool          `json:"HasAttachments"`
	HasErrors      bool          `json:"HasErrors"`
	Contact        struct {
		ContactID           string        `json:"ContactID"`
		Name                string        `json:"Name"`
		Addresses           []interface{} `json:"Addresses"`
		Phones              []interface{} `json:"Phones"`
		ContactGroups       []interface{} `json:"ContactGroups"`
		ContactPersons      []interface{} `json:"ContactPersons"`
		HasValidationErrors bool          `json:"HasValidationErrors"`
	} `json:"Contact"`
	DateString      string        `json:"DateString"`
	DueDateString   string        `json:"DueDateString"`
	Status          string        `json:"Status"`
	LineAmountTypes string        `json:"LineAmountTypes"`
	LineItems       []interface{} `json:"LineItems"`
	SubTotal        float64       `json:"SubTotal"`
	TotalTax        float64       `json:"TotalTax"`
	Total           float64       `json:"Total"`
	UpdatedDateUTC  string        `json:"UpdatedDateUTC"`
	CurrencyCode    string        `json:"CurrencyCode"`
}

type Invoices struct {
	Invoices []Invoice `json:"Invoices"`
}
