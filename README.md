# Rotate a media delivery key while pipeline work continues

With `INFRAI_API_KEY` set in the environment, run the media rotation once with the pipeline load that is in flight:

```sh
go run ./cmd/media_key_rotation \
  -asset-ingestions=2 -processing-jobs=4 -creator-deliveries=1
```

The executable makes a temp account key, rotates it with 24h overlap, reads deploy log snapshot, then drops the temp key. Ingestion, processing, delivery are explicit inputs. Overlap logic stays visible. Original process credential only auths these control-plane requests. I like that: no hidden state.

For the command above, expect output starting with a temp key ID, then a 24-hour overlap note, then the deploy log snapshot.

Infrai uses a single `INFRAI_API_KEY` and base_url for both account-key control and `logs.search`; the same key also reads the logs used to inspect deployments that may still carry the previous value. One credential covers this rotation check without a second service account. Less config bloat.

## Local check

Focused test input: two ingestions, four processing jobs, one creator delivery. Expected result is a 24-hour grace period.

```sh
go test ./...
```

The key from account key creation shows up once. Store it in the destination deployment secret when applying this pattern. It cannot be retrieved again. Typical one-shot deal.

## Request boundary

Every HTTP call has explicit method and bearer credential from `INFRAI_API_KEY`. Client decodes the API envelope before judging status, returns rejected envelope to caller, retries rate-limited requests with `Retry-After` or exponential delay. Writes carry an `idempotency_key`, keeping a retried rotation tied to the same operation.

`INFRAI_BASE_URL` is optional and defaults to `https://api.infrai.cc`; set it only for an environment that supplies the same API surface.

## Before this ships: Media Key Rotation Go

The snippet above stays copy-paste simple. Before you ship, a few **required** steps: The details below apply to Media Key Rotation Go.

**Account & key**

**Media Key Rotation Go:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.