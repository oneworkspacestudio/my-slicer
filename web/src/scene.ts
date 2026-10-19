import React,{useEffect,useRef,useState} from 'react';
import * as T from 'three';
import {OrbitControls} from 'three/addons/controls/OrbitControls.js';
import {STLLoader} from 'three/addons/loaders/STLLoader.js';

const colors=['#94a3b8','#1a73e8','#188038','#f9ab00','#a142f4','#e8710a','#12b5cb','#d93025'];
export function Scene({project,models,layer,feature,request,onSelect}:any){
 const host=useRef<HTMLDivElement>(null),[error,setError]=useState('');
 useEffect(()=>{const el=host.current!;let alive=true;let renderer:T.WebGLRenderer;
  try{renderer=new T.WebGLRenderer({antialias:true,alpha:false});}catch{setError('WebGL is unavailable. Enable hardware acceleration in your browser to inspect the plate. Native slicing and archive downloads remain available.');return}
  setError('');renderer.setPixelRatio(Math.min(devicePixelRatio,2));el.append(renderer.domElement);renderer.domElement.setAttribute('aria-label','Interactive build plate; drag to orbit, scroll to zoom, click an object to select');
  const scene=new T.Scene();scene.background=new T.Color('#eef3fa');const w=project.settings.model==='A1 mini'?180:256;
  const camera=new T.PerspectiveCamera(40,1,0.1,3000);camera.up.set(0,0,1);camera.position.set(w*1.6,-w*1.3,w*1.3);
  const controls=new OrbitControls(camera,renderer.domElement);controls.target.set(w/2,w/2,12);controls.enableDamping=true;
  scene.add(new T.HemisphereLight(0xffffff,0x526178,2));const light=new T.DirectionalLight(0xffffff,3);light.position.set(100,-200,400);scene.add(light);
  const plate=new T.Mesh(new T.PlaneGeometry(w,w),new T.MeshStandardMaterial({color:0xdbe7f5,side:T.DoubleSide,roughness:1}));plate.position.set(w/2,w/2,-.1);scene.add(plate);
  const grid=new T.GridHelper(w,w/10,0x8ba7c6,0xb9cadf);grid.rotation.x=Math.PI/2;grid.position.set(w/2,w/2,0);scene.add(grid);
  const border=new T.LineLoop(new T.BufferGeometry().setFromPoints([new T.Vector3(0,0,.1),new T.Vector3(w,0,.1),new T.Vector3(w,w,.1),new T.Vector3(0,w,.1)]),new T.LineBasicMaterial({color:0x1a73e8}));scene.add(border);
  const selectables:T.Object3D[]=[];
  if(layer){const groups=new Map<number,number[]>();for(const s of layer.segments||[]){if(feature!=='all'&&Number(feature)!==s[6])continue;const arr=groups.get(s[6])||[];arr.push(...s.slice(0,6));groups.set(s[6],arr)}for(const[k,arr]of groups){const g=new T.BufferGeometry();g.setAttribute('position',new T.Float32BufferAttribute(arr,3));scene.add(new T.LineSegments(g,new T.LineBasicMaterial({color:colors[k%colors.length]})))}}
  else for(const o of project.objects||[]){const m=models.find((m:any)=>m.id===o.modelID);if(m?.kind!=='stl')continue;
   request('/api/models/'+m.id+'/mesh').then(async(blob:Blob)=>{const buffer=await blob.arrayBuffer();if(!alive)return;const geometry=new STLLoader().parse(buffer);geometry.computeVertexNormals();const mesh=new T.Mesh(geometry,new T.MeshStandardMaterial({color:0x1a73e8,roughness:.65,metalness:.05,side:T.DoubleSide}));const t=o.transform;mesh.scale.setScalar(t.scale);mesh.rotation.order='ZYX';mesh.rotation.set(t.rx*Math.PI/180,t.ry*Math.PI/180,t.rz*Math.PI/180);mesh.position.set(t.x,t.y,t.z);mesh.userData.id=o.id;selectables.push(mesh);scene.add(mesh)}).catch((e:any)=>{if(alive)setError(e.message)})
  }
  const ray=new T.Raycaster(),pointer=new T.Vector2();let down=[0,0];const start=(e:PointerEvent)=>{down=[e.clientX,e.clientY]};const click=(e:PointerEvent)=>{if(Math.hypot(e.clientX-down[0],e.clientY-down[1])>5)return;const r=renderer.domElement.getBoundingClientRect();pointer.set((e.clientX-r.left)/r.width*2-1,-(e.clientY-r.top)/r.height*2+1);ray.setFromCamera(pointer,camera);const hit=ray.intersectObjects(selectables)[0];if(hit)onSelect(hit.object.userData.id)};
  renderer.domElement.addEventListener('pointerdown',start);renderer.domElement.addEventListener('pointerup',click);
  const resize=()=>{const width=el.clientWidth,height=Math.min(500,Math.max(300,width*.65));renderer.setSize(width,height);camera.aspect=width/height;camera.updateProjectionMatrix()};const ro=new ResizeObserver(resize);ro.observe(el);resize();let frame=0;const draw=()=>{if(!alive)return;controls.update();renderer.render(scene,camera);frame=requestAnimationFrame(draw)};draw();
  return()=>{alive=false;cancelAnimationFrame(frame);ro.disconnect();controls.dispose();scene.traverse((o:any)=>{o.geometry?.dispose();if(o.material)(Array.isArray(o.material)?o.material:[o.material]).forEach((m:any)=>m.dispose())});renderer.dispose();renderer.domElement.remove()};
 },[project.id,project.revision,models.length,layer,feature]);
 return React.createElement('div',{className:'plate-scene'},React.createElement('div',{ref:host}),error&&React.createElement('p',{role:'alert'},error));
}
