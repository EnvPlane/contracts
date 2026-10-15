import { readFileSync } from 'node:fs';
import * as assert from 'node:assert/strict';
import { ed25519 } from '@noble/curves/ed25519.js';
import { sha256 } from '@noble/hashes/sha2.js';
import { AUDIENCE, LIVE_AUDIENCE, b64, intentFrame, proofFrame, sessionFrame, matchSession, parseVerifiedRequest, validateIntent } from '../sdk/typescript/installation-billing.ts';
const c = JSON.parse(readFileSync(new URL('../installationbilling/testdata/v1.json', import.meta.url),'utf8'));
const now=BigInt(c.now), encode=(r:unknown)=>new TextEncoder().encode(JSON.stringify(r));
const seed=Uint8Array.from({length:32},(_,i)=>i);
for(const p of c.positive){
 const original=parseVerifiedRequest(new TextEncoder().encode(p.request),now);
 const live=structuredClone(original);
 live.intent.scope.mode='live'; live.intent.scope.audience=LIVE_AUDIENCE;
 assert.throws(()=>parseVerifiedRequest(encode(live),now));
 const session=p.session && live.sessionHandle!==''?structuredClone(p.session):undefined;
 if(session){session.scope=structuredClone(live.intent.scope);live.intent.sessionDigest=b64(sha256(sessionFrame(session)));}
 const next=Uint8Array.from({length:32},()=>42);
 if(live.intent.operation==='key.rotate') live.intent.nextPublicKey=b64(ed25519.getPublicKey(next));
 live.challenge.intentDigest=b64(sha256(intentFrame(live.intent)));
 live.signature=b64(ed25519.sign(proofFrame(live),seed));
 if(live.intent.operation==='key.rotate') live.nextSignature=b64(ed25519.sign(proofFrame(live),next));
 parseVerifiedRequest(encode(live),now);
 if(session){matchSession(live,session,now);assert.throws(()=>matchSession(live,p.session,now));assert.throws(()=>matchSession(original,session,now));}
 live.intent.scope.mode='test'; live.intent.scope.audience=AUDIENCE;
 assert.throws(()=>parseVerifiedRequest(encode(live),now));
 for(const [mode,audience] of [['test',LIVE_AUDIENCE],['live',AUDIENCE]]) {
  const foreign=structuredClone(original.intent); foreign.scope.mode=mode; foreign.scope.audience=audience;
  assert.throws(()=>validateIntent(foreign));
 }
}
console.log('PASS: historical corpus, mode/audience and signature replay isolation');
