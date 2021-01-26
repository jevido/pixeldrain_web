package webcontroller

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

func (wc *WebController) patreonLinkForm(td *TemplateData, r *http.Request) (f Form) {
	f.Name = "link_patreon_subscription"
	f.Title = "Link Patreon pledge to pixeldrain account"
	f.SubmitLabel = "Submit"

	if r.FormValue("key") == "" {
		f.Submitted = true
		f.SubmitMessages = []template.HTML{"Patron ID not found"}
		return f
	}

	patron, err := td.PixelAPI.PatreonByID(r.FormValue("key"))
	if err != nil && err.Error() == "not_found" {
		f.Submitted = true
		f.SubmitMessages = []template.HTML{"Patron ID not found"}
		return f
	} else if err != nil {
		f.Submitted = true
		formAPIError(err, &f)
		return f
	}

	f.Fields = []Field{{
		Name:         "1",
		Label:        "",
		DefaultValue: "",
		Description:  "<h3>Please confirm that the following information is correct:</h3>",
		Type:         FieldTypeDescription,
	}, {
		Name:         "2",
		Label:        "Pixeldrain username",
		DefaultValue: td.User.Username,
		Type:         FieldTypeDescription,
	}, {
		Name:         "3",
		Label:        "Pixeldrain e-mail",
		DefaultValue: td.User.Email,
		Type:         FieldTypeDescription,
	}, {
		Name:         "4",
		Label:        "Patreon username",
		DefaultValue: patron.FullName,
		Type:         FieldTypeDescription,
	}, {
		Name:         "5",
		Label:        "Patreon e-mail",
		DefaultValue: patron.UserEmail,
		Type:         FieldTypeDescription,
	}, {
		Name:         "6",
		Label:        "Subscription name",
		DefaultValue: patron.Subscription.Name,
		Type:         FieldTypeDescription,
	}, {
		Name:         "7",
		Label:        "Monthly contribution",
		DefaultValue: fmt.Sprintf("€ %.2f / month", float64(patron.PledgeAmountCents)/100.0),
		Type:         FieldTypeDescription,
	}, {
		Name: "8",
		Description: "When clicking submit your patreon pledge will be linked " +
			"to your pixeldrain account and you will be able to use " +
			"pixeldrain's premium features. If you would like to update or " +
			"cancel your subscription later on you can do so through " +
			"patreon's dashboard",
		Type: FieldTypeDescription,
	}}

	if f.ReadInput(r) {
		if err := td.PixelAPI.PatreonLink(r.FormValue("key")); err != nil {
			formAPIError(err, &f)
		} else {
			// Request was a success
			f.SubmitSuccess = true
			f.SubmitMessages = []template.HTML{template.HTML(
				"Success! Your account has been upgraded to the " + patron.Subscription.Name + " plan.",
			)}
		}
	}
	return f
}

func (wc *WebController) knoxfsLinkForm(td *TemplateData, r *http.Request) (f Form) {
	f.Name = "link_subscription"
	f.Title = "Activate KnoxFS promo"
	f.PreFormHTML = template.HTML(
		`<div style="text-align: center;">
			<img src="/res/img/knoxfs.png" alt="KnoxFS logo" style="max-width: 14em;" />
		</div>`,
	)
	f.SubmitLabel = "Confirm"

	if r.FormValue("key") == "" {
		f.Submitted = true
		f.SubmitMessages = []template.HTML{"Subscription ID not found"}
		return f
	}

	sub, err := td.PixelAPI.SubscriptionByID(r.FormValue("key"))
	if err != nil && err.Error() == "not_found" {
		f.Submitted = true
		f.SubmitMessages = []template.HTML{"Subscription ID not found"}
		return f
	} else if err != nil {
		f.Submitted = true
		formAPIError(err, &f)
		return f
	}

	f.Fields = []Field{{
		Name:         "1",
		Label:        "",
		DefaultValue: "",
		Description:  "<h3>Please confirm that the following information is correct:</h3>",
		Type:         FieldTypeDescription,
	}, {
		Name:         "2",
		Label:        "Pixeldrain username",
		DefaultValue: td.User.Username,
		Type:         FieldTypeDescription,
	}, {
		Name:         "3",
		Label:        "Pixeldrain e-mail",
		DefaultValue: td.User.Email,
		Type:         FieldTypeDescription,
	}, {
		Name:         "4",
		Label:        "Subscription",
		DefaultValue: sub.SubscriptionType.Name,
		Type:         FieldTypeDescription,
	}, {
		Name:         "5",
		Label:        "Duration",
		DefaultValue: fmt.Sprintf("%d days", sub.DurationDays),
		Type:         FieldTypeDescription,
	}, {
		Name:         "6",
		Label:        "End date",
		DefaultValue: time.Now().AddDate(0, 0, sub.DurationDays).Format("2006-01-02"),
		Type:         FieldTypeDescription,
	}, {
		Name: "7",
		Description: "When clicking submit this subscription will be linked " +
			"to your pixeldrain account and you will be able to use " +
			"pixeldrain's pro features. If you already have a pixeldrain " +
			"subscription it will be overwritten",
		Type: FieldTypeDescription,
	}}
	if sub.Used {
		f.Submitted = true
		f.SubmitRed = true
		f.SubmitMessages = []template.HTML{"This subscription is already linked to a pixeldrain account. It can't be linked again"}
		return f
	}

	if f.ReadInput(r) {
		if err := td.PixelAPI.SubscriptionLink(r.FormValue("key")); err != nil {
			formAPIError(err, &f)
		} else {
			// Request was a success
			f.SubmitSuccess = true
			f.SubmitMessages = []template.HTML{template.HTML(
				"Success! Your account has been upgraded to the " + sub.SubscriptionType.Name + " plan.",
			)}
		}
	}
	return f
}
