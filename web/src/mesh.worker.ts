import {STLLoader} from 'three/addons/loaders/STLLoader.js';
self.onmessage=({data}:{data:ArrayBuffer})=>{
 try{const geometry=new STLLoader().parse(data);geometry.computeVertexNormals();const positions=geometry.getAttribute('position').array as Float32Array,normals=geometry.getAttribute('normal').array as Float32Array;
 if(!positions.length)throw Error('This model has no visible surfaces. Choose another file.');
 self.postMessage({positions,normals},{transfer:[positions.buffer,normals.buffer]});geometry.dispose();
 }catch(e:any){self.postMessage({error:e.message||'The model preview could not be drawn.'})}
};
