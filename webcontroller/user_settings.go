package webcontroller

import (
	"fmt"
	"html/template"
	"net/http"

	"fornaxian.tech/pixeldrain_server/api/restapi/apiclient"
	"github.com/Fornaxian/log"
	"github.com/julienschmidt/httprouter"
)

// formAPIError makes it easier to display errors returned by the pixeldrain
// API. TO make use of this function the form fields should be named exactly the
// same as the API parameters
func formAPIError(err error, f *Form) {
	fieldLabel := func(name string) string {
		for _, v := range f.Fields {
			if v.Name == name {
				return v.Label
			}
		}
		return name
	}

	if err, ok := err.(apiclient.Error); ok {
		if err.StatusCode == "multiple_errors" {
			for _, err := range err.Errors {
				// Modify the message to make it more user-friendly
				if err.StatusCode == "string_out_of_range" {
					err.Message = fmt.Sprintf(
						"%s is too long or too short. Should be between %v and %v characters. Current length: %v",
						fieldLabel(err.Extra["field"].(string)),
						err.Extra["min_len"],
						err.Extra["max_len"],
						err.Extra["len"],
					)
				} else if err.StatusCode == "field_contains_illegal_character" {
					err.Message = fmt.Sprintf(
						"Character '%v' is not allowed in %s",
						err.Extra["char"],
						fieldLabel(err.Extra["field"].(string)),
					)
				}

				f.SubmitMessages = append(f.SubmitMessages, template.HTML(err.Message))
			}
		} else {
			f.SubmitMessages = append(f.SubmitMessages, template.HTML(err.Message))
		}
	} else {
		log.Error("Error submitting form: %s", err)
		f.SubmitMessages = []template.HTML{"Internal Server Error"}
	}
}

func (wc *WebController) serveUserSettings(
	w http.ResponseWriter,
	r *http.Request,
	p httprouter.Params,
) {
	td := wc.newTemplateData(w, r)

	if !td.Authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	td.Title = "Account settings"
	td.Other = struct {
		PasswordForm Form
		EmailForm    Form
		UsernameForm Form
	}{
		PasswordForm: wc.passwordForm(td, r),
		EmailForm:    wc.emailForm(td, r),
		UsernameForm: wc.usernameForm(td, r),
	}
	wc.templates.Get().ExecuteTemplate(w, "user_settings", td)
}

func (wc *WebController) passwordForm(td *TemplateData, r *http.Request) (f Form) {
	f = Form{
		Name:  "password_change",
		Title: "Change password",
		Fields: []Field{
			{
				Name:  "old_password",
				Label: "Old Password",
				Type:  FieldTypeCurrentPassword,
			}, {
				Name:  "new_password",
				Label: "New Password",
				Type:  FieldTypeNewPassword,
			}, {
				Name:  "new_password2",
				Label: "New Password again",
				Description: "we need you to repeat your password so you " +
					"won't be locked out of your account if you make a " +
					"typing error",
				Type: FieldTypeNewPassword,
			},
		},
		SubmitLabel: "Submit",
	}

	if f.ReadInput(r) {
		if f.FieldVal("new_password") != f.FieldVal("new_password2") {
			f.SubmitMessages = []template.HTML{
				"Password verification failed. Please enter the same " +
					"password in both new password fields"}
			return f
		}

		// Passwords match, send the request and fill in the response in the
		// form
		if err := td.PixelAPI.UserPasswordSet(
			f.FieldVal("old_password"),
			f.FieldVal("new_password"),
		); err != nil {
			formAPIError(err, &f)
		} else {
			// Request was a success
			f.SubmitSuccess = true
			f.SubmitMessages = []template.HTML{"Success! Your password has been updated"}
		}
	}
	return f
}

func (wc *WebController) emailForm(td *TemplateData, r *http.Request) (f Form) {
	f = Form{
		Name:  "email_change",
		Title: "Change e-mail address",
		Fields: []Field{
			{
				Name:  "new_email",
				Label: "New e-mail address",
				Description: `we will send an e-mail to the new address to
					verify that it's real. The address will be saved once the
					link in the message is clicked. If the e-mail doesn't arrive
					right away please check your spam box too`,
				Type: FieldTypeEmail,
			},
		},
		SubmitLabel: "Submit",
	}

	if f.ReadInput(r) {
		if err := td.PixelAPI.UserEmailReset(
			f.FieldVal("new_email"),
			false,
		); err != nil {
			formAPIError(err, &f)
		} else {
			// Request was a success
			f.SubmitSuccess = true
			f.SubmitMessages = []template.HTML{"Success! E-mail sent"}
		}
	}
	return f
}

func (wc *WebController) serveEmailConfirm(
	w http.ResponseWriter,
	r *http.Request,
	p httprouter.Params,
) {
	var err error
	var status string

	api := apiclient.New(wc.apiURLInternal)
	err = api.UserEmailResetConfirm(r.FormValue("key"))
	if err != nil && err.Error() == "not_found" {
		status = "not_found"
	} else if err != nil {
		log.Error("E-mail reset fail: %s", err)
		status = "internal_error"
	} else {
		status = "success"
	}

	td := wc.newTemplateData(w, r)
	td.Other = status

	wc.templates.Get().ExecuteTemplate(w, "email_confirm", td)
}

func (wc *WebController) usernameForm(td *TemplateData, r *http.Request) (f Form) {
	f = Form{
		Name:  "username_change",
		Title: "Change username",
		Fields: []Field{
			{
				Name:  "new_username",
				Label: "New username",
				Description: "changing your username also changes the name " +
					"used to log in. If you forget your username you can " +
					"still log in using your e-mail address if you have one " +
					"configured",
				Type: FieldTypeUsername,
			},
		},
		SubmitLabel: "Submit",
	}

	if f.ReadInput(r) {
		if err := td.PixelAPI.UserSetUsername(f.FieldVal("new_username")); err != nil {
			formAPIError(err, &f)
		} else {
			// Request was a success
			f.SubmitSuccess = true
			f.SubmitMessages = []template.HTML{template.HTML(
				"Success! You are now " + f.FieldVal("new_username"),
			)}
		}
	}
	return f
}

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
			f.SubmitMessages = []template.HTML{template.HTML("Success!")}
		}
	}
	return f
}
