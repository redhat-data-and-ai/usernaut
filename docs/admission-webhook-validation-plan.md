# Admission webhook validation plan (usernaut)

This document exports the plan for adding validating admission webhook support to **usernaut**, aligned with **data-platform-operator** (Ishita’s work) and the first rule: **when any backend type is `rover`, require `ServiceAccount` `usernaut-prod` in namespace `usernaut`.**

Reference commits in data-platform-operator (for diff review):

- `4e09a78d80b56dfd9130968f2223eb5adfce91ed`
- `bab8da3de39e0272d139c0aff6d47bbe9f5c1fb7`
- `5bcc04e437d1bd0ed6e581b90ff780108ed494b5`
- `3f93f853f59e386f58a8d791f2799ff7bb2051c6`
- `37923477d28cdf6b20ca7fb8f96d9a05ff139c69`

---

## Context

### What the DPO pattern does

1. **Runtime:** `controller-runtime` `webhook.Server` with TLS certs on disk (`CertDir`, `CertName`, `KeyName`).
2. **Handler:** A **unified** validating webhook on a fixed HTTP path; uses `admission.WithCustomValidator` with `runtime.Unknown` and a **decoder** when `ValidatingWebhookConfiguration` matches broad `resources` (e.g. `*`) under the API group.
3. **Delegation:** After decoding to a concrete type, optional per-type `ValidateCreate` / `ValidateUpdate` / `ValidateDelete`.
4. **Cluster:** `ValidatingWebhookConfiguration` → `Service` → operator pod (e.g. `:9443`), **cert-manager** (or equivalent) for serving certs and **CA injection** on the webhook config.

### Current usernaut state

- `cmd/main.go` creates `webhook.Server` but does **not** register handlers or mount webhook certs (default scaffold).
- Single CRD: **`Group`** (`groups.operator.dataverse.redhat.com`). Backends: `spec.backends[]` with `name` and `type` (`api/v1alpha1/group_types.go`).
- Rover type in code/config: **`rover`** (`pkg/clients/client.go`, appconfig samples).
- **`WATCHED_NAMESPACE`** defaults to **`usernaut`** when unset.

### First validation rule

- If **any** `spec.backends[]` has `type` **rover** (recommend `strings.EqualFold` vs `"rover"` unless you require exact casing), then **`ServiceAccount` `usernaut-prod`** must exist in namespace **`usernaut`** (or match `WATCHED_NAMESPACE` if you want env-accurate behavior).

### Topology warning

Usernaut’s `Group` uses API group **`operator.dataverse.redhat.com`**, same as data-platform-operator. If **both** operators register a validating webhook for the **same** `groups` GVR in one cluster, admission can **conflict** or run twice. Confirm only usernaut owns `groups` admission in target clusters.

---

## Design options

| Approach | Pros | Cons |
|----------|------|------|
| **A. Group-only webhook** (`resources: groups`) | No `runtime.Unknown` decoding; smaller code. | More CRDs later → add rules or merge into unified handler. |
| **B. Unified webhook (like DPO)** | Same extensibility as Ishita’s pattern. | More boilerplate for one CRD today. |

**Recommendation:** **A** for first use case; **B** if you want parity with DPO and many future rules.

---

## Step-by-step implementation

### 1. Scope and HTTP path

- **Group-only:** VWC rules: `resources: ["groups"]`. Path e.g. `/validate-operator-dataverse-redhat-com-v1alpha1-group` (must match `clientConfig.service.path`).
- **Unified:** Mirror DPO (`resources: ["*"]` or equivalent), shared path, **`decodeUnknownObject`** pattern from `api/v1alpha1/unified_webhook.go`.

Use **usernaut-specific** VWC name (e.g. `usernaut-validating-webhook-configuration`) to avoid clashing with DPO.

### 2. Validation logic (Go)

- Add handler (e.g. `api/v1alpha1/group_webhook.go` or `internal/webhook/`).
- Inject **`client.Client`** from the manager.
- **Create:** If new `Group` spec includes rover backend → `Get` `ServiceAccount` `usernaut-prod` in `usernaut`; if not found → admission error with clear message.
- **Update:** Run check on **new** spec when rover still present; if rover removed, skip SA check.
- **Delete:** Typically allow (unless product requires otherwise).

Wire with `builder.WebhookManagedBy(mgr).For(&Group{}).WithValidator(...).Complete()` or `mgr.GetWebhookServer().Register(...)` like DPO.

### 3. `cmd/main.go`

- Optional **`--disable-webhook`** or env (e.g. `ENABLE_WEBHOOKS`) for dev without certs.
- When enabled: set **`CertDir` / `CertName` / `KeyName`** to match volume mount (e.g. `/tmp/k8s-webhook-server/serving-certs`).
- After `NewManager`, call setup e.g. `SetupGroupWebhookWithManager(mgr)`.

### 4. Kustomize / manifests

Enable **[WEBHOOK]** and **[CERTMANAGER]** blocks in `config/default/base/kustomization.yaml` (stubs exist, commented).

1. **`config/webhook/`:** `Service` for webhook; `ValidatingWebhookConfiguration` (`failurePolicy: Fail` or `Ignore` if you accept bypass when down), `sideEffects: None`, rules for `operator.dataverse.redhat.com` / `v1alpha1` / `groups` (or `*` if unified), ops `CREATE`, `UPDATE` (+ `DELETE` only if implemented).
2. **`manager_webhook_patch.yaml`:** container port **9443**, volume mount certs, secret volume.
3. **`config/certmanager/`:** `Certificate` with DNS for `webhook-service.<ns>.svc` and `.cluster.local`; **`webhookcainjection_patch.yaml`** with `cert-manager.io/inject-ca-from`.
4. Overlays: consistent namespace for Service, Certificate, VWC `clientConfig.service.namespace`.
5. **NetworkPolicy:** If used, allow apiserver → webhook path (see DPO `networkpolicy-webhook-ingress.yaml`).

### 5. RBAC

- Grant **`get`** (and optionally **`list`**) on **`serviceaccounts`** in namespace **`usernaut`** (or wherever the SA must exist) to the manager’s `ServiceAccount`.

### 6. Testing

- **envtest:** webhooks + CA/VWC if supported in project.
- **Kind / e2e:** Optional, like DPO (`SKIP_VALIDATING_WEBHOOK`, cert generation, CA patch).
- **Manual:** Group with rover backend without SA → denied; create `usernaut-prod` → allowed.

### 7. Map DPO commits to work items

1. API + webhook handler (unified or Group-only).  
2. `main.go` webhook options + registration + disable flag.  
3. Kustomize: service, VWC, manager patch, cert-manager, CA injection.  
4. E2E / hack scripts (if needed).  
5. RBAC for webhook dependencies (here: `ServiceAccount`).

---

## Summary

Add three layers: **(1)** validating handler + manager client, **(2)** TLS + `ValidatingWebhookConfiguration` + `Service`, **(3)** RBAC to read `ServiceAccount` `usernaut-prod` in **`usernaut`**, with rover detection on **`spec.backends[].type`**. Prefer **Group-only** webhook unless you want a **unified** layout; confirm no duplicate webhook on the same `Group` GVR from another operator.
