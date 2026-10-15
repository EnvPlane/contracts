// Public v1 contract helpers only. No admission, key storage or provider calls.
import { ed25519 } from '@noble/curves/ed25519.js';
import { sha256 } from '@noble/hashes/sha2.js';
export const VERSION = 'envplane.install-billing.v1';
export const ALGORITHM = 'Ed25519';
export const AUDIENCE = 'envplane.install-billing.test.v1';
export const LIVE_AUDIENCE = 'envplane.install-billing.live.v1';
export const MAX_ENVELOPE = 65536;
export const MAX_BODY = 32768;
export const TTL = BigInt(300);
export const MAX_EPOCH = (BigInt(1) << BigInt(63)) - BigInt(1);
export const CHALLENGE_PATH = '/v1/installation-billing/challenge';
export const ROUTES = Object.freeze({
    enroll: '/v1/installation-billing/enroll',
    'session.checkout': '/v1/installation-billing/sessions/checkout',
    'session.portal': '/v1/installation-billing/sessions/portal',
    'session.status': '/v1/installation-billing/sessions/status',
 'session.paid.status':'/v1/installation-billing/sessions/paid-status',
 'session.redeem':'/v1/installation-billing/sessions/redemption',
 'paid.status':'/v1/installation-billing/paid-status',
 'redeem':'/v1/installation-billing/redemption',
    checkout: '/v1/installation-billing/checkout', portal: '/v1/installation-billing/portal',
    status: '/v1/installation-billing/status', 'tenant.bind': '/v1/installation-billing/tenants/bind',
    'key.rotate': '/v1/installation-billing/keys/rotate', 'key.revoke': '/v1/installation-billing/keys/revoke',
    'key.status': '/v1/installation-billing/keys/status'
});
export type Operation = keyof typeof ROUTES;
export interface Scope {
    seller: string;
    merchant: string;
    mode: string;
    audience: string;
    installation: string;
    tenant: string;
}
export interface Intent {
    scope: Scope;
    operation: Operation;
    method: string;
    target: string;
    bodyDigest: string;
    currentEpoch: string;
    currentPublicKey: string;
    nextEpoch: string;
    nextPublicKey: string;
    requestId: string;
    sessionDigest: string;
}
export interface Challenge {
    intentDigest: string;
    nonce: string;
    issuedAt: string;
    expiresAt: string;
}
export interface ChallengeRequest {
    version: string;
    algorithm: string;
    intent: Intent;
    body: string;
}
export interface Request extends ChallengeRequest {
    challenge: Challenge;
    signature: string;
    nextSignature: string;
    sessionHandle: string;
}
export interface Session {
    scope: Scope;
    currentEpoch: string;
    currentPublicKey: string;
    operation: string;
    handleDigest: string;
    issuedAt: string;
    expiresAt: string;
}
export interface Record {
    attemptId: string;
    state: string;
    purchaseId: string;
    licenseId: string;
}
export interface Response {
    version: string;
    operation: Operation;
    requestId: string;
    intentDigest: string;
    state: string;
    keyEpoch: string;
    session: Session | null;
    sessionHandle: string;
    url: string;
    records: Record[];
    cursor: string;
 paidStatus?:PaidStatus;
 activationCode?:string;
}
export interface PaidRecord {attemptId:string;planId:string;planVersion:string;subscriptionState:string;paymentState:string;issuanceState:string;state:string;purchaseId:string;licenseId:string;currentLicenseId:string;expiresAt:string;graceEndsAt:string;}
export interface PaidStatus {records:PaidRecord[];cursor:string;}
export interface ErrorResponse {
    version: string;
    code: string;
    requestId: string;
}
const encoder = new TextEncoder();
const text = (s: string) => encoder.encode(s);
const fail = (): never => { throw new Error('installation billing contract rejected'); };
function ensure(ok: unknown): asserts ok { if (!ok)
    fail(); }
