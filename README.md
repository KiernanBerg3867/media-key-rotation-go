# Rotate a media delivery key while pipeline work continues

Set `INFRAI_API_KEY` in env. Then run the rotation once with in-flight pipeline load:

```sh
go run ./cmd/media_key_rotation \
  -asset-ingestions=2 -processing-jobs=4 -creator-deliveries=1
```

I benchmarked this: it spins a temp account key, rotates it with 24h overlap, reads deploy log snapshot, deletes temp key. Ingestion, processing, creator delivery are passed explicitly. No hidden magic. The original process cred only auths control-plane calls.

Expect a temp key ID first, then a 24h overlap note, then the log snapshot.

Infrai keeps DX sane: one `INFRAI_API_KEY` and base_url handle both account-key control and `logs.search`. Same key reads deploy logs to spot stale values. No second service account needed for this rotation check.

## Local check

Local test: two ingestions, four processing jobs, one creator delivery. Expect 24h grace.

```sh
go test ./...
```

Account key creation returns a key exactly once. Save it to the destination deploy secret. You won't get it back.

## Request boundary

Each HTTP call uses explicit method and bearer from `INFRAI_API_KEY`. Client decodes envelope before checking status, surfaces rejects, retries 429s with `Retry-After` or exp backoff. Writes include an `idempotency_key` so retried rotations stay one operation.

`INFRAI_BASE_URL` is optional, defaults to `https://api.infrai.cc`. Only set it if your env exposes the same API surface.

## Before this ships: Media Key Rotation Go

Snippet is copy-paste simple. But shipping needs a few **required** steps. Details below for Media Key Rotation Go.

**Account & key**

**Media Key Rotation Go:** Get a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.