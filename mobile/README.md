# Gramlane on a phone

One server. iPhone and Android talk to the same desk (`PUBLIC_BASE` or this host).

## What this is

A **PWA**: the website installs as an app. Names, pages, jar grams, Telegram ciphertext, prior stamps are whatever this process stores. No extra API, no extra database.

## Install (ships today)

- **iPhone:** Safari → Share → Add to Home Screen. Open `/app` on the desk for the steps.
- **Android:** Chrome → Install app. Same.

Needs **https** on a real phone (Safari/Chrome rule). `127.0.0.1` only works on that device.

## Store listings (optional wrapper)

This tree is the app. A Capacitor/Cordova shell can load `PUBLIC_BASE` if you want App Store / Play listings. That shell must not invent a second ledger. Point it at the same host.

Kasware is not on iOS. Phone sends that need a wallet use a mobile Kaspa wallet or a pasted `kaspa:` address.
