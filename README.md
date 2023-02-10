# bankirc

A simple IRC bot that posts bank account transactions to a (private) IRC channel.

Example log after letting the bot run for a week:
```
<bankirc> foobark: 2022-11-15: -300.0 EUR: VIR INST COOL
<bankirc> foobark: 2022-11-17: 10.16 EUR: BAGUETTELAND
<bankirc> foobark: 2022-11-17: -17.5 EUR: CB ONLY ELECTRIC FANS
<bankirc> coolbank: 2022-11-18: -300 EUR: VIR VIREMENT OBAMAS
```

## Why?

I have several bank accounts. I'd like to have a centralized place where I can look up all my past bank transactions
from all my devices.

Hence, a bot that posts new transactions to IRC. Using a bouncer and IRCv3 CHATHISTORY, I can then access all that data
from all my devices.

## Setup

- Register an app on: https://ob.nordigen.com/overview/
- Generate a client ID and client secret on: https://ob.nordigen.com/user-secrets/
- Copy `bankirc.sample.yaml` into `bankirc.yaml` and edit it to set your client ID and client secret

Then, for each of the bank accounts you want to add:
- Find the bank you want to add on: https://ob.nordigen.com/api/v2/institutions/
  - If you get a 401, login again to: https://ob.nordigen.com/overview/
- Copy its ID (the `id` field)
- Run `go run ./cmd/bankirc-init -bank <id> -name <name>`
  - `<id>` is the bank ID from above
  - `<name>` is the friendly bank name you want to give, which will be displayed in transaction messages

Then, run: `go run ./cmd/bankirc`. New transactions will be copied to the IRC channel over time.

## Background

A European law called Payment Services Directive 2 (PSD2) forces bank accounts to expose an API that gives access to
both read-only account savings data and transactions data, and write access to emit transactions. This API is called
OpenBanking and is (more or less) supported by all European banks.

However, to be able to use this API, the consumer needs to be certified as a *Third Party Provider*, which consists of
showing your national banking regulation organism that you are a large company that has the required scale to handle
personal banking data of individuals. This makes it impossible for small open-source projects to connect to banks
directly.

To work around this, [Nordigen](https://nordigen.com/) was created. It acts like a banking information proxy, that
connects to banks API directly because it is certified as a TPP, then exposes the data it gets with its own API. The
important part is that the Nordigen API is free to use without limitation for read-only data, such as transactions.

This is perfect for this project's use case: the bot asks Nordigen for access to your banking data. Nordigen then
requests your bank for access. After you give access to your bank account to Nordigen, the bot gets access to the
banking data.

## License

MIT
