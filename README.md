## CenturyLink Cloud v2 API

This is a Go implementation of the [CLC v2 API](https://www.ctl.io/api-docs/v2).

It requires go >= 1.6 (since this package supports context in combination with client timeout).

### Getting started

Download from inside your `$GOPATH`:
```bash
> go get -d  github.com/grrtrr/clcv2
```

Try some of the examples in the `examples/` folder. These illustrate individual API calls.

### Environment and login

Most examples have help screens (`-h`). The library supports _debug output_ via `-d`.

_Credentials_ can be passed in one of two forms:

1. Via _commandline flags_:
  + `-u <your CLC-Portal-Username>`,
  + `-p <your CLC-Portal-Password>`.

2. Using _environment variables_:
  + `CLC_USER=<CLC-Portal-Username>`,
  + `CLC_PASSWORD=<CLC-Portal-Password>`.

If you are using a _CLC sub-account_, _override the default Account Alias_ via one of

* _Flag_: `-a <AccountAlias>` or
* _Environment_: `CLC_ACCOUNT=<AccountAlias>`

Likewise, to _set a Default Data Centre_ (e.g. `wa1`), use one of

* _Flag_: `-l <your Default DataCentre>` or
* _Environment_: `CLC_LOCATION=<your default DataCentre>`

***Caveat***: Be very careful with the credentials. Using 3 times the wrong username and/or password will cause the account to be locked.
