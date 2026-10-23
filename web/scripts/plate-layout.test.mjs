import test from 'node:test';
import assert from 'node:assert/strict';
import {nativePlateOrigin,displayPlateOrigin} from '../src/plate-layout.ts';

test('native multi-plate meshes retain their position within the displayed bed',()=>{
 // Saved Bambu A1 mini world positions use a 216mm stride. Plate two's
 // model at (300,90) must appear at (84,90) inside that plate, not off the bed.
 assert.deepEqual(nativePlateOrigin(1,4,180),[216,0,0]);
 assert.deepEqual(nativePlateOrigin(2,4,180),[0,-216,0]);
 const native=nativePlateOrigin(1,4,180),shown=displayPlateOrigin(1,4,180);
 assert.deepEqual([300-native[0]+shown[0],90-native[1]+shown[1]],[302,90]);
 // Nine imported plates require a third column; none may overlap another bed.
 const positions=Array.from({length:9},(_,n)=>displayPlateOrigin(n,9,256));
 assert.equal(new Set(positions.map(p=>p.join(','))).size,9);
 assert.deepEqual(positions[8],[588,588,0]);
});
