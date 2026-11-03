import test from 'node:test';
import assert from 'node:assert/strict';
import {MeshBufferCache} from '../src/mesh-cache.ts';

test('full mesh buffers remain transferable and bounded by bytes',()=>{
 const cache=new MeshBufferCache(10,8);
 cache.set('a',new Uint8Array([1,2,3,4,5,6]).buffer);
 const workerBuffer=cache.get('a');structuredClone(workerBuffer,{transfer:[workerBuffer]});
 assert.equal(workerBuffer.byteLength,0);assert.equal(cache.get('a').byteLength,6);
 cache.set('b',new ArrayBuffer(6));assert.equal(cache.get('a'),undefined);assert.equal(cache.byteLength,6);
 cache.set('too-big',new ArrayBuffer(11));assert.equal(cache.get('too-big'),undefined);assert.equal(cache.byteLength,6);
});

test('cache keeps recently viewed meshes and accounts for replacements',()=>{
 const cache=new MeshBufferCache(10,2);
 cache.set('a',new ArrayBuffer(3));cache.set('b',new ArrayBuffer(3));cache.get('a');
 cache.set('c',new ArrayBuffer(3));assert.equal(cache.get('b'),undefined);
 cache.set('a',new ArrayBuffer(1));assert.equal(cache.byteLength,4);
 cache.set('a',new ArrayBuffer(11));assert.equal(cache.get('a'),undefined);assert.equal(cache.byteLength,3);
});
