## CenturyLink Cloud v2 API

This is a Go implementation of the [CLC v2 API](https://www.ctl.io/api-docs/v2).

It is _not yet complete_. Contributions are welcome.

### Getting started

Get this package from inside your `$GOPATH`:
```bash
> go get -d  github.com/grrtrr/clcv2
```

Try some of the examples in the `examples/` folder. These illustrate individual API calls.

Most have help screens (`-h`). The library supports _debug output_ via `-d`.

_Credentials_ can be passed in one of two forms:

1. Via _commandline flags_:
  + `-u <your CLC-Portal-Username>`,
  + `-p <your CLC-Portal-Password>`.
2. Using _environment variables_:
  + `CLC_V2_API_USERNAME=<CLC-Portal-Username>`,
  + `CLC_V2_API_PASSWORD=<CLC-Portal-Password>`.

To _override the default Account Alias_, use one of

* flag: `-a <AccountAlias>` or
* environment: `CLC_ACCOUNT=<AccountAlias>`

***Caveat***: Be careful with the credentials. Using 3 times the wrong username and/or password will cause the account to be locked.

### TODO

This is as yet a partial implementation. Also many examples are missing.
After that, refactoring is desirable.
