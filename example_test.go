package hook_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/BenLubar/hook"
)

func Example() {
	var (
		beforeRequest func(*http.Request, http.Header) (*http.Request, error)
		BeforeRequest = hook.NewFilter(&beforeRequest).(func(func(*http.Request, http.Header) (*http.Request, error), int))
	)

	type printerValue struct{}
	matcher := language.NewMatcher([]language.Tag{language.AmericanEnglish, language.French, language.German, language.Spanish})
	_ = message.SetString(language.French, "Hello, world!\n", "Bonjour le monde!\n")
	_ = message.SetString(language.German, "Hello, world!\n", "Hallo Welt!\n")
	_ = message.SetString(language.Spanish, "Hello, world!\n", "¡Hola Mundo!\n")

	BeforeRequest(func(r *http.Request, headers http.Header) (*http.Request, error) {
		headers.Add("Vary", "Accept-Language")
		tags, _, err := language.ParseAcceptLanguage(r.Header.Get("Accept-Language"))
		if err != nil {
			return r, err
		}
		tag, _, _ := matcher.Match(tags...)
		return r.WithContext(context.WithValue(r.Context(), printerValue{}, message.NewPrinter(tag))), nil
	}, -100)

	PrinterForRequest := func(r *http.Request) *message.Printer {
		return r.Context().Value(printerValue{}).(*message.Printer)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r, err := beforeRequest(r, w.Header())
		if err != nil {
			panic(err)
		}

		_, _ = PrinterForRequest(r).Fprintf(w, "Hello, world!\n")
	}))
	defer srv.Close()

	req, err := http.NewRequest("GET", srv.URL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept-Language", "es")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		panic(resp.Status)
	}

	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		panic(err)
	}

	// Output: ¡Hola Mundo!
}
