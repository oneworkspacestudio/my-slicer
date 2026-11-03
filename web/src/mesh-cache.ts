// Cache raw meshes by bytes as well as entries: complete high-detail surfaces
// are larger than the old sampled previews. Never reduce geometry to save space.
export class MeshBufferCache {
 private entries=new Map<string,ArrayBuffer>();
 private used=0;
 private maxBytes:number;
 private maxEntries:number;
 constructor(maxBytes=64*1024*1024,maxEntries=8){this.maxBytes=maxBytes;this.maxEntries=maxEntries}
 get byteLength(){return this.used}
 get(key:string){
  const value=this.entries.get(key);if(!value)return undefined;
  this.entries.delete(key);this.entries.set(key,value);
  return value.slice(0); // Transferring to the worker must not detach the cache.
 }
 set(key:string,value:ArrayBuffer){
  const old=this.entries.get(key);if(old){this.used-=old.byteLength;this.entries.delete(key)}
  if(value.byteLength>this.maxBytes||this.maxEntries<1)return;
  while(this.entries.size>=this.maxEntries||this.used+value.byteLength>this.maxBytes){
   const oldest=this.entries.keys().next().value!;
   this.used-=this.entries.get(oldest)!.byteLength;this.entries.delete(oldest);
  }
  this.entries.set(key,value.slice(0));this.used+=value.byteLength;
 }
}
