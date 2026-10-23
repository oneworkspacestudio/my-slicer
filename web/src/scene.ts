import React,{useEffect,useRef,useState} from 'react';
import * as T from 'three';
import {OrbitControls} from 'three/addons/controls/OrbitControls.js';
import {nativePlateOrigin,displayPlateOrigin} from './plate-layout';
import {SceneStatus} from './generated/SceneStatus';

const colors=['#00ae42','#78c77c','#188038','#f9ab00','#a142f4','#e8710a','#12b5cb','#d93025'];
export function Scene({project,models,layer,feature,request,onSelect,uploading,overview=false,pathPercent=100,viewMode,viewToken}:any){
 const host=useRef<HTMLDivElement>(null),[error,setError]=useState(''),[loading,setLoading]=useState(''),[attempt,setAttempt]=useState(0);
 const cache=useRef(new Map<string,ArrayBuffer>());
 const modelKey=(project.objects||[]).map((o:any)=>models.find((m:any)=>m.id===o.modelID)?.id||'pending').join(',');
 useEffect(()=>{const el=host.current!;let alive=true;let renderer:T.WebGLRenderer;
  try{renderer=new T.WebGLRenderer({antialias:true,alpha:false});}catch{setError('WebGL is unavailable. Enable hardware acceleration in your browser to inspect the plate. Native slicing and archive downloads remain available.');return}
  setError('');renderer.localClippingEnabled=true;renderer.setPixelRatio(Math.min(devicePixelRatio,2));el.append(renderer.domElement);renderer.domElement.setAttribute('aria-label','Interactive build plate; drag to orbit, scroll to zoom, click an object to select');
  const scene=new T.Scene();scene.background=new T.Color('#e6e6e6');const w=project.settings.model==='A1 mini'?180:256;
  const camera=new T.PerspectiveCamera(40,1,0.1,3000);camera.up.set(0,0,1);camera.position.set(w*1.6,-w*1.3,w*1.3);
  const controls=new OrbitControls(camera,renderer.domElement);controls.target.set(w/2,w/2,12);controls.enableDamping=true;
  scene.add(new T.HemisphereLight(0xffffff,0x526178,2));const light=new T.DirectionalLight(0xffffff,3);light.position.set(100,-200,400);scene.add(light);
  const nativeModel=models.find((m:any)=>m.id===project.objects?.[0]?.modelID&&m.kind==='3mf');
  const plates=overview&&nativeModel?.import?.plates?.length?nativeModel.import.plates:[{index:project.settings.nativePlate||1}];
  const offset=(index:number)=>new T.Vector3(...displayPlateOrigin(index,plates.length,w));
  const sourceWidth=/(A1.?mini|A1M)/i.test(nativeModel?.import?.printer||'')?180:256;
  const columns=Math.ceil(Math.sqrt(plates.length)),rows=Math.ceil(plates.length/columns);
  for(let index=0;index<plates.length;index++){
  const shift=offset(index);
  const plate=new T.Mesh(new T.PlaneGeometry(w,w),new T.MeshStandardMaterial({color:0x414745,side:T.DoubleSide,roughness:1}));plate.position.set(w/2,w/2,-.1).add(shift);scene.add(plate);
  const grid=new T.GridHelper(w,w/10,0xb1b5b2,0x7f8581);grid.rotation.x=Math.PI/2;grid.position.set(w/2,w/2,0).add(shift);scene.add(grid);
  const border=new T.LineLoop(new T.BufferGeometry().setFromPoints([new T.Vector3(0,0,.1),new T.Vector3(w,0,.1),new T.Vector3(w,w,.1),new T.Vector3(0,w,.1)]),new T.LineBasicMaterial({color:0x00ae42}));border.position.copy(shift);scene.add(border);
  }
  if(viewMode==='top'){camera.position.set(w/2,w/2-.001,w*2);controls.update()}
  scene.add(new T.AxesHelper(35));
  const selectables:T.Object3D[]=[];const workers:Worker[]=[];let pending=0;setLoading('');
 const fit=()=>{const box=new T.Box3();for(const o of selectables)box.expandByObject(o);if(box.isEmpty())return;box.expandByPoint(new T.Vector3(0,0,0));box.expandByPoint(new T.Vector3(columns*w+(columns-1)*38,rows*w+(rows-1)*38,0));const center=box.getCenter(new T.Vector3()),size=box.getSize(new T.Vector3());controls.target.copy(center);const distance=Math.max(70,size.length()*1.35);camera.position.copy(center).add(new T.Vector3(1.1,-1,.85).normalize().multiplyScalar(distance));if(viewMode==='top'){camera.position.copy(center).add(new T.Vector3(0,-.001,distance));}controls.update()};
  if(layer){const groups=new Map<number,number[]>();for(const s of (layer.segments||[]).slice(0,Math.ceil((layer.segments?.length||0)*pathPercent/100))){if(feature!=='all'&&Number(feature)!==s[6])continue;const arr=groups.get(s[6])||[];arr.push(...s.slice(0,6));groups.set(s[6],arr)}for(const[k,arr]of groups){const g=new T.BufferGeometry();g.setAttribute('position',new T.Float32BufferAttribute(arr,3));scene.add(new T.LineSegments(g,new T.LineBasicMaterial({color:colors[k%colors.length]})))}}
  for(const o of (project.objects||[]).flatMap((o:any)=>plates.map((plate:any,index:number)=>({...o,previewPlate:plate.index,shift:offset(index).sub(new T.Vector3(...nativePlateOrigin(Math.max(0,(nativeModel?.import?.plates||[]).findIndex((p:any)=>p.index===plate.index)),nativeModel?.import?.plates?.length||1,sourceWidth)))})))){const m=models.find((m:any)=>m.id===o.modelID);if(!m){setLoading('Preparing your model…');continue}
   pending++;setLoading('Loading '+m.name+'…');
   const url='/api/models/'+m.id+'/mesh'+(m.kind==='3mf'?'?plate='+(o.previewPlate||project.settings.nativePlate||1):'');
   const getBuffer=async()=>{const saved=cache.current.get(url);if(saved)return saved.slice(0);const blob:Blob=await request(url);const buffer=await blob.arrayBuffer();if(alive){if(cache.current.size>=8)cache.current.delete(cache.current.keys().next().value!);cache.current.set(url,buffer.slice(0))}return buffer};
   // A wire outline gives immediate spatial feedback while the real mesh loads.
   const lo=m.min,hi=m.max;let outline:T.Mesh|undefined;
   const place=(mesh:T.Object3D)=>{const t=o.transform;mesh.scale.setScalar(t.scale);mesh.rotation.order='ZYX';mesh.rotation.set(t.rx*Math.PI/180,t.ry*Math.PI/180,t.rz*Math.PI/180);mesh.position.set(t.x,t.y,t.z).add(o.shift||new T.Vector3())};
   if(m.kind==='stl'&&hi&&lo){outline=new T.Mesh(new T.BoxGeometry(hi[0]-lo[0],hi[1]-lo[1],hi[2]-lo[2]),new T.MeshBasicMaterial({color:0x00ae42,wireframe:true,transparent:true,opacity:.35}));outline.geometry.translate((hi[0]+lo[0])/2,(hi[1]+lo[1])/2,(hi[2]+lo[2])/2);place(outline);scene.add(outline)}
   getBuffer().then(buffer=>{if(!alive)return;
    const worker=new Worker(new URL('./mesh.worker.ts',import.meta.url),{type:'module'});workers.push(worker);
    const fail=(message:string)=>{if(!alive)return;setError(message);if(--pending===0)setLoading('');worker.terminate()};
    worker.onerror=()=>fail('The model preview could not be drawn. Try loading it again.');
    worker.onmessage=({data})=>{if(!alive)return;if(data.error){fail(data.error);return}
     const geometry=new T.BufferGeometry();geometry.setAttribute('position',new T.BufferAttribute(data.positions,3));geometry.setAttribute('normal',new T.BufferAttribute(data.normals,3));
     const mesh=new T.Mesh(geometry,new T.MeshStandardMaterial({color:layer?0x779081:0xc9ad99,roughness:.8,metalness:0,side:T.DoubleSide,transparent:!!layer,opacity:layer?.22:1,depthWrite:!layer,clippingPlanes:layer?[new T.Plane(new T.Vector3(0,0,-1),Number(layer.z)||0)]:[]}));place(mesh);mesh.userData.id=o.id;selectables.push(mesh);scene.add(mesh);
     if(outline){scene.remove(outline);outline.geometry.dispose();(outline.material as T.Material).dispose()}
     if(--pending===0){setLoading('');fit();renderer.domElement.dataset.modelsReady=String(selectables.length)}worker.terminate();
    };worker.postMessage(buffer,[buffer]);
   }).catch((e:any)=>{if(alive){setError(e.message||'The model preview could not be loaded.');if(--pending===0)setLoading('')}})
  }
  const ray=new T.Raycaster(),pointer=new T.Vector2();let down=[0,0];const start=(e:PointerEvent)=>{down=[e.clientX,e.clientY]};const click=(e:PointerEvent)=>{if(Math.hypot(e.clientX-down[0],e.clientY-down[1])>5)return;const r=renderer.domElement.getBoundingClientRect();pointer.set((e.clientX-r.left)/r.width*2-1,-(e.clientY-r.top)/r.height*2+1);ray.setFromCamera(pointer,camera);const hit=ray.intersectObjects(selectables)[0];if(hit)onSelect(hit.object.userData.id)};
  renderer.domElement.addEventListener('pointerdown',start);renderer.domElement.addEventListener('pointerup',click);
 const resize=()=>{const width=Math.max(1,el.clientWidth),height=Math.max(240,el.clientHeight);renderer.setSize(width,height);camera.aspect=width/height;camera.updateProjectionMatrix()};const ro=new ResizeObserver(resize);ro.observe(el);resize();let frame=0;const draw=()=>{if(!alive)return;controls.update();renderer.render(scene,camera);frame=requestAnimationFrame(draw)};draw();
  return()=>{alive=false;workers.forEach(w=>w.terminate());cancelAnimationFrame(frame);ro.disconnect();controls.dispose();scene.traverse((o:any)=>{o.geometry?.dispose();if(o.material)(Array.isArray(o.material)?o.material:[o.material]).forEach((m:any)=>m.dispose())});renderer.dispose();renderer.domElement.remove()};
 },[project.id,project.revision,project.settings.model,project.settings.nativePlate,modelKey,layer,feature,attempt,overview,pathPercent,viewMode,viewToken]);
 return React.createElement('div',{className:'plate-scene'},React.createElement('div',{ref:host}),React.createElement(SceneStatus,{data:{loading:uploading||loading,error,ready:!uploading&&!loading&&!error&&!!project.objects?.length},onAction:()=>setAttempt(a=>a+1)}));
}
