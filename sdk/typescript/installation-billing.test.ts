import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import * as ib from './installation-billing.ts';
const corpus = JSON.parse(readFileSync(new URL('../../installationbilling/testdata/v1.json', import.meta.url), 'utf8'));
const raw = (s: string) => new TextEncoder().encode(s);
const hex = (b: Uint8Array) => Array.from(b, x => x.toString(16).padStart(2, '0')).join('');
const clone = <T>(v: T): T => JSON.parse(JSON.stringify(v));
test('Go/TypeScript positive frames, signatures, DTOs and sessions', () => {
    assert.equal(corpus.positive.length, 11);
    for (const p of corpus.positive) {
        const r = ib.parseRequest(raw(p.request));
        ib.verifyProof(r, 1800000000n);
        assert.equal(hex(ib.intentFrame(r.intent)), p.intentFrameHex);
        assert.equal(hex(ib.proofFrame(r)), p.proofFrameHex);
        assert.equal(ib.digest(ib.intentFrame(r.intent)), p.intentDigest);
        ib.parseChallengeRequest(raw(p.challengeRequest));
        ib.parseChallengeResponse(raw(p.challengeResponse));
        const response = ib.parseResponse(raw(p.response));
        ib.matchResponse(r, response);
        response.requestId = ib.b64(new Uint8Array(16).fill(255));
        assert.throws(() => ib.matchResponse(r, response));
        ib.validateHTTP('POST', r.intent.target, 'identity');
        if (r.sessionHandle)
            ib.matchSession(r, p.session, 1800000000n);
    }
});
test('shared negative corpus fails closed', () => {
    assert.ok(corpus.negative.length >= 40);
    for (const n of corpus.negative) {
        assert.throws(() => {
            if (n.phase === 'request')
                ib.parseRequest(raw(n.raw));
            else if (n.phase === 'response')
                ib.parseResponse(raw(n.raw));
            else if (n.phase === 'proof')
                ib.verifyProof(ib.parseRequest(raw(n.raw)), 1800000000n);
            else
                throw new Error('unknown phase');
        }, n.name);
    }
});
test('expiry/skew/session owner and rotation second signature', () => {
    for (const p of corpus.positive) {
        const r = ib.parseRequest(raw(p.request));
        assert.throws(() => ib.verifyProof(r, 1800000300n));
        assert.throws(() => ib.verifyProof(r, 1799999969n));
        ib.verifyProof(r, 1799999970n);
        if (r.sessionHandle) {
            const s = clone(p.session);
            s.scope.tenant = 'foreign';
            assert.throws(() => ib.matchSession(r, s, 1800000000n));
            assert.throws(() => ib.matchSession(r, p.session, 1800000300n));
        }
        if (p.name === 'key.rotate') {
            r.nextSignature = r.signature;
            assert.throws(() => ib.verifyProof(r, 1800000000n));
        }
    }
});
test('bounded strict parser, numeric grammar, Unicode and HTTP target', () => {
    assert.throws(() => ib.strictJSON(new Uint8Array(65537)));
    for (const s of ['{"a":1,"a":2}', '{"a":1,"\\u0061":2}', '{"a":"\\ud800"}', '{"a":1e0}', '{"a":1.0}', '{}{}'])
        assert.throws(() => ib.strictJSON(raw(s)));
    assert.throws(() => ib.strictJSON(Uint8Array.from([255])));
    for (const target of [ib.CHALLENGE_PATH + '?a=1', ib.CHALLENGE_PATH + '/', '/v1/internal/billing/checkout', '/v1/installation-billing/%63hallenge'])
        assert.throws(() => ib.validateHTTP('POST', target, ''));
    assert.throws(() => ib.validateHTTP('GET', ib.CHALLENGE_PATH, ''));
    assert.throws(() => ib.validateHTTP('POST', ib.CHALLENGE_PATH, 'gzip'));
    for (const key of [new Uint8Array(32), Uint8Array.from([1, ...Array(31).fill(0)]), new Uint8Array(32).fill(255), Uint8Array.from([2, ...Array(31).fill(0)])])
        assert.throws(() => ib.installationID(key));
    ib.parseError(raw(JSON.stringify({ version: ib.VERSION, code: 'temporarily_unavailable', requestId: '' })));
    assert.throws(() => ib.parseError(raw(JSON.stringify({ version: ib.VERSION, code: 'raw_provider_error', requestId: '' }))));
});
test('uint64 decimal strings preserve exact frame bytes beyond JS safe integer', () => {
    const c = { intentDigest: ib.b64(new Uint8Array(32)), nonce: ib.b64(new Uint8Array(32).fill(1)), issuedAt: '18446744073709551315', expiresAt: '18446744073709551615' };
    assert.equal(hex(ib.challengeFrame(c).slice(-8)), 'ffffffffffffffff');
    const r = ib.parseRequest(raw(corpus.positive[0].request));
    r.intent.currentEpoch = '9007199254740993';
    assert.ok(hex(ib.intentFrame(r.intent)).includes('0020000000000001'));
    for (const v of ['1e3', '1.0', '01', '1\n', '+1', '18446744073709551616']) {
        c.issuedAt = v;
        assert.throws(() => ib.challengeFrame(c));
    }
});
