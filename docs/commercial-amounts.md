# Signed financial display contract

`CommercialSnapshot.amounts` is optional for activation-v1 backwards compatibility.
New inclusive fixed-offer grants use `inclusive-gross-v1`: list price, net subtotal,
discount, tax, invoice total and actual paid total are distinct integer minor-unit
values sharing the snapshot currency. Paid arithmetic requires net + tax = gross
total = actual paid = fixed list price, with discount zero. Grace declares zero
new payment and keeps its paid-tier list price separately. `amountMinor` must
agree with actual paid total when the normalized breakdown is present.

Unknown presentation versions, wrong arithmetic, unsupported discount and a
grace snapshot claiming a charge fail validation. Legacy snapshots without the
field retain their serialized shape and historical tax semantics. Money does
not grant features/limits; activation binding, signature and ordinary permissions
remain the authorization boundary. The canonical OpenAPI and generated SDK digest
include the additive schema; public inputs do not accept merchant approval profiles.

`MerchantInvoiceProfile.ValidateLiveLaunch` requires a complete seller-controlled
owner/accountant-approved review record and applicable tax-policy reference. This
is a prerequisite/completeness gate, not authenticity proof or a jurisdiction/rate
decision. Missing policy blocks future live enablement. The VAT-inclusive EUR
99/299 presentation decision is already approved, separate from tax applicability.
