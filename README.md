prometheus-gmail-exporter-go
============================

A prometheus exporter for gmail.

Heavily inspired by https://github.com/jamesread/prometheus-gmail-exporter, but written in go instead of Python.

Authentication
--------------

You'll need to set up your own project in GCP and create an OAuth application.

Follow the steps in the [Gmail for Developers Go quickstart](https://developers.google.com/gmail/api/quickstart/go)
to set this up. Store your client credentials in `credentials.json`.

The exporter will use your client credentials to get an access token which lets
it talk to the Gmail API. It will store the token in `token.json`.


