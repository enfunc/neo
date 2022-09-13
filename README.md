[![Go Reference](https://pkg.go.dev/badge/github.com/enfunc/neo.svg)](https://pkg.go.dev/github.com/enfunc/neo)

# Neo

Neo is the unofficial [Neonomics API](https://docs.neonomics.io/api-references/) Go library.

```shell
go get -u github.com/enfunc/neo
```

### A quick run-through

Follow the [Development quickstart](https://docs.neonomics.io/documentation/development/quickstart) to get the client ID, secret ID and the (optional) encryption key. Then, create a client and get the API for a specific device:

```go
client := neo.NewSandboxClient("clientID", "secretID", http.DefaultClient)
api, err := client.API(ctx, "deviceID")
```

If no `err` occurs, you're now authenticated and ready to consume the API. To get the list of all available banks on the platform, do the following:

```go
banks, err := api.Banks(ctx)
```
You can also use `api.BanksByCountryCode`, `api.BanksByName` and `api.BankByID` if you need to be more granular.

To retrieve accounts and process payments, strong customer authentication (SCA) might be required. Have a quick read to fully understand what it entails
in the [official docs](https://docs.neonomics.io/documentation/development/consent). The library is designed to handle SCA on your behalf. You can, however, opt out and handle it by yourself by setting the `API.Mapper` to `nil`.

Here's how you retrieve the accounts:

```go
// Retrieving accounts and processing payments requires a valid session.
// Let's create one.
session, err := api.NewSession(ctx, "bankID")
if err != nil {
	return err
}

// Try to retrieve the accounts using the session ID.
accounts, sca, err := api.Accounts(ctx, session.ID)
if err != nil {
	return err
}
if sca != nil {
	// If sca != nil, it means the bank requires the end-user consent
	// for the Neonomics platform to access the account information on their behalf.
	// You need to propagate the sca.URL to the end user and folow the
	// required steps provided by the website.
	makeEndUserConsentTo(sca.URL)

	// Once the end-user consents to account retrieval, 
	// retry the original request.
	_, err := sca.Retry(ctx, &accounts)
	if err != nil {
		return err
	}
}
// Accounts should now be available and ready to use.
for _, acc := range accounts {
	println(acc.AccountName, acc.IBAN)
}
```

Processing payments works similarly:

```go
payment, sca, err := api.SEPAPayment(ctx, session.ID, &neo.PaymentRequest{  
   DebtorName: "Knut",  
   DebtorAccount: &neo.AccountInfo{  
      IBAN: "NO7013086520592",  
   },  
   CreditorName: "Sven",  
   CreditorAccount: &neo.AccountInfo{  
      IBAN: "SE3750000000054400047881",  
   },  
   RemittanceInfoUnstructured: "test-payment",  
   InstrumentedAmount:         "1.00",  
   Currency:                   "EUR",  
   EndToEndIdentification:     "test-identification",  
   PaymentMetadata: &neo.PaymentMetadata{  
      Address: &neo.Address{  
         StreetName:     "Potetveien",  
         BuildingNumber: "15",  
         PostalCode:     "0150",  
         City:           "Oslo",  
         Country:        "Norway",  
      },  
   },  
})  
if err != nil {  
   return err  
}  
// If sca != nil, the end-user might need to consent to payment initiation and/or payment completion. 
// The library handles the cases behind the scenes, but still requires you to propagate the  
// links to the end-user.
for sca != nil {  
   makeEndUserConsentTo(sca.URL)  
   sca, err = sca.Retry(ctx, payment)  
   if err != nil {  
      return err  
   }  
}
// The library will complete the payment behind the scenes, if required.
println(payment.Status)
```

Some banks require sensitive end-user data (sometimes called Payment Service User information or PSU), such as national identity number, to allow certain operations in their API. Here's how you handle this using the library:

```go
// First, follow the docs here:  
//     https://docs.neonomics.io/documentation/development/sensitive-end-user-data  
// to get the necessary encryption key. Then, create a new Encrypter func  
// by passing in the path to your downloaded key.  
encrypt, err := neo.NewEncrypter("path/to/your/encryption.key")  
if err != nil {  
   return err  
}  
  
// Encrypt the sensitive information.  
ssn, err := encrypt("31125461118")  
if err != nil {  
   return err  
}  
  
// Finally, pass it along the request and handle the accounts
// as demonstrated above.
accounts, sca, err := api.Accounts(ctx, session.ID, neo.PsuID(ssn), neo.PsuIP("109.74.179.3"))  
```

Please open an issue or submit a pull request for any requests, bugs, or comments.

### License

MIT