// Unpublished source integration only; does not alter published consumer provenance.
import {readFileSync} from 'node:fs';
import * as assert from 'node:assert/strict';
import {parseRequest,parseResponse,parseVerifiedRequest,matchResponse,intentFrame,proofFrame,matchSession,validateBody} from '../sdk/typescript/installation-billing.ts';
const c=JSON.parse(readFileSync(new URL('../installationbilling/testdata/paid-v1.json',import.meta.url),'utf8')); 
for(const p of c.positive){const r=parseVerifiedRequest(new TextEncoder().encode(p.request),BigInt(c.now));assert.equal(Buffer.from(intentFrame(r.intent)).toString('hex'),p.intentFrameHex);assert.equal(Buffer.from(proofFrame(r)).toString('hex'),p.proofFrameHex);const v=parseResponse(new TextEncoder().encode(p.response));matchResponse(r,v);if(r.sessionHandle)matchSession(r,p.session,BigInt(c.now));}
for(const n of c.negative)assert.throws(()=>n.phase==='response'?parseResponse(new TextEncoder().encode(n.raw)):parseRequest(new TextEncoder().encode(n.raw)),n.name);
for(const op of ['paid.status','redeem'] as const)for(const raw of ['null','[]','{"attemptId":"owned","licenseId":"foreign"}','{"attemptId":"a","attemptId":"b"}'])assert.throws(()=>validateBody(op,new TextEncoder().encode(raw)));
console.log(`PASS unpublished paid Go/TS vectors: ${c.positive.length} positive / ${c.negative.length} negative`);
