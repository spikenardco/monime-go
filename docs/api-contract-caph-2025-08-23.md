# Monime API Contract: `caph.2025-08-23`

This document freezes the versioned Monime API contract used by this SDK. It
was retrieved on 2026-09-03. The versioned endpoint pages are authoritative;
the older aggregate OpenAPI document and `monimejs` are not used to choose
wire shapes.

## Shared protocol

All requests use `https://api.monime.io/v1`. `Authorization: Bearer <token>`
and `Monime-Space-Id` are required. `Monime-Version: caph.2025-08-23` pins the
contract. JSON requests also require `Content-Type: application/json`.

Every creating or side-effecting POST requires `Idempotency-Key`; it is scoped
to a Space, cached for successful responses for 24 hours, and must be reused
for a retry. A key reused with a different URL or body returns `409` with
reason `idempotency_key_in_use`. POST and GET can be retried for transient
network failures and HTTP `429`, `500`, `502`, `503`, and `504`; PATCH and
DELETE have no documented replay guarantee.

Successful object responses use `{success, messages, result}`. Lists add
`pagination: {count, next}`, where `next` may be null or absent. API errors
use `{success: false, messages, error: {code, reason, message, details}}`.
`Retry-After` is seconds or an HTTP date; a `429` also carries
`Monime-Rate-Limit`. Diagnostic response headers include `Monime-Request-Id`,
`Monime-Request-Checksum`, `Monime-Request-Timestamp`,
`Monime-Request-Duration`, and `Monime-Cache`.

All list endpoints support `limit` (1–50) and opaque `after`; `next` is the
only end-of-list signal.

## Operation matrix

`object` below means the resource object described in the next section,
`list` means an array of those objects, and `empty` accepts a successful empty
body or an envelope whose result is ignored.

| Service | Operation | Method and path | Query / request fields | Result |
| --- | --- | --- | --- | --- |
| Banks | List | `GET /banks` | `country`, `limit`, `after` | list |
| Banks | Get | `GET /banks/{providerId}` | path `providerId` | object |
| Mobile money | List | `GET /momos` | `country`, `limit`, `after` | list |
| Mobile money | Get | `GET /momos/{providerId}` | path `providerId` | object |
| Provider KYC | Get | `GET /provider-kyc/{providerId}` | path `providerId`; `accountId` | object |
| Financial accounts | Create | `POST /financial-accounts` | `name`, `currency`, `reference?`, `description?`, `metadata?` | object |
| Financial accounts | Get | `GET /financial-accounts/{id}` | path `id`, `withBalance?` | object |
| Financial accounts | List | `GET /financial-accounts` | `limit`, `after`, `currency?`, `reference?` | list |
| Financial accounts | Update | `PATCH /financial-accounts/{id}` | path `id`; `name?`, `reference?`, `description?`, `metadata?` | object |
| Financial transactions | Get | `GET /financial-transactions/{id}` | path `id` | object |
| Financial transactions | List | `GET /financial-transactions` | `limit`, `after`, `financialAccountId?`, `type?`, `reference?` | list |
| Internal transfers | Create | `POST /internal-transfers` | `amount`, `sourceFinancialAccountId`, `destinationFinancialAccountId`, `description?`, `metadata?` | object |
| Internal transfers | Get | `GET /internal-transfers/{id}` | path `id` | object |
| Internal transfers | List | `GET /internal-transfers` | `limit`, `after`, `status?`, `sourceFinancialAccountId?`, `destinationFinancialAccountId?` | list |
| Internal transfers | Update | `PATCH /internal-transfers/{id}` | path `id`; `description?`, `metadata?` | object |
| Internal transfers | Delete | `DELETE /internal-transfers/{id}` | path `id` | empty |
| Payment codes | Create | `POST /payment-codes` | `mode?`, `name`, `enable?`, `amount?`, `duration?`, `customer?`, `reference?`, `authorizedProviders?`, `authorizedPhoneNumber?`, `recurrentPaymentTarget?`, `financialAccountId?`, `metadata?` | object |
| Payment codes | Get | `GET /payment-codes/{id}` | path `id` | object |
| Payment codes | List | `GET /payment-codes` | `limit`, `after`, `mode?`, `ussdCode?`, `status?` | list |
| Payment codes | Update | `PATCH /payment-codes/{id}` | path `id`; create fields that are mutable | object |
| Payment codes | Delete | `DELETE /payment-codes/{id}` | path `id` | empty |
| Payments | Get | `GET /payments/{id}` | path `id` | object |
| Payments | List | `GET /payments` | `limit`, `after`, `status?`, `createTime?` | list |
| Payments | Update | `PATCH /payments/{id}` | path `id`; `name?`, `reference?`, `metadata?` | object |
| Checkout sessions | Create | `POST /checkout-sessions` | `name`, `lineItems`, `description?`, `cancelUrl?`, `successUrl?`, `callbackState?`, `reference?`, `financialAccountId?`, `paymentOptions?`, `brandingOptions?`, `metadata?` | object |
| Checkout sessions | Get | `GET /checkout-sessions/{id}` | path `id` | object |
| Checkout sessions | List | `GET /checkout-sessions` | `limit`, `after`, `status?` | list |
| Checkout sessions | Delete | `DELETE /checkout-sessions/{id}` | path `id` | empty |
| Payouts | Create | `POST /payouts` | `amount`, `destination`, `source?`, `metadata?` | object |
| Payouts | Get | `GET /payouts/{id}` | path `id` | object |
| Payouts | List | `GET /payouts` | `limit`, `after`, `status?` | list |
| Payouts | Update | `PATCH /payouts/{id}` | path `id`; mutable destination/metadata fields | object |
| Payouts | Delete | `DELETE /payouts/{id}` | path `id` | empty |
| Receipts | Get | `GET /receipts/{orderNumber}` | path `orderNumber` | object |
| Receipts | Redeem | `POST /receipts/{orderNumber}/redeem` | path `orderNumber`; redeem all or selected entitlements | object |
| USSD OTPs | Create | `POST /ussd-otps` | `authorizedPhoneNumber`, `verificationMessage?`, `duration?`, `metadata?` | object |
| USSD OTPs | Get | `GET /ussd-otps/{id}` | path `id` | object |
| USSD OTPs | List | `GET /ussd-otps` | `limit`, `after`, `status?`, `authorizedPhoneNumber?` | list |
| USSD OTPs | Delete | `DELETE /ussd-otps/{id}` | path `id` | empty |
| Webhooks | Create | `POST /webhooks` | `name`, `url`, `events`, `verificationMethod?`, `headers?`, `alertEmails?`, `metadata?` | object |
| Webhooks | Get | `GET /webhooks/{id}` | path `id` | object |
| Webhooks | List | `GET /webhooks` | `limit`, `after` | list |
| Webhooks | Update | `PATCH /webhooks/{id}` | path `id`; mutable create fields | object |
| Webhooks | Delete | `DELETE /webhooks/{id}` | path `id` | empty |