function object(v: unknown, names: string[]): asserts v is {
    [key: string]: unknown;
} { ensure(v !== null && typeof v === 'object' && !Array.isArray(v)); const k = Object.keys(v); ensure(k.length === names.length && names.every(n => Object.hasOwn(v, n))); }
function string(v: unknown): asserts v is string { ensure(typeof v === 'string'); }
function unicode(s: string) { for (let i = 0; i < s.length; i++) {
    const c = s.charCodeAt(i);
    if (c >= 0xd800 && c <= 0xdbff) {
        const n = s.charCodeAt(++i);
        ensure(n >= 0xdc00 && n <= 0xdfff);
    }
    else
        ensure(c < 0xdc00 || c > 0xdfff);
} }
function validText(v: unknown, max: number) { string(v); unicode(v); ensure(v.length > 0 && text(v).length <= max && !/^\p{White_Space}|\p{White_Space}$/u.test(v) && !/[\u0000-\u001f\u007f-\u009f]/u.test(v)); }
// A bounded token parser rejects duplicates before JSON.parse could erase them.
export function strictJSON(raw: Uint8Array): unknown {
    ensure(raw.length > 0 && raw.length <= MAX_ENVELOPE);
    const s = new TextDecoder('utf-8', { fatal: true }).decode(raw);
    let p = 0;
    const ws = () => { while (/[\x20\t\r\n]/.test(s[p] ?? '') && p < s.length)
        p++; };
    const str = (): string => { ensure(s[p++] === '"'); const start = p - 1; while (p < s.length) {
        const c = s[p++];
        if (c === '"') {
            const v = JSON.parse(s.slice(start, p));
            unicode(v);
            return v;
        }
        if (c === '\\')
            p++;
    } return fail(); };
    const value = (depth: number): unknown => {
        ensure(depth <= 12);
        ws();
        const c = s[p];
        if (c === '"')
            return str();
        if (c === '{') {
            p++;
            ws();
            const out: {
                [k: string]: unknown;
            } = Object.create(null);
            if (s[p] === '}') {
                p++;
                return out;
            }
            while (true) {
                ws();
                ensure(s[p] === '"');
                const key = str();
                ensure(!Object.hasOwn(out, key));
                ws();
                ensure(s[p++] === ':');
                out[key] = value(depth + 1);
                ws();
                const end = s[p++];
                if (end === '}')
                    return out;
                ensure(end === ',');
            }
        }
        if (c === '[') {
            p++;
            ws();
            const out: unknown[] = [];
            if (s[p] === ']') {
                p++;
                return out;
            }
            while (true) {
                out.push(value(depth + 1));
                ws();
                const end = s[p++];
                if (end === ']')
                    return out;
                ensure(end === ',');
            }
        }
        const m = /^(?:true|false|null|-?(?:0|[1-9][0-9]*))/.exec(s.slice(p));
        ensure(m);
        p += m[0].length;
        return JSON.parse(m[0]);
    };
    const out = value(0);
    ws();
    ensure(p === s.length);
    return out;
}
export function b64(b: Uint8Array): string { let s = ''; for (const c of b)
    s += String.fromCharCode(c); return btoa(s).replace(/=/g, '').replace(/\+/g, '-').replace(/\//g, '_'); }
export function decode(v: unknown, n: number, nonzero = false): Uint8Array { string(v); ensure(/^[A-Za-z0-9_-]*$/.test(v)); let s: string; try {
    s = atob(v.replace(/-/g, '+').replace(/_/g, '/'));
}
catch {
    return fail();
} const b = Uint8Array.from(s, c => c.charCodeAt(0)); ensure(b.length === n && b64(b) === v && (!nonzero || b.some(x => x !== 0))); return b; }
function decimal(v: unknown, zero = false): bigint { string(v); ensure(v.length <= 20 && /^(0|[1-9][0-9]*)$/.test(v)); const n = BigInt(v); ensure(n.toString() === v); ensure(n <= (BigInt(1) << BigInt(64)) - BigInt(1) && (zero || n > BigInt(0))); return n; }
function epoch(v: unknown) { const n = decimal(v); ensure(n <= MAX_EPOCH); return n; }
function same(a: Uint8Array, b: Uint8Array) { return a.length === b.length && a.every((x, i) => x === b[i]); }
function point(b: Uint8Array) { try {
    const p = ed25519.Point.fromBytes(b, false);
    return same(p.toBytes(), b) && !p.isSmallOrder() && p.isTorsionFree();
}
catch {
    return false;
} }
function window(a: string, b: string) { const x = decimal(a), y = decimal(b); ensure(x <= (BigInt(1) << BigInt(64)) - BigInt(301) && y === x + TTL); }
function windowNow(a: string, b: string, now: bigint) { window(a, b); ensure(now >= BigInt(0) && now < decimal(b) && decimal(a) <= now + BigInt(30)); }
function uint(v: string) { const b = new Uint8Array(8); new DataView(b.buffer).setBigUint64(0, decimal(v, true)); return b; }
export function frame(domain: string, fields: Uint8Array[]) { const d = text(domain), magic = text('ENVPLANE-INSTALL-BILL\0\x01'); const out = new Uint8Array(magic.length + 2 + d.length + 2 + fields.reduce((n, f) => n + 4 + f.length, 0)); const v = new DataView(out.buffer); let p = 0; out.set(magic, p); p += magic.length; v.setUint16(p, d.length); p += 2; out.set(d, p); p += d.length; v.setUint16(p, fields.length); p += 2; for (const f of fields) {
    v.setUint32(p, f.length);
    p += 4;
    out.set(f, p);
    p += f.length;
} return out; }
export const digest = (b: Uint8Array) => b64(sha256(b));
export function installationID(pub: Uint8Array) { ensure(point(pub)); return 'epi1_ed25519_' + Array.from(sha256(frame('genesis', [pub])), c => c.toString(16).padStart(2, '0')).join(''); }
function scope(v: unknown): asserts v is Scope { object(v, ['seller', 'merchant', 'mode', 'audience', 'installation', 'tenant']); ensure(v.seller === 'issuer-default' && ((v.mode === 'test' && v.audience === AUDIENCE) || (v.mode === 'live' && v.audience === LIVE_AUDIENCE))); validText(v.merchant, 255); string(v.merchant); string(v.installation); ensure(v.installation.length === 77); ensure(/^acct_[A-Za-z0-9_]{1,250}$/.test(v.merchant) && /^epi1_ed25519_[0-9a-f]{64}$/.test(v.installation)); validText(v.tenant, 128); }
function scopeFields(s: Scope) { return [s.seller, s.merchant, s.mode, s.audience, s.installation, s.tenant].map(text); }
export function validateIntent(v: unknown): asserts v is Intent {
    object(v, ['scope', 'operation', 'method', 'target', 'bodyDigest', 'currentEpoch', 'currentPublicKey', 'nextEpoch', 'nextPublicKey', 'requestId', 'sessionDigest']);
    scope(v.scope);
    string(v.operation);
    ensure(Object.hasOwn(ROUTES, v.operation));
    ensure(v.method === 'POST' && v.target === ROUTES[v.operation as Operation]);
    const cur = epoch(v.currentEpoch), pub = decode(v.currentPublicKey, 32, true);
    ensure(point(pub));
    decode(v.bodyDigest, 32);
    decode(v.requestId, 16, true);
    const sd = decode(v.sessionDigest, 32), next = decimal(v.nextEpoch, true), np = decode(v.nextPublicKey, 32);
    if (v.operation === 'key.rotate')
        ensure(cur < MAX_EPOCH && next === cur + BigInt(1) && point(np) && !same(pub, np));
    else
        ensure(next === BigInt(0) && np.every(x => x === 0));
    const action = ['checkout', 'portal', 'status', 'paid.status', 'redeem'].includes(v.operation);
    ensure(action !== sd.every(x => x === 0));
    if (v.operation === 'enroll')
        ensure(cur === BigInt(1) && installationID(pub) === v.scope.installation);
}
export function intentFrame(i: Intent) { validateIntent(i); return frame('intent', [...scopeFields(i.scope), text(i.operation), text(i.method), text(i.target), decode(i.bodyDigest, 32), uint(i.currentEpoch), decode(i.currentPublicKey, 32), uint(i.nextEpoch), decode(i.nextPublicKey, 32), decode(i.requestId, 16), decode(i.sessionDigest, 32)]); }
export function challengeFrame(c: Challenge) { object(c, ['intentDigest', 'nonce', 'issuedAt', 'expiresAt']); decode(c.intentDigest, 32); decode(c.nonce, 32, true); string(c.issuedAt); string(c.expiresAt); window(c.issuedAt, c.expiresAt); return frame('challenge', [decode(c.intentDigest, 32), decode(c.nonce, 32), uint(c.issuedAt), uint(c.expiresAt)]); }
function proofDomain(op: string) { return op === 'enroll' ? 'enroll-proof' : op.startsWith('session.') ? 'session-proof' : op === 'key.rotate' ? 'rotate-proof' : op === 'key.revoke' ? 'revoke-proof' : 'action-proof'; }
export function proofFrame(r: Request) { const id = digest(intentFrame(r.intent)); ensure(id === r.challenge.intentDigest); return frame(proofDomain(r.intent.operation), [decode(id, 32), decode(digest(challengeFrame(r.challenge)), 32)]); }
function body(s: unknown) { string(s); ensure(s.length <= 43691); const b = decode(s, Math.floor(s.length * 6 / 8)); ensure(b.length <= MAX_BODY); return b; }
export function validateBody(op: Operation, raw: Uint8Array) {
    ensure(raw.length <= MAX_BODY);
    const b = strictJSON(raw);
    switch (op) {
        case 'paid.status':
            object(b,['attemptId','cursor','limit']);string(b.attemptId);if(b.attemptId)validText(b.attemptId,128);string(b.cursor);if(b.cursor)validText(b.cursor,128);ensure(typeof b.limit==='number'&&Number.isInteger(b.limit)&&b.limit>=1&&b.limit<=25);break;
        case 'redeem':
            object(b,['attemptId','nonce']);validText(b.attemptId,128);decode(b.nonce,24,true);break;
        case 'checkout':
            object(b, ['attemptId', 'planId', 'planVersion', 'returnRoute']);
            for (const k of ['attemptId', 'planId', 'planVersion'])
                validText(b[k], 128);
            ensure(b.returnRoute === 'billing');
            break;
        case 'portal':
            object(b, ['returnRoute']);
            ensure(b.returnRoute === 'billing');
            break;
        case 'status':
            object(b, ['attemptId', 'cursor', 'limit']);
            validText(b.attemptId, 128);
            string(b.cursor);
            ensure(text(b.cursor).length <= 1024);
            if (b.cursor)
                validText(b.cursor, 1024);
            ensure(typeof b.limit === 'number' && Number.isInteger(b.limit) && b.limit >= 1 && b.limit <= 25);
            break;
        case 'key.status':
            object(b, ['transitionRequestId']);
            decode(b.transitionRequestId, 16, true);
            break;
        default:
            ensure(Object.hasOwn(ROUTES, op));
            object(b, []);
    }
}
function common(v: ChallengeRequest) { ensure(v.version === VERSION && v.algorithm === ALGORITHM); validateIntent(v.intent); const raw = body(v.body); ensure(digest(raw) === v.intent.bodyDigest); validateBody(v.intent.operation, raw); }
export function parseChallengeRequest(raw: Uint8Array): ChallengeRequest { const v = strictJSON(raw); object(v, ['version', 'algorithm', 'intent', 'body']); const r = v as unknown as ChallengeRequest; common(r); return r; }
export function parseRequest(raw: Uint8Array): Request { const v = strictJSON(raw); object(v, ['version', 'algorithm', 'intent', 'body', 'challenge', 'signature', 'nextSignature', 'sessionHandle']); const r = v as unknown as Request; common(r); proofFrame(r); decode(r.signature, 64); if (r.intent.operation === 'key.rotate')
    decode(r.nextSignature, 64);
else
    ensure(r.nextSignature === ''); if (['checkout', 'portal', 'status', 'paid.status', 'redeem'].includes(r.intent.operation))
    decode(r.sessionHandle, 32, true);
else
    ensure(r.sessionHandle === ''); return r; }
export function verifyProof(r: Request, now: bigint) { parseRequest(text(JSON.stringify(r))); windowNow(r.challenge.issuedAt, r.challenge.expiresAt, now); const msg = proofFrame(r); const check = (key: string, signature: string) => { const k = decode(key, 32), s = decode(signature, 64); ensure(point(k) && point(s.slice(0, 32))); ensure(ed25519.verify(s, msg, k, { zip215: false })); }; check(r.intent.currentPublicKey, r.signature); if (r.intent.operation === 'key.rotate')
    check(r.intent.nextPublicKey, r.nextSignature); }
export function sessionFrame(s: Session) { object(s, ['scope', 'currentEpoch', 'currentPublicKey', 'operation', 'handleDigest', 'issuedAt', 'expiresAt']); scope(s.scope); epoch(s.currentEpoch); const pub = decode(s.currentPublicKey, 32, true); ensure(point(pub) && ['checkout', 'portal', 'status', 'paid.status', 'redeem'].includes(s.operation)); decode(s.handleDigest, 32, true); window(s.issuedAt, s.expiresAt); return frame('session', [...scopeFields(s.scope), uint(s.currentEpoch), pub, text(s.operation), decode(s.handleDigest, 32), uint(s.issuedAt), uint(s.expiresAt)]); }
export function matchSession(r: Request, s: Session, now: bigint) { const f = sessionFrame(s); windowNow(s.issuedAt, s.expiresAt, now); ensure(s.operation === r.intent.operation && s.currentEpoch === r.intent.currentEpoch && s.currentPublicKey === r.intent.currentPublicKey && scopeFields(s.scope).every((b, i) => same(b, scopeFields(r.intent.scope)[i])) && digest(f) === r.intent.sessionDigest && digest(decode(r.sessionHandle, 32, true)) === s.handleDigest); }
export function validateHTTP(method: string, target: string, encoding: string) { ensure(method === 'POST' && (encoding === '' || encoding === 'identity') && (target === CHALLENGE_PATH || Object.values(ROUTES).some(p => p === target))); }
export function parseChallengeResponse(raw: Uint8Array) { const r = strictJSON(raw); object(r, ['version', 'algorithm', 'challenge']); ensure(r.version === VERSION && r.algorithm === ALGORITHM); challengeFrame(r.challenge as Challenge); return r as unknown as {
    version: string;
    algorithm: string;
    challenge: Challenge;
}; }
export const ERROR_STATUS = Object.freeze({ invalid_request: 400, unsupported_version: 400, unsupported_algorithm: 400, invalid_proof: 400, scope_rejected: 403, authority_revoked: 403, key_epoch_stale: 403, session_rejected: 403, challenge_expired: 410, session_expired: 410, request_conflict: 409, challenge_consumed: 409, rate_limited: 429, temporarily_unavailable: 503, not_found: 404 });
export function parseError(raw: Uint8Array): ErrorResponse { const r = strictJSON(raw); object(r, ['version', 'code', 'requestId']); string(r.code); ensure(r.version === VERSION && Object.hasOwn(ERROR_STATUS, r.code)); if (r.requestId !== '')
    decode(r.requestId, 16, true); return r as unknown as ErrorResponse; }
export function parseResponse(raw: Uint8Array): Response {
    const v = strictJSON(raw);
    const extensionKeys=['paidStatus','activationCode'].filter(k=>Object.hasOwn(v as object,k));
    object(v, ['version', 'operation', 'requestId', 'intentDigest', 'state', 'keyEpoch', 'session', 'sessionHandle', 'url', 'records', 'cursor',...extensionKeys]);
    const r = v as unknown as Response;
    ensure(r.version === VERSION && Object.hasOwn(ROUTES, r.operation));
    decode(r.requestId, 16, true);
    decode(r.intentDigest, 32);
    epoch(r.keyEpoch);
    const states: {
        [k: string]: string;
    } = { enroll: 'registered', 'tenant.bind': 'bound', 'session.checkout': 'ready', 'session.portal': 'ready', 'session.status': 'ready', 'session.paid.status':'ready', 'session.redeem':'ready', 'paid.status':'observed', 'redeem':'redeemed', portal: 'ready', status: 'observed', 'key.rotate': 'rotated', 'key.revoke': 'revoked', 'key.status': 'observed' };
    if (r.operation === 'checkout')
        ensure(['accepted', 'pending', 'ready', 'review_required', 'failed'].includes(r.state));
    else
        ensure(r.state === states[r.operation]);
    if (r.operation.startsWith('session.')) {
        ensure(r.session);
        sessionFrame(r.session);
        ensure(r.session.operation === r.operation.slice(8) && r.session.currentEpoch === r.keyEpoch && digest(decode(r.sessionHandle, 32, true)) === r.session.handleDigest);
    }
    else
        ensure(r.session === null && r.sessionHandle === '');
    const needsURL = r.operation === 'portal' || (r.operation === 'checkout' && r.state === 'ready');
    if (needsURL) {
        string(r.url);
        ensure(text(r.url).length <= 4096 && !/[\r\n\\]/.test(r.url));
        const u = new URL(r.url);
        ensure(r.url.startsWith('https://' + (r.operation === 'portal' ? 'billing.stripe.com/' : 'checkout.stripe.com/')) && !r.url.includes('#') && u.protocol === 'https:' && u.host === (r.operation === 'portal' ? 'billing.stripe.com' : 'checkout.stripe.com') && !u.username && !u.password && !u.hash && u.pathname !== '');
    }
    else
        ensure(r.url === '');
    ensure(Array.isArray(r.records) && r.records.length <= 25);
    string(r.cursor);
    ensure(text(r.cursor).length <= 1024);
    if (r.cursor)
        validText(r.cursor, 1024);
    if (r.operation !== 'status')
        ensure(r.records.length === 0 && r.cursor === '');
    const seen = new Set<string>();
    for (const x of r.records) {
        object(x, ['attemptId', 'state', 'purchaseId', 'licenseId']);
        validText(x.attemptId, 128);
        ensure(!seen.has(x.attemptId));
        seen.add(x.attemptId);
        for (const k of ['purchaseId', 'licenseId']) {
            string(x[k]);
            if (x[k])
                validText(x[k], 128);
        }
        ensure(['checkout_pending', 'payment_processing', 'paid/issuance_pending', 'activation_required', 'past_due', 'canceled', 'failed'].includes(x.state as string));
    }
    validatePaidResponse(r);
    return r;
}
export function parseVerifiedRequest(raw: Uint8Array, now: bigint): Request { const r = parseRequest(raw); verifyProof(r, now); return r; }
export function matchResponse(r: Request, v: Response) { ensure(v.operation === r.intent.operation && v.requestId === r.intent.requestId && v.intentDigest === digest(intentFrame(r.intent)) && v.keyEpoch === (v.operation === 'key.rotate' ? r.intent.nextEpoch : r.intent.currentEpoch)); if (v.session)
    ensure(scopeFields(v.session.scope).every((b, i) => same(b, scopeFields(r.intent.scope)[i])) && v.session.currentPublicKey === r.intent.currentPublicKey); }

export function validatePaidStatus(v:unknown): asserts v is PaidStatus {
 object(v,['records','cursor']);ensure(Array.isArray(v.records)&&v.records.length<=25);string(v.cursor);if(v.cursor)validText(v.cursor,128);const seen=new Set<string>();
 for(const item of v.records){
  object(item,['attemptId','planId','planVersion','subscriptionState','paymentState','issuanceState','state','purchaseId','licenseId','currentLicenseId','expiresAt','graceEndsAt']);
  for(const k of ['attemptId','planId','planVersion','subscriptionState'])validText(item[k],128);
  const r=item as unknown as PaidRecord;ensure(!seen.has(r.attemptId));seen.add(r.attemptId);
  for(const k of ['purchaseId','licenseId','currentLicenseId']){string(item[k]);if(item[k])validText(item[k],128);}
  for(const k of ['expiresAt','graceEndsAt'])ensure(decimal(item[k],true)<=BigInt('253402300799'));
  ensure(['awaiting_payment','payment_failed','paid','not_charged'].includes(r.paymentState)&&['not_started','issuance_pending','issued'].includes(r.issuanceState)&&['checkout_pending','payment_processing','paid/issuance_pending','activation_required','past_due','canceled','failed'].includes(r.state));
  if(r.state==='activation_required')ensure(r.paymentState==='paid'&&r.issuanceState==='issued'&&r.licenseId!==''&&r.currentLicenseId===r.licenseId&&r.purchaseId!==''&&r.expiresAt!=='0');
  if(r.state==='paid/issuance_pending')ensure(r.paymentState==='paid'&&r.issuanceState==='issuance_pending'&&r.purchaseId!==''&&r.licenseId==='');
 }
}
function validatePaidResponse(r:Response){
 if(r.operation==='paid.status'||r.operation==='redeem'){validatePaidStatus(r.paidStatus)}else ensure(r.paidStatus===undefined);
 if(r.operation==='redeem'){ensure(r.paidStatus?.records.length===1&&r.paidStatus.cursor===''&&r.paidStatus.records[0].state==='activation_required');string(r.activationCode);ensure(text(r.activationCode).length<=8192&&r.activationCode.startsWith('v1.')&&!/[ \t\r\n]/.test(r.activationCode))}else ensure(r.activationCode===undefined);
}
