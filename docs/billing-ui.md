# Billing UI contract

The tenant billing page is a read-through of the canonical `Capabilities` plan
catalog and the tenant-scoped `Subscription` resource. Usage is calculated from
the tenant-scoped projects, environments, and remote-clusters APIs; the UI does
not define alternate limits.

Only a tenant membership with the backend `billing.write` permission renders
plan-change and customer-portal actions. The backend remains authoritative and
must reject unauthorized or stale requests. Every checkout and portal transition
uses a short-lived HTTPS session returned by the server. The browser never
accepts card data and never constructs a provider URL.

Before a plan change, the UI shows target-plan limits and any current usage or
features that would be affected. Downgrade does not delete resources; read,
delete, cleanup, and export remain available. The page refreshes after focus or
visibility return so webhook-driven subscription state is not presented as
durable client state.