## Resource shapes

All `Amount` values are `{currency: string, value: integer}`; `value` is an
`int64` minor-unit quantity. Metadata is an object of string values (up to 64
properties according to endpoint schemas). Fields shown with `?` are nullable
in response schemas when the endpoint OpenAPI marks them nullable; optional
request fields are omitted unless an explicit `Field[T]` is used for PATCH.

| Resource | Fields |
| --- | --- |
| Bank / MobileMoney | `providerId`, `name`, `country`, `status.active`, `featureSet.{payout,payment,kycVerification}`, `createTime`, `updateTime` |
| ProviderKYC | `account.{id,name,holderName,metadata}`, `provider.{id,type,name}` |
| FinancialAccount | `id`, `uvan`, `name`, `currency`, `reference?`, `description?`, `balance?.available`, `createTime`, `updateTime?`, `metadata?` |
| FinancialTransaction | `id`, `type`, `amount`, `timestamp`, `reference`, `financialAccount.{id,balance.after}`, `originatingReversal?`, `originatingFee?`, `ownershipGraph?`, `metadata?` |
| InternalTransfer | `id`, `status`, `amount`, `sourceFinancialAccount.id`, `destinationFinancialAccount.id`, `financialTransactionReference?`, `description?`, `failureDetail?`, `ownershipGraph?`, `createTime`, `updateTime?`, `metadata?` |
| PaymentCode | `id`, `mode`, `status`, `name?`, `amount?`, `enable`, `expireTime`, `customer?`, `ussdCode`, `reference?`, `authorizedProviders?`, `authorizedPhoneNumber?`, `recurrentPaymentTarget?`, `financialAccountId?`, `processedPaymentData?`, times, ownership graph, `metadata?` |
| Payment | `id`, `status`, `amount`, `channel.type`, `name?`, `reference?`, `orderNumber?`, `financialAccountId?`, `financialTransactionReference?`, `fees`, times, ownership graph, `metadata?` |
| CheckoutSession | `id`, `status`, `name`, `orderNumber?`, `reference?`, `description?`, `redirectUrl`, `cancelUrl`, `successUrl`, `lineItems.data`, `financialAccountId?`, `brandingOptions?`, `expireTime`, `createTime`, ownership graph, `metadata?` |
| Payout | `id`, `status`, `amount`, `source?`, `destination`, `fees`, `failureDetail?`, times, ownership graph, `metadata?` |
| Receipt | `status`, `orderName`, `orderNumber`, `orderAmount`, times, `entitlements`, `metadata?` |
| USSDOTP | `id`, `status`, `dialCode`, `authorizedPhoneNumber`, `verificationMessage?`, `createTime`, `expireTime`, `metadata?` |
| Webhook | `id`, `name`, `url`, `enabled`, `events`, `apiRelease`, `verificationMethod`, `headers?`, `alertEmails`, times, `metadata?` |

## PATCH nullability

The current endpoint pages describe PATCH as partial update but do not provide
a uniform null-clearing contract. Therefore every update input uses omission
by default and only exposes `Field[T]` for fields explicitly marked nullable
by that endpoint’s request schema. `Field[T].Null()` is reserved for those
fields; callers cannot accidentally send null for omitted-only fields.

## Sources

- [Documentation index](https://docs.monime.io/llms.txt)
- [Standard headers](https://docs.monime.io/developer-resources/api-basics/standard-headers.md)
- [Idempotency](https://docs.monime.io/developer-resources/api-basics/idempotency.md)
- [Pagination](https://docs.monime.io/developer-resources/api-basics/pagination.md)
- [Rate limits](https://docs.monime.io/developer-resources/api-basics/rate-limiting.md)
- [Versioned resource pages](https://docs.monime.io/apis/versions/caph-2025-08-23/financial-account/object.md) (replace `financial-account` and the final page name with each resource and operation in the matrix)
